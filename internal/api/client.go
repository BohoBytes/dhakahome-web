package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const defaultStatusFilter = "listed_rental,listed_sale"

// PropertyService defines the interface for property operations
// This allows both real API client and mock service to implement the same interface
type PropertyService interface {
	SearchProperties(q url.Values) (PropertyList, error)
	GetProperty(id string) (Property, error)
	GetRequiredDocuments(assetType string) ([]Document, error)
	GetTopNeighborhoods(limit int, city string) ([]NeighborhoodStat, error)
	SubmitLead(in LeadReq) error
}

type Client struct {
	Base         string
	Token        string
	HC           *http.Client
	tokenURL     string
	clientID     string
	clientSecret string
	scope        string

	mu          sync.Mutex
	cachedToken string
	tokenExpiry time.Time

	// Mock mode
	mockEnabled     bool
	mockAuthEnabled bool

	// Last request metrics (for debugging)
	LastRequestURL      string
	LastRequestDuration time.Duration
	LastResponseStatus  int
	LastResponseError   error
}

func New() *Client {
	// Check if mock mode is enabled
	mockEnabled := strings.ToLower(strings.TrimSpace(os.Getenv("MOCK_ENABLED")))
	useMock := mockEnabled == "true" || mockEnabled == "1" || mockEnabled == "yes"
	mockAuth := useMock
	if v := strings.ToLower(strings.TrimSpace(os.Getenv("MOCK_AUTH_ENABLED"))); v != "" {
		mockAuth = v == "true" || v == "1" || v == "yes"
	}

	base := getenv("API_BASE_URL", "http://localhost:3000/api/v1")
	scope := strings.TrimSpace(getenv("API_TOKEN_SCOPE", "assets.read"))
	if scope == "" {
		scope = "assets.read"
	}

	staticToken := strings.TrimSpace(getenv("API_AUTH_TOKEN", ""))
	clientID := strings.TrimSpace(os.Getenv("API_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("API_CLIENT_SECRET"))
	tokenURL := strings.TrimSpace(getenv("API_AUTH_URL", deriveTokenURL(base)))

	if useMock {
		log.Printf("🎭 API Client: MOCK MODE ENABLED - property searches use mock data; leads will still call Nestlo APIs")
	}

	log.Printf("API Client initialized:")
	log.Printf("  Base URL: %s", base)
	log.Printf("  Static Token: %v (length: %d)", staticToken != "", len(staticToken))
	log.Printf("  OAuth Client ID: %s", clientID)
	log.Printf("  OAuth Token URL: %s", tokenURL)

	return &Client{
		Base:            base,
		Token:           staticToken,
		HC:              &http.Client{Timeout: 10 * time.Second},
		tokenURL:        tokenURL,
		clientID:        clientID,
		clientSecret:    clientSecret,
		scope:           scope,
		mockEnabled:     useMock,
		mockAuthEnabled: mockAuth,
	}
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func deriveTokenURL(base string) string {
	u, err := url.Parse(base)
	if err != nil {
		return "http://localhost:3000/oauth/token"
	}
	u.Path = "/oauth/token"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

type Property struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Address       string   `json:"address"`
	Description   string   `json:"description,omitempty"`
	Price         float64  `json:"price"`
	Currency      string   `json:"currency"`
	Type          string   `json:"type"`
	ListingType   string   `json:"listingType"`
	BuildYear     int      `json:"buildYear,omitempty"`
	Images        []string `json:"images"`
	Badges        []string `json:"badges"`
	Amenities     []string `json:"amenities,omitempty"`
	ListingYear   int      `json:"listingYear,omitempty"`
	ListingDate   string   `json:"listingDate,omitempty"`
	Bedrooms      int      `json:"bedrooms"`
	Bathrooms     int      `json:"bathrooms"`
	Area          int      `json:"area"` // in square feet
	Parking       int      `json:"parking"`
	Gallery       []string `json:"-"`
	HasImages     bool     `json:"-"`
	IsShortlisted bool     `json:"is_shortlisted,omitempty"`
	ShortlistID   string   `json:"shortlist_id,omitempty"`
	ContactPhone  string   `json:"contactPhone,omitempty"`
	ContactEmail  string   `json:"contactEmail,omitempty"`
	Latitude      float64  `json:"latitude,omitempty"`
	Longitude     float64  `json:"longitude,omitempty"`
}

type Document struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	IsRequired bool   `json:"isRequired"`
}

type NeighborhoodStat struct {
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	Count        int    `json:"count"`
}

type PropertyList struct {
	Items []Property `json:"items"`
	Page  int        `json:"page"`
	Pages int        `json:"pages"`
	Total int        `json:"total"`
}

type ShortlistStatus struct {
	AssetID       string `json:"asset_id"`
	ShortlistID   string `json:"shortlist_id,omitempty"`
	IsShortlisted bool   `json:"is_shortlisted"`
}

type assetListResponse struct {
	Data  []map[string]any `json:"data"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

func (c *Client) SearchProperties(q url.Values) (PropertyList, error) {
	params := buildAssetSearchParams(q)

	// If mock mode is enabled, use mock data built from normalized params
	if c.mockEnabled {
		return c.getMockSearchResults(params), nil
	}

	// Track request metrics for debugging
	startTime := time.Now()
	c.LastRequestURL = c.Base + "/assets?" + params.Encode()

	log.Printf("API: Calling GET /assets with params: %s", params.Encode())
	res, err := c.doGet("/assets", params)

	c.LastRequestDuration = time.Since(startTime)
	c.LastResponseError = err

	if err != nil {
		c.LastResponseStatus = 0
		log.Printf("API: Request failed after %dms: %v - using mock data", c.LastRequestDuration.Milliseconds(), err)
		return c.getMockSearchResults(params), nil
	}
	defer res.Body.Close()

	c.LastResponseStatus = res.StatusCode

	if res.StatusCode != http.StatusOK {
		log.Printf("API: Status %d after %dms - using mock data", res.StatusCode, c.LastRequestDuration.Milliseconds())
		return c.getMockSearchResults(params), nil
	}

	var payload assetListResponse
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		log.Printf("API: JSON decode failed: %v - using mock data", err)
		return c.getMockSearchResults(params), nil
	}

	log.Printf("API: Successfully fetched %d properties from backend", len(payload.Data))
	props := make([]Property, 0, len(payload.Data))
	for _, asset := range payload.Data {
		prop := mapAssetToProperty(asset)
		if prop.ID == "" {
			continue
		}
		props = append(props, prop)
	}

	page := payload.Page
	if page <= 0 {
		if p, err := strconv.Atoi(params.Get("page")); err == nil && p > 0 {
			page = p
		} else {
			page = 1
		}
	}
	limit := payload.Limit
	if limit <= 0 {
		if l, err := strconv.Atoi(params.Get("limit")); err == nil && l > 0 {
			limit = l
		} else {
			limit = len(props)
			if limit == 0 {
				limit = 1
			}
		}
	}
	total := payload.Total
	if total <= 0 {
		total = len(props)
	}
	pages := 1
	if total > 0 && limit > 0 {
		pages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return PropertyList{
		Items: props,
		Page:  page,
		Pages: pages,
		Total: total,
	}, nil
}

// CheckShortlist returns whether a property is shortlisted for the authenticated user.
func (c *Client) CheckShortlist(assetID, userToken string) (ShortlistStatus, error) {
	assetID = strings.TrimSpace(assetID)
	if assetID == "" {
		return ShortlistStatus{}, fmt.Errorf("asset id is required")
	}

	if c.mockEnabled {
		return c.mockCheckShortlist(assetID, userToken), nil
	}

	req, err := http.NewRequest(http.MethodGet, c.buildURL(fmt.Sprintf("/shortlists/check/%s", assetID), nil), nil)
	if err != nil {
		return ShortlistStatus{}, err
	}
	if err := c.decorateUserRequest(req, userToken); err != nil {
		return ShortlistStatus{}, err
	}

	res, err := c.HC.Do(req)
	if err != nil {
		return ShortlistStatus{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return ShortlistStatus{}, &APIError{StatusCode: res.StatusCode, Message: "unauthorized"}
	}
	if res.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return ShortlistStatus{}, fmt.Errorf("shortlist check: %s %s", res.Status, strings.TrimSpace(string(detail)))
	}

	var payload ShortlistStatus
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return ShortlistStatus{}, err
	}
	if payload.AssetID == "" {
		payload.AssetID = assetID
	}
	return payload, nil
}

// AddToShortlist adds a property to the default shortlist for the authenticated user.
func (c *Client) AddToShortlist(assetID, userToken string) (ShortlistStatus, error) {
	assetID = strings.TrimSpace(assetID)
	if assetID == "" {
		return ShortlistStatus{}, fmt.Errorf("asset id is required")
	}

	if c.mockEnabled {
		return c.mockAddToShortlist(assetID, userToken), nil
	}

	body, _ := json.Marshal(map[string]string{"asset_id": assetID})
	req, err := http.NewRequest(http.MethodPost, c.buildURL("/shortlists/items", nil), bytes.NewReader(body))
	if err != nil {
		return ShortlistStatus{}, err
	}
	if err := c.decorateUserRequest(req, userToken); err != nil {
		return ShortlistStatus{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HC.Do(req)
	if err != nil {
		return ShortlistStatus{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return ShortlistStatus{}, &APIError{StatusCode: res.StatusCode, Message: "unauthorized"}
	}
	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return ShortlistStatus{}, fmt.Errorf("shortlist add: %s %s", res.Status, strings.TrimSpace(string(detail)))
	}

	var payload ShortlistStatus
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		// Fallback to minimal payload
		payload.AssetID = assetID
		payload.IsShortlisted = true
		return payload, nil
	}
	if payload.AssetID == "" {
		payload.AssetID = assetID
	}
	payload.IsShortlisted = true
	return payload, nil
}

// RemoveFromShortlist removes a property from all shortlists for the authenticated user.
func (c *Client) RemoveFromShortlist(assetID, userToken string) (ShortlistStatus, error) {
	assetID = strings.TrimSpace(assetID)
	if assetID == "" {
		return ShortlistStatus{}, fmt.Errorf("asset id is required")
	}

	if c.mockEnabled {
		return c.mockRemoveFromShortlist(assetID, userToken), nil
	}

	req, err := http.NewRequest(http.MethodDelete, c.buildURL(fmt.Sprintf("/shortlists/items/%s", assetID), nil), nil)
	if err != nil {
		return ShortlistStatus{}, err
	}
	if err := c.decorateUserRequest(req, userToken); err != nil {
		return ShortlistStatus{}, err
	}

	res, err := c.HC.Do(req)
	if err != nil {
		return ShortlistStatus{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return ShortlistStatus{}, &APIError{StatusCode: res.StatusCode, Message: "unauthorized"}
	}
	if res.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return ShortlistStatus{}, fmt.Errorf("shortlist remove: %s %s", res.Status, strings.TrimSpace(string(detail)))
	}

	status := ShortlistStatus{
		AssetID:       assetID,
		IsShortlisted: false,
	}

	var payload ShortlistStatus
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err == nil {
		if payload.AssetID != "" {
			status.AssetID = payload.AssetID
		}
		status.ShortlistID = payload.ShortlistID
		status.IsShortlisted = payload.IsShortlisted
	}

	return status, nil
}

// ListShortlisted fetches the current user's shortlisted properties with pagination support.
func (c *Client) ListShortlisted(userToken string, page, limit int) (PropertyList, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 9
	}

	if c.mockEnabled {
		return c.mockListShortlisted(userToken, page, limit), nil
	}

	shortlistID, err := c.getDefaultShortlistID(userToken)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no shortlist") {
			return PropertyList{
				Items: []Property{},
				Page:  1,
				Pages: 1,
				Total: 0,
			}, nil
		}
		return PropertyList{}, err
	}

	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("limit", strconv.Itoa(limit))

	res, err := c.userRequest(http.MethodGet, fmt.Sprintf("/shortlists/%s", shortlistID), params, nil, userToken)
	if err != nil {
		return PropertyList{}, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return PropertyList{}, &APIError{StatusCode: res.StatusCode, Message: "unauthorized"}
	}
	if res.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return PropertyList{}, fmt.Errorf("shortlist list: %s %s", res.Status, strings.TrimSpace(string(detail)))
	}

	var payload map[string]any
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return PropertyList{}, err
	}

	itemsRaw := pickSlice(payload, "items")
	props := make([]Property, 0, len(itemsRaw))
	for _, row := range itemsRaw {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		asset := pickMap(m, "asset", "Asset")
		prop := mapAssetToProperty(asset)
		if prop.ID == "" {
			prop.ID = firstString(m, "asset_id", "assetId", "id")
		}
		prop.IsShortlisted = true
		prop.ShortlistID = shortlistID
		props = append(props, prop)
	}

	if v, ok := intFrom(payload, "page"); ok && v > 0 {
		page = v
	}
	if v, ok := intFrom(payload, "limit"); ok && v > 0 {
		limit = v
	}

	total := len(props)
	if v, ok := intFrom(payload, "item_count", "total", "TotalItems"); ok && v > 0 {
		total = v
	}
	pages := 1
	if v, ok := intFrom(payload, "pages"); ok && v > 0 {
		pages = v
	} else if limit > 0 && total > 0 {
		pages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return PropertyList{
		Items: props,
		Page:  page,
		Pages: pages,
		Total: total,
	}, nil
}

func (c *Client) getDefaultShortlistID(userToken string) (string, error) {
	if c.mockEnabled {
		return mockShortlists.defaultShortlistID(), nil
	}

	res, err := c.userRequest(http.MethodGet, "/shortlists", nil, nil, userToken)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return "", &APIError{StatusCode: res.StatusCode, Message: "unauthorized"}
	}
	if res.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return "", fmt.Errorf("shortlists: %s %s", res.Status, strings.TrimSpace(string(detail)))
	}

	var rows []map[string]any
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&rows); err != nil {
		return "", err
	}

	var fallback string
	for _, row := range rows {
		id := firstString(row, "id", "ID")
		if id == "" {
			continue
		}
		if fallback == "" {
			fallback = id
		}
		if def, ok := boolFrom(row, "is_default", "isDefault"); ok && def {
			return id, nil
		}
	}

	if fallback != "" {
		return fallback, nil
	}

	return "", fmt.Errorf("no shortlist available for user")
}

func (c *Client) mockCheckShortlist(assetID, userToken string) ShortlistStatus {
	return mockShortlists.status(userToken, assetID)
}

func (c *Client) mockAddToShortlist(assetID, userToken string) ShortlistStatus {
	return mockShortlists.add(userToken, assetID)
}

func (c *Client) mockRemoveFromShortlist(assetID, userToken string) ShortlistStatus {
	return mockShortlists.remove(userToken, assetID)
}

func (c *Client) mockListShortlisted(userToken string, page, limit int) PropertyList {
	return mockShortlists.list(userToken, page, limit)
}

func buildAssetSearchParams(q url.Values) url.Values {
	params := url.Values{}

	page := strings.TrimSpace(q.Get("page"))
	if _, err := strconv.Atoi(page); err != nil || page == "" || page == "0" {
		page = "1"
	}
	params.Set("page", page)

	limit := strings.TrimSpace(q.Get("limit"))
	if _, err := strconv.Atoi(limit); err != nil || limit == "" {
		limit = "9"
	}
	params.Set("limit", limit)

	status := cleanAnyValue(q.Get("status"))
	if status == "" {
		if listingType := normalizeListingType(cleanAnyValue(firstNonEmpty(q.Get("listing_type"), q.Get("listingType")))); listingType != "" {
			status = listingType
		} else {
			status = defaultStatusFilter
		}
	}
	params.Set("status", status)

	if rawQ := cleanAnyValue(q.Get("q")); rawQ != "" {
		params.Set("q", rawQ)
	} else {
		if loc := cleanAnyValue(q.Get("location")); loc != "" {
			params.Set("q", loc)
		} else if area := cleanAnyValue(q.Get("area")); area != "" {
			params.Set("q", area)
		}
	}

	if loc := cleanAnyValue(q.Get("location")); loc != "" {
		params.Set("location", loc)
	}

	if neighborhood := cleanAnyValue(q.Get("neighborhood")); neighborhood != "" {
		params.Set("neighborhood", neighborhood)
	} else if area := cleanAnyValue(q.Get("area")); area != "" {
		params.Set("neighborhood", area)
	}

	if rawTypes := cleanAnyValue(q.Get("types")); rawTypes != "" {
		params.Set("types", rawTypes)
	} else if t := normalizeTypeValue(cleanAnyValue(q.Get("type"))); t != "" {
		params.Set("types", t)
	}

	if priceMax := cleanAnyValue(q.Get("price_max")); priceMax != "" {
		if parsed, ok := parsePriceField(priceMax); ok {
			params.Set("price_max", strconv.FormatFloat(parsed, 'f', -1, 64))
		} else {
			params.Set("price_max", priceMax)
		}
	} else if maxPrice, ok := parsePriceField(q.Get("maxPrice")); ok {
		params.Set("price_max", strconv.FormatFloat(maxPrice, 'f', -1, 64))
	}

	if priceMin := cleanAnyValue(q.Get("price_min")); priceMin != "" {
		if parsed, ok := parsePriceField(priceMin); ok {
			params.Set("price_min", strconv.FormatFloat(parsed, 'f', -1, 64))
		} else {
			params.Set("price_min", priceMin)
		}
	} else if minPrice, ok := parsePriceField(q.Get("minPrice")); ok {
		params.Set("price_min", strconv.FormatFloat(minPrice, 'f', -1, 64))
	}

	for _, key := range []string{
		"city",
		"bedrooms",
		"bathrooms",
		"price_min",
		"parking",
		"serviced",
		"shared_room",
		"area_min",
		"area_max",
		"furnished",
		"subunit_type",
		"exclude_leased",
		"sort_by",
		"order",
	} {
		if val := cleanAnyValue(q.Get(key)); val != "" {
			params.Set(key, val)
		}
	}

	return params
}

func cleanAnyValue(v string) string {
	v = strings.TrimSpace(v)
	if isAnyValue(v) {
		return ""
	}
	return v
}

func isAnyValue(v string) bool {
	return strings.EqualFold(strings.TrimSpace(v), "any")
}

func normalizeListingType(v string) string {
	clean := strings.TrimSpace(strings.ToLower(v))
	switch clean {
	case "", "both":
		return ""
	case "rent", "rental", "listed_rental", "lease", "to-let", "to_let", "tolet":
		return "listed_rental"
	case "sale", "sell", "listed_sale", "for_sale":
		return "listed_sale"
	default:
		return clean
	}
}

func normalizeTypeValue(v string) string {
	clean := strings.TrimSpace(strings.ToLower(v))
	switch clean {
	case "":
		return ""
	case "any":
		return ""
	case "residential":
		return "Residential"
	case "commercial":
		return "Commercial"
	case "land":
		return "Plot"
	default:
		return titleize(clean)
	}
}

func parsePriceField(raw string) (float64, bool) {
	if raw == "" {
		return 0, false
	}
	var b strings.Builder
	for _, r := range raw {
		switch {
		case unicode.IsDigit(r):
			if val, err := strconv.Atoi(string(r)); err == nil {
				b.WriteString(strconv.Itoa(val))
			}
		case r == '.':
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return 0, false
	}
	val, err := strconv.ParseFloat(b.String(), 64)
	if err != nil || val <= 0 {
		return 0, false
	}
	return val, true
}

func (c *Client) GetCities() ([]string, error) {
	if c.mockEnabled {
		return mockCities(), nil
	}

	params := url.Values{}
	params.Set("status", defaultStatusFilter)

	res, err := c.doGet("/assets/cities", params)
	if err != nil {
		log.Printf("API: cities request failed: %v - using mock data", err)
		return mockCities(), nil
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("API: cities status %s - using mock data", res.Status)
		return mockCities(), nil
	}

	var payload any
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		log.Printf("API: cities decode failed: %v - using mock data", err)
		return mockCities(), nil
	}

	cities := parseStringList(payload)
	if len(cities) == 0 {
		log.Printf("API: cities response empty - using mock data")
		return mockCities(), nil
	}

	return cities, nil
}

func (c *Client) GetNeighborhoods(city string) ([]string, error) {
	city = cleanAnyValue(city)
	if city == "" {
		return nil, fmt.Errorf("neighborhoods: city is required")
	}

	if c.mockEnabled {
		return mockNeighborhoods(city), nil
	}

	params := url.Values{}
	params.Set("city", city)
	params.Set("status", defaultStatusFilter)

	res, err := c.doGet("/assets/neighborhoods", params)
	if err != nil {
		log.Printf("API: neighborhoods request failed for city=%s: %v - using mock data", city, err)
		return mockNeighborhoods(city), nil
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("API: neighborhoods status %s for city=%s - using mock data", res.Status, city)
		return mockNeighborhoods(city), nil
	}

	var payload any
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		log.Printf("API: neighborhoods decode failed for city=%s: %v - using mock data", city, err)
		return mockNeighborhoods(city), nil
	}

	areas := parseStringList(payload)
	if len(areas) == 0 {
		log.Printf("API: neighborhoods empty for city=%s - using mock data", city)
		return mockNeighborhoods(city), nil
	}

	return areas, nil
}

func (c *Client) GetTopNeighborhoods(limit int, city string) ([]NeighborhoodStat, error) {
	if limit <= 0 {
		limit = 10
	}
	city = cleanAnyValue(city)

	if c.mockEnabled {
		return mockTopNeighborhoods(limit, city), nil
	}

	params := url.Values{}
	params.Set("limit", strconv.Itoa(limit))
	params.Set("status", defaultStatusFilter)
	if city != "" {
		params.Set("city", city)
	}

	res, err := c.doGet("/assets/neighborhoods/top", params)
	if err != nil {
		log.Printf("API: top neighborhoods request failed: %v - using mock data", err)
		return mockTopNeighborhoods(limit, city), nil
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		log.Printf("API: top neighborhoods status %s - using mock data", res.Status)
		return mockTopNeighborhoods(limit, city), nil
	}

	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	var payload []NeighborhoodStat
	if err := dec.Decode(&payload); err != nil {
		log.Printf("API: top neighborhoods decode failed: %v - using mock data", err)
		return mockTopNeighborhoods(limit, city), nil
	}

	cleaned := make([]NeighborhoodStat, 0, len(payload))
	for _, item := range payload {
		if strings.TrimSpace(item.Neighborhood) == "" {
			continue
		}
		if item.City == "" {
			item.City = city
		}
		cleaned = append(cleaned, item)
	}

	if len(cleaned) == 0 {
		log.Printf("API: top neighborhoods response empty - using mock data")
		return mockTopNeighborhoods(limit, city), nil
	}

	if len(cleaned) > limit {
		cleaned = cleaned[:limit]
	}

	return cleaned, nil
}

func (c *Client) doGet(path string, params url.Values) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, c.buildURL(path, params), nil)
	if err != nil {
		return nil, err
	}
	c.decorateRequest(req)
	return c.HC.Do(req)
}

func (c *Client) decorateRequest(req *http.Request) {
	if header := c.authorizationHeader(); header != "" {
		req.Header.Set("Authorization", header)
	}
	req.Header.Set("Accept", "application/json")
}

func (c *Client) decorateUserRequest(req *http.Request, userToken string) error {
	token := strings.TrimSpace(userToken)
	if token == "" {
		return fmt.Errorf("user token is required for shortlist requests")
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/json")
	return nil
}

func (c *Client) userRequest(method, path string, params url.Values, body io.Reader, userToken string) (*http.Response, error) {
	req, err := http.NewRequest(method, c.buildURL(path, params), body)
	if err != nil {
		return nil, err
	}
	if err := c.decorateUserRequest(req, userToken); err != nil {
		return nil, err
	}
	return c.HC.Do(req)
}

func (c *Client) buildURL(path string, params url.Values) string {
	base := strings.TrimRight(c.Base, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if params != nil && len(params) > 0 {
		return fmt.Sprintf("%s%s?%s", base, path, params.Encode())
	}
	return base + path
}

func (c *Client) authorizationHeader() string {
	// If static token is provided, use it directly (simplest approach)
	if c.Token != "" {
		log.Printf("API: Using static JWT token")
		return fmt.Sprintf("Bearer %s", c.Token)
	}

	// Otherwise, try OAuth client credentials flow
	log.Printf("API: No static token, attempting OAuth with client_id: %s", c.clientID)
	token, err := c.getOAuthToken()
	if err != nil {
		log.Printf("API: OAuth token error: %v", err)
		return ""
	}
	if token == "" {
		log.Printf("API: OAuth returned empty token")
		return ""
	}
	log.Printf("API: Using OAuth token")
	return fmt.Sprintf("Bearer %s", token)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *Client) getOAuthToken() (string, error) {
	if c.clientID == "" || c.clientSecret == "" {
		return "", fmt.Errorf("oauth credentials missing")
	}
	tokenURL := c.tokenURL
	if tokenURL == "" {
		return "", fmt.Errorf("oauth token URL missing")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cachedToken != "" && time.Until(c.tokenExpiry) > time.Minute {
		log.Printf("API: Using cached OAuth token (expires in %v)", time.Until(c.tokenExpiry))
		return c.cachedToken, nil
	}

	// Nestlo backend expects JSON body (not form-encoded)
	requestBody := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	}
	if c.scope != "" {
		requestBody["scope"] = c.scope
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	log.Printf("API: Requesting OAuth token from %s", tokenURL)
	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HC.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		log.Printf("API: OAuth token request failed: %s %s", res.Status, strings.TrimSpace(string(body)))
		return "", fmt.Errorf("oauth token: %s %s", res.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		log.Printf("API: Failed to parse OAuth response: %v", err)
		return "", err
	}
	if payload.AccessToken == "" {
		log.Printf("API: OAuth response contained empty access_token")
		return "", fmt.Errorf("oauth token: empty access_token")
	}

	expiresIn := time.Duration(payload.ExpiresIn) * time.Second
	if expiresIn <= 0 {
		// Nestlo default: 15 minutes (900 seconds)
		expiresIn = 15 * time.Minute
	}
	// Refresh 2 minutes before expiration (or 10% of lifetime, whichever is smaller)
	refreshBefore := 2 * time.Minute
	tenPercent := expiresIn / 10
	if tenPercent < refreshBefore {
		refreshBefore = tenPercent
	}

	c.cachedToken = payload.AccessToken
	c.tokenExpiry = time.Now().Add(expiresIn - refreshBefore)

	log.Printf("API: ✅ OAuth token obtained successfully (expires in %v, will refresh at %v)",
		expiresIn, time.Until(c.tokenExpiry))

	return payload.AccessToken, nil
}

func mapAssetToProperty(raw map[string]any) Property {
	if raw == nil {
		return Property{}
	}

	details := pickMap(raw, "Details", "details")
	location := pickMap(raw, "Location", "location")

	prop := Property{
		ID:          firstString(raw, "ID", "id"),
		Currency:    "৳",
		Type:        titleize(firstString(raw, "Type", "type")),
		ListingType: titleize(firstString(raw, "Status", "status")),
		Title: firstNonEmpty(
			firstString(details, "listing_title", "listingTitle", "title"),
			firstString(raw, "Name", "name"),
		),
	}
	if prop.Title == "" {
		prop.Title = "Property"
	}

	if lat, ok := floatFrom(location, "lat", "latitude", "Lat", "Latitude"); ok {
		prop.Latitude = lat
	}
	if lng, ok := floatFrom(location, "lng", "lon", "longitude", "Longitude", "Long"); ok {
		prop.Longitude = lng
	}
	if prop.Latitude == 0 && prop.Longitude == 0 {
		if lat, lng, ok := coordsFromSlice(pickSlice(location, "coordinates", "coords")); ok {
			prop.Latitude = lat
			prop.Longitude = lng
		}
	}
	if prop.Latitude == 0 && prop.Longitude == 0 {
		if lat, ok := floatFrom(raw, "lat", "latitude"); ok {
			prop.Latitude = lat
		}
		if lng, ok := floatFrom(raw, "lng", "lon", "longitude", "long"); ok {
			prop.Longitude = lng
		}
	}

	prop.Address = firstNonEmpty(
		firstString(raw, "Address", "address"),
		buildAddress(location),
		firstString(location, "raw"),
	)

	prop.Description = firstNonEmpty(
		firstString(details, "description", "listing_description", "listingDescription", "overview", "remarks"),
		firstString(raw, "description", "Description"),
	)

	prop.ContactPhone = firstNonEmpty(
		firstString(details, "contact_phone", "contactPhone", "phone", "owner_phone", "ownerPhone"),
		firstString(raw, "contact_phone", "contactPhone", "phone"),
	)
	prop.ContactEmail = firstNonEmpty(
		firstString(details, "contact_email", "contactEmail", "email"),
		firstString(raw, "contact_email", "contactEmail", "email"),
	)

	prop.Gallery = selectPhotoURLs(pickSlice(raw, "photos", "Photos"))

	if details != nil {
		if v, ok := intFrom(details, "bedrooms"); ok {
			prop.Bedrooms = v
		}
		if v, ok := intFrom(details, "bathrooms"); ok {
			prop.Bathrooms = v
		}
		if v, ok := floatFrom(details, "sizeSqft", "size_sqft"); ok {
			prop.Area = int(v)
		}
		if v, ok := boolFrom(details, "hasParking", "has_parking"); ok && v {
			prop.Parking = 1
		}
		if price := extractPrice(details); price > 0 {
			prop.Price = price
		}
		if prop.ListingType == "" {
			prop.ListingType = titleize(firstString(details, "listing_type", "listingType"))
		}
		if prop.Type == "" {
			prop.Type = titleize(firstString(details, "property_type", "propertyType"))
		}

		if v, ok := intFrom(details, "build_year", "buildYear", "year_built", "yearBuilt"); ok && v > 0 {
			prop.BuildYear = v
		}

		if d := firstString(details, "listing_date", "listingDate", "available_from", "availableFrom", "created_at", "createdAt"); d != "" {
			if parsed, ok := parseDateTime(d); ok {
				prop.ListingYear = parsed.Year()
				prop.ListingDate = parsed.Format("Jan 02, 2006")
			} else {
				prop.ListingDate = d
			}
		}
	}

	if prop.Price == 0 {
		if v, ok := floatFrom(raw, "rent_price", "RentPrice", "monthly_rent"); ok {
			prop.Price = v
		}
	}

	badges := []string{
		prop.Type,
		prop.ListingType,
		titleize(firstString(location, "city")),
		titleize(firstString(location, "neighborhood")),
		titleize(firstString(details, "furnishingStatus", "furnishing_status")),
	}
	prop.Badges = dedupStrings(badges)

	amenities := extractAmenities(details)
	if len(amenities) == 0 {
		amenities = extractAmenities(raw)
	}
	prop.Amenities = dedupStrings(amenities)

	return finalizeProperty(prop)
}

func finalizeProperty(prop Property) Property {
	// Ensure we have a gallery reference to know if real photos exist
	if len(prop.Gallery) == 0 && len(prop.Images) > 0 {
		prop.Gallery = prop.Images
	}
	if len(prop.Gallery) > 0 {
		prop.HasImages = true
	}

	// Keep Images in sync when we do have gallery assets; otherwise allow templates to show their own fallback
	if len(prop.Images) == 0 && len(prop.Gallery) > 0 {
		prop.Images = prop.Gallery
	}

	if prop.Type == "" {
		if t := deriveTypeFromBadges(prop.Badges); t != "" {
			prop.Type = t
		}
	}
	if prop.ListingType == "" {
		if lt := deriveListingTypeFromBadges(prop.Badges); lt != "" {
			prop.ListingType = lt
		}
	}

	if prop.Currency == "" {
		prop.Currency = "৳"
	}

	// Keep amenities populated for display; fallback if none provided
	if len(prop.Amenities) == 0 {
		prop.Amenities = defaultAmenities()
	}

	if prop.Latitude == 0 && prop.Longitude == 0 {
		if lat, lng, ok := fallbackCoordinates(prop); ok {
			prop.Latitude = lat
			prop.Longitude = lng
		}
	}

	return prop
}

func deriveTypeFromBadges(badges []string) string {
	for _, badge := range badges {
		clean := strings.ToLower(strings.TrimSpace(badge))
		switch clean {
		case "residential", "commercial", "land", "plot":
			return titleize(clean)
		}
	}
	return ""
}

func deriveListingTypeFromBadges(badges []string) string {
	for _, badge := range badges {
		clean := strings.ToLower(strings.TrimSpace(badge))
		switch {
		case strings.Contains(clean, "sale"):
			return "For Sale"
		case strings.Contains(clean, "to-let"), strings.Contains(clean, "rent"):
			return "To-let"
		case strings.Contains(clean, "lease"):
			return "Lease"
		}
	}
	return ""
}

var approximateAreaCoords = map[string][2]float64{
	"uttara":      {23.874219, 90.396475},
	"gulshan":     {23.792521, 90.414047},
	"banani":      {23.793478, 90.404137},
	"dhanmondi":   {23.746105, 90.374007},
	"mirpur":      {23.804174, 90.353605},
	"bashundhara": {23.815216, 90.423018},
	"mohammadpur": {23.758726, 90.358072},
	"mohakhali":   {23.780195, 90.400438},
	"baridhara":   {23.810151, 90.422426},
	"nikunja":     {23.826702, 90.422935},
	"badda":       {23.780917, 90.426642},
}

func fallbackCoordinates(prop Property) (float64, float64, bool) {
	haystack := strings.ToLower(strings.Join([]string{
		prop.Address,
		strings.Join(prop.Badges, " "),
		prop.Title,
	}, " "))
	for key, coords := range approximateAreaCoords {
		if strings.Contains(haystack, key) {
			return coords[0], coords[1], true
		}
	}
	return 0, 0, false
}

func parseDateTime(raw string) (time.Time, bool) {
	clean := strings.TrimSpace(raw)
	if clean == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02",
		"2006/01/02",
		"02 Jan 2006",
		"Jan 2, 2006",
		"Jan 02, 2006",
		"02-01-2006",
		"02/01/2006",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, clean); err == nil {
			return t, true
		}
	}
	// try unix seconds
	if secs, err := strconv.ParseInt(clean, 10, 64); err == nil {
		return time.Unix(secs, 0), true
	}
	return time.Time{}, false
}

type mockShortlistStore struct {
	mu        sync.Mutex
	items     map[string]map[string]time.Time
	defaultID string
}

var mockShortlists = newMockShortlistStore()

func newMockShortlistStore() *mockShortlistStore {
	store := &mockShortlistStore{
		items:     make(map[string]map[string]time.Time),
		defaultID: "mock-shortlist-favorites",
	}

	seed := []string{
		"mock-res-uttara-01",
		"mock-res-uttara-03",
		"mock-com-badda-01",
		"mock-com-mohakhali-01",
	}
	now := time.Now()
	store.items["demo"] = make(map[string]time.Time)
	for i, id := range seed {
		store.items["demo"][id] = now.Add(-time.Duration(i) * time.Minute)
	}
	return store
}

func (s *mockShortlistStore) keyFor(token string) string {
	return strings.TrimSpace(token)
}

func (s *mockShortlistStore) defaultShortlistID() string {
	return s.defaultID
}

func (s *mockShortlistStore) ensureUser(token string) map[string]time.Time {
	key := s.keyFor(token)
	if key == "" {
		key = "demo"
	}
	if s.items[key] == nil {
		s.items[key] = make(map[string]time.Time)
	}
	return s.items[key]
}

func (s *mockShortlistStore) status(token, assetID string) ShortlistStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	userItems := s.ensureUser(token)
	_, ok := userItems[assetID]
	return ShortlistStatus{
		AssetID:       assetID,
		ShortlistID:   s.defaultID,
		IsShortlisted: ok,
	}
}

func (s *mockShortlistStore) add(token, assetID string) ShortlistStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	userItems := s.ensureUser(token)
	userItems[assetID] = time.Now()
	return ShortlistStatus{
		AssetID:       assetID,
		ShortlistID:   s.defaultID,
		IsShortlisted: true,
	}
}

func (s *mockShortlistStore) remove(token, assetID string) ShortlistStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	userItems := s.ensureUser(token)
	delete(userItems, assetID)
	return ShortlistStatus{
		AssetID:       assetID,
		ShortlistID:   s.defaultID,
		IsShortlisted: false,
	}
}

func (s *mockShortlistStore) list(token string, page, limit int) PropertyList {
	s.mu.Lock()
	defer s.mu.Unlock()
	userItems := s.ensureUser(token)

	type record struct {
		id   string
		time time.Time
	}
	rows := make([]record, 0, len(userItems))
	for id, added := range userItems {
		rows = append(rows, record{id: id, time: added})
	}

	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].time.After(rows[j].time)
	})

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 9
	}

	total := len(rows)
	pages := int(math.Ceil(float64(total) / float64(limit)))
	if pages == 0 {
		pages = 1
	}
	if page > pages {
		page = pages
	}

	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	items := make([]Property, 0, end-start)
	for _, row := range rows[start:end] {
		if prop, ok := mockPropertyByID(row.id); ok {
			prop.IsShortlisted = true
			prop.ShortlistID = s.defaultID
			items = append(items, finalizeProperty(prop))
		}
	}

	return PropertyList{
		Items: items,
		Page:  page,
		Pages: pages,
		Total: total,
	}
}

func mockRequiredDocuments(assetType string) []Document {
	return []Document{
		{ID: "923dad", Label: "NID", IsRequired: true},
		{ID: "23243fasf", Label: "Employment letter", IsRequired: true},
		{ID: "da5da", Label: "Bank Statement", IsRequired: false},
		{ID: "da67g5da", Label: "Solvency Certificate", IsRequired: false},
	}
}

func mockCities() []string {
	return []string{
		"Dhaka",
		"Chittagong",
		"Sylhet",
		"Khulna",
		"Rajshahi",
	}
}

func mockNeighborhoods(city string) []string {
	switch strings.ToLower(strings.TrimSpace(city)) {
	case "dhaka":
		return []string{"Gulshan", "Banani", "Uttara", "Dhanmondi", "Bashundhara", "Mirpur"}
	case "chittagong":
		return []string{"Agrabad", "Nasirabad", "Pahartali"}
	case "sylhet":
		return []string{"Zinda Bazar", "Amberkhana", "Mirabazar"}
	case "khulna":
		return []string{"Sonadanga", "Khalishpur", "Mujgunni"}
	case "rajshahi":
		return []string{"Uttara", "Boalia", "Rajpara"}
	default:
		return []string{"Central", "North", "South"}
	}
}

func mockTopNeighborhoods(limit int, city string) []NeighborhoodStat {
	if limit <= 0 {
		limit = 10
	}

	city = titleize(firstNonEmpty(cleanAnyValue(city), "Dhaka"))
	areas := mockNeighborhoods(city)

	counts := map[string]int{}
	for _, area := range areas {
		if clean := strings.TrimSpace(area); clean != "" {
			counts[clean] = 0
		}
	}

	for _, prop := range getAllMockProperties() {
		address := strings.ToLower(strings.TrimSpace(prop.Address))
		title := strings.ToLower(strings.TrimSpace(prop.Title))
		for area := range counts {
			areaLower := strings.ToLower(area)
			if strings.Contains(address, areaLower) || strings.Contains(title, areaLower) {
				counts[area]++
			}
		}
	}

	stats := make([]NeighborhoodStat, 0, len(counts))
	for area, count := range counts {
		if count == 0 {
			continue
		}
		stats = append(stats, NeighborhoodStat{
			Neighborhood: area,
			City:         city,
			Count:        count,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Count == stats[j].Count {
			return stats[i].Neighborhood < stats[j].Neighborhood
		}
		return stats[i].Count > stats[j].Count
	})

	if len(stats) > limit {
		stats = stats[:limit]
	}

	return stats
}

func mockPropertyByID(id string) (Property, bool) {
	for _, p := range getAllMockProperties() {
		if strings.EqualFold(p.ID, id) {
			return p, true
		}
	}
	return Property{}, false
}

func defaultAmenities() []string {
	base := []string{
		"Gas Supply",
		"Boundary Wall",
		"Kitchen Cabinet",
		"Power Backup",
		"Parking",
		"Lift",
		"Servant Room",
		"Furnished",
	}
	out := make([]string, len(base))
	copy(out, base)
	return out
}

func extractPrice(details map[string]any) float64 {
	if details == nil {
		return 0
	}
	if pricing := pickMap(details, "pricing", "Pricing"); pricing != nil {
		if v, ok := floatFrom(pricing, "monthly_rent", "rent_price"); ok && v > 0 {
			return v
		}
		if v, ok := floatFrom(pricing, "sale_price", "SalePrice"); ok && v > 0 {
			return v
		}
	}
	if v, ok := floatFrom(details, "sale_price", "SalePrice"); ok && v > 0 {
		return v
	}
	if v, ok := floatFrom(details, "rent_price", "RentPrice"); ok && v > 0 {
		return v
	}
	return 0
}

func firstString(m map[string]any, keys ...string) string {
	if m == nil {
		return ""
	}
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if s := toString(val); s != "" {
				return s
			}
		}
	}
	return ""
}

func firstBool(m map[string]any, keys ...string) bool {
	if m == nil {
		return false
	}
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case bool:
				return v
			case string:
				if b, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
					return b
				}
			case json.Number:
				if num, err := v.Float64(); err == nil {
					return num != 0
				}
			case float64:
				return v != 0
			case int:
				return v != 0
			}
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func toString(v any) string {
	switch value := v.(type) {
	case string:
		return strings.TrimSpace(value)
	case json.Number:
		return value.String()
	case float64:
		if math.IsNaN(value) {
			return ""
		}
		return strconv.FormatFloat(value, 'f', -1, 64)
	case int:
		return strconv.Itoa(value)
	case int64:
		return strconv.FormatInt(value, 10)
	case fmt.Stringer:
		return strings.TrimSpace(value.String())
	default:
		return strings.TrimSpace(fmt.Sprintf("%v", value))
	}
}

func pickMap(m map[string]any, keys ...string) map[string]any {
	if m == nil {
		return nil
	}
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if mm, ok := val.(map[string]any); ok && len(mm) > 0 {
				return mm
			}
			if str, ok := val.(string); ok {
				var out map[string]any
				if err := json.Unmarshal([]byte(str), &out); err == nil && len(out) > 0 {
					return out
				}
			}
		}
	}
	return nil
}

func pickSlice(m map[string]any, keys ...string) []any {
	if m == nil {
		return nil
	}
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch arr := val.(type) {
			case []any:
				return arr
			}
		}
	}
	return nil
}

func intFrom(m map[string]any, keys ...string) (int, bool) {
	if m == nil {
		return 0, false
	}
	if f, ok := floatFrom(m, keys...); ok {
		return int(math.Round(f)), true
	}
	return 0, false
}

func floatFrom(m map[string]any, keys ...string) (float64, bool) {
	if m == nil {
		return 0, false
	}
	for _, key := range keys {
		if val, ok := m[key]; ok {
			if num, ok := parseNumber(val); ok {
				return num, true
			}
		}
	}
	return 0, false
}

func boolFrom(m map[string]any, keys ...string) (bool, bool) {
	if m == nil {
		return false, false
	}
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch value := val.(type) {
			case bool:
				return value, true
			case string:
				clean := strings.TrimSpace(value)
				if clean == "" {
					continue
				}
				if b, err := strconv.ParseBool(clean); err == nil {
					return b, true
				}
				if num, err := strconv.ParseFloat(clean, 64); err == nil {
					return num != 0, true
				}
			case json.Number:
				if num, err := value.Float64(); err == nil {
					return num != 0, true
				}
			case float64:
				return value != 0, true
			case int:
				return value != 0, true
			}
		}
	}
	return false, false
}

func parseNumber(val any) (float64, bool) {
	switch value := val.(type) {
	case json.Number:
		f, err := value.Float64()
		if err != nil {
			return 0, false
		}
		return f, true
	case float64:
		if math.IsNaN(value) {
			return 0, false
		}
		return value, true
	case int:
		return float64(value), true
	case int64:
		return float64(value), true
	case string:
		clean := strings.TrimSpace(value)
		if clean == "" {
			return 0, false
		}
		f, err := strconv.ParseFloat(clean, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	default:
		return 0, false
	}
}

func coordsFromSlice(values []any) (float64, float64, bool) {
	if len(values) < 2 {
		return 0, 0, false
	}
	first, ok1 := parseNumber(values[0])
	second, ok2 := parseNumber(values[1])
	if !ok1 || !ok2 {
		return 0, 0, false
	}
	switch {
	case math.Abs(first) > 60 && math.Abs(second) <= 60:
		// Likely [lng, lat]
		return second, first, true
	case math.Abs(second) > 60 && math.Abs(first) <= 60:
		// Likely [lat, lng]
		return first, second, true
	case math.Abs(second) > math.Abs(first):
		// Dhaka: lon (~90) > lat (~23)
		return first, second, true
	default:
		return second, first, true
	}
}

func buildAddress(location map[string]any) string {
	if location == nil {
		return ""
	}
	parts := []string{}
	for _, key := range []string{"address", "neighborhood", "city"} {
		if v := firstString(location, key); v != "" {
			parts = append(parts, v)
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}
	return firstString(location, "raw")
}

func selectPhotoURLs(items []any) []string {
	if len(items) == 0 {
		return nil
	}
	var cover, others []string
	for _, item := range items {
		photo, ok := item.(map[string]any)
		if !ok || len(photo) == 0 {
			continue
		}
		url := firstString(photo, "FileURL", "file_url", "fileUrl")
		if url == "" {
			continue
		}
		if isCover, ok := boolFrom(photo, "IsCover", "is_cover"); ok && isCover {
			cover = append(cover, url)
		} else {
			others = append(others, url)
		}
	}
	return append(cover, others...)
}

func dedupStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

func stringsFromSlice(values []any) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if s := strings.TrimSpace(toString(v)); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func parseStringList(payload any) []string {
	switch v := payload.(type) {
	case []string:
		return dedupStrings(v)
	case []any:
		return dedupStrings(stringsFromSlice(v))
	case map[string]any:
		if data, ok := v["data"]; ok {
			return parseStringList(data)
		}
	}
	return nil
}

func extractAmenities(m map[string]any) []string {
	if m == nil {
		return nil
	}
	keys := []string{"amenities", "Amenities", "features", "featureList", "features_list", "Features"}
	for _, key := range keys {
		if val, ok := m[key]; ok {
			switch v := val.(type) {
			case []string:
				return v
			case []any:
				return stringsFromSlice(v)
			case string:
				if v == "" {
					continue
				}
				parts := strings.Split(v, ",")
				return parts
			}
		}
	}
	return nil
}

func mapToDocument(m map[string]any) Document {
	return Document{
		ID:         firstString(m, "id", "ID"),
		Label:      firstString(m, "label", "name", "title"),
		IsRequired: firstBool(m, "isRequired", "required", "is_required"),
	}
}

func titleize(input string) string {
	clean := strings.TrimSpace(strings.ReplaceAll(input, "_", " "))
	if clean == "" {
		return ""
	}
	words := strings.Fields(clean)
	for i, word := range words {
		runes := []rune(word)
		for j, r := range runes {
			if j == 0 {
				runes[j] = unicode.ToUpper(r)
			} else {
				runes[j] = unicode.ToLower(r)
			}
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func (c *Client) getMockSearchResults(q url.Values) PropertyList {
	log.Printf("🎭 Mock: Searching properties with params: %s", q.Encode())

	// Parse pagination parameters
	page := parseIntParam(q.Get("page"), 1)
	limit := parseIntParam(q.Get("limit"), 9)

	// Get all mock properties (this will be loaded from mock package)
	mockProperties := getAllMockProperties()

	// Apply filters
	filtered := make([]Property, 0, len(mockProperties))
	for _, prop := range mockProperties {
		if !matchesMockFilters(prop, q) {
			continue
		}
		filtered = append(filtered, prop)
	}

	log.Printf("🎭 Mock: Found %d properties after filtering", len(filtered))

	// Apply pagination
	total := len(filtered)
	pages := int(math.Ceil(float64(total) / float64(limit)))
	if pages == 0 {
		pages = 1
	}

	// Ensure page is within bounds
	if page < 1 {
		page = 1
	}
	if page > pages {
		page = pages
	}

	// Calculate slice bounds
	start := (page - 1) * limit
	end := start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	items := filtered[start:end]
	for i := range items {
		items[i] = finalizeProperty(items[i])
	}

	return PropertyList{
		Items: items,
		Page:  page,
		Pages: pages,
		Total: total,
	}
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func containsAny(slice []string, vals ...string) bool {
	for _, val := range vals {
		valLower := strings.ToLower(val)
		for _, s := range slice {
			sLower := strings.ToLower(s)
			if sLower == valLower || strings.Contains(sLower, valLower) {
				return true
			}
		}
	}
	return false
}

// parseIntParam parses an integer parameter with a default value
func parseIntParam(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || val < 0 {
		return defaultVal
	}
	return val
}

// parseFloatParam parses a float parameter with a default value
func parseFloatParam(s string, defaultVal float64) float64 {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil || val < 0 {
		return defaultVal
	}
	return val
}

// getAllMockProperties returns all mock properties
func getAllMockProperties() []Property {
	// Comprehensive mock dataset
	return []Property{
		// Residential Properties - Uttara Area
		{
			ID:          "mock-res-uttara-01",
			Title:       "Luxury Apartment in Uttara Sec 7",
			Address:     "House 12, Road 7, Sector 7, Uttara, Dhaka",
			Price:       45000,
			Currency:    "৳",
			Images:      []string{"/assets/images/mock-properties/db6726f48a0bae50917980327e8ff5eb40ae871e.png"},
			Badges:      []string{"To-let", "Verified", "Residential", "Fully Furnished"},
			BuildYear:   2020,
			ListingDate: "2024-09-18",
			Description: "Spacious luxury apartment with modern finishes, abundant natural light, and easy access to Uttara's prime conveniences.",
			Bedrooms:    3,
			Bathrooms:   3,
			Area:        1800,
			Parking:     2,
		},
		{
			ID:          "mock-res-uttara-02",
			Title:       "Modern Family Home Uttara Sec 10",
			Address:     "Plot 25, Uttara Sec 10, Dhaka",
			Price:       8500000,
			Currency:    "৳",
			Images:      []string{"/assets/images/mock-properties/8abeccd3fd2f4096a7b4a66a184c5ae36074637a.png"},
			Badges:      []string{"For Sale", "Verified", "Residential"},
			BuildYear:   2018,
			ListingDate: "2024-10-05",
			Description: "Step into this spacious and thoughtfully planned 3-bedroom apartment, ideal for families seeking comfort, convenience, and style. Spanning 1450 square feet, this home features three generously sized bedrooms, each designed to ensure privacy and natural light. The four well-appointed bathrooms, including attached ones, offer added ease for busy households.",
			Bedrooms:    4,
			Bathrooms:   4,
			Area:        2200,
			Parking:     2,
		},
		{
			ID:          "mock-res-uttara-03",
			Title:       "Cozy Studio Apartment Uttara South",
			Address:     "Uttara South, Sector 3, Dhaka",
			Price:       18000,
			Currency:    "৳",
			Images:      []string{"/assets/images/mock-properties/1f002be890c252fab41bc52a14801210d4fa2535.png"},
			Badges:      []string{"To-let", "Verified", "Residential", "Semi-Furnished"},
			BuildYear:   2016,
			ListingDate: "2024-08-12",
			Description: "Efficient studio with smart layout, ideal for single living close to transport and retail.",
			Bedrooms:    1,
			Bathrooms:   1,
			Area:        650,
			Parking:     1,
		},
		{
			ID:          "mock-res-uttara-04",
			Title:       "Spacious 4BR Apartment Uttara Sec 12",
			Address:     "Road 15, Sector 12, Uttara, Dhaka",
			Price:       55000,
			Currency:    "৳",
			Images:      []string{"/assets/images/mock-properties/2f8fe8dfbde9fb83f633da9c0e8bdff775034700.png"},
			Badges:      []string{"To-let", "Verified", "Residential", "Fully Furnished"},
			BuildYear:   2019,
			ListingDate: "2024-09-01",
			Description: "Large four-bedroom with attached baths, ready-to-move furnishings, and cross-ventilation.",
			Bedrooms:    4,
			Bathrooms:   3,
			Area:        2000,
			Parking:     2,
		},
		// Commercial Properties
		{
			ID:          "mock-com-uttara-01",
			Title:       "Premium Office Space Uttara Sec 11",
			Address:     "Building: Crystal Tower, Sector 11, Uttara, Dhaka",
			Price:       120000,
			Currency:    "৳",
			Images:      []string{"/assets/images/mock-properties/d466fbc3c6a3829176f4bf45c88ed96204288a39.png"},
			Badges:      []string{"To-let", "Verified", "Commercial", "Office Space"},
			BuildYear:   2015,
			ListingDate: "2024-07-20",
			Description: "Grade-A office floor with open layout, ample light, and parking allocation.",
			Bedrooms:    0,
			Bathrooms:   2,
			Area:        2500,
			Parking:     3,
		},
		{
			ID:          "mock-com-uttara-02",
			Title:       "Retail Shop Space Uttara Sec 4",
			Address:     "Shop 5, Ground Floor, Uttara Sec 4, Dhaka",
			Price:       3500000,
			Currency:    "৳",
			Images:      []string{"/assets/images/mock-properties/8abeccd3fd2f4096a7b4a66a184c5ae36074637a.png"},
			Badges:      []string{"For Sale", "Verified", "Commercial", "Retail"},
			BuildYear:   2014,
			ListingDate: "2024-06-15",
			Description: "Street-facing retail bay with steady footfall and clear frontage.",
			Bedrooms:    0,
			Bathrooms:   1,
			Area:        800,
			Parking:     0,
		},
		// Gulshan Area
		{
			ID:        "mock-res-gulshan-01",
			Title:     "Elegant Penthouse in Gulshan 2",
			Address:   "Road 78, Gulshan 2, Dhaka",
			Price:     95000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/db6726f48a0bae50917980327e8ff5eb40ae871e.png"},
			Badges:    []string{"To-let", "Verified", "Residential", "Fully Furnished", "Luxury"},
			Bedrooms:  5,
			Bathrooms: 5,
			Area:      3500,
			Parking:   3,
		},
		{
			ID:        "mock-res-gulshan-02",
			Title:     "Modern 3BR Flat Gulshan 1",
			Address:   "House 45, Road 12, Gulshan 1, Dhaka",
			Price:     65000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/2f8fe8dfbde9fb83f633da9c0e8bdff775034700.png"},
			Badges:    []string{"To-let", "Verified", "Residential", "Semi-Furnished"},
			Bedrooms:  3,
			Bathrooms: 2,
			Area:      1600,
			Parking:   2,
		},
		{
			ID:        "mock-com-gulshan-01",
			Title:     "Corporate Office Gulshan Avenue",
			Address:   "Gulshan Avenue, Gulshan 1, Dhaka",
			Price:     250000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/d466fbc3c6a3829176f4bf45c88ed96204288a39.png"},
			Badges:    []string{"To-let", "Verified", "Commercial", "Office Space", "Premium"},
			Bedrooms:  0,
			Bathrooms: 4,
			Area:      4000,
			Parking:   5,
		},
		// Banani Area
		{
			ID:        "mock-res-banani-01",
			Title:     "Luxurious Apartment Banani DOHS",
			Address:   "Block C, Road 5, Banani DOHS, Dhaka",
			Price:     75000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/1f002be890c252fab41bc52a14801210d4fa2535.png"},
			Badges:    []string{"To-let", "Verified", "Residential", "Fully Furnished"},
			Bedrooms:  4,
			Bathrooms: 4,
			Area:      2400,
			Parking:   2,
		},
		{
			ID:        "mock-res-banani-02",
			Title:     "2 Bedroom Apartment in Banani",
			Address:   "Road 11, Banani, Dhaka",
			Price:     35000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/8abeccd3fd2f4096a7b4a66a184c5ae36074637a.png"},
			Badges:    []string{"To-let", "Verified", "Residential"},
			Bedrooms:  2,
			Bathrooms: 2,
			Area:      1100,
			Parking:   1,
		},
		// Dhanmondi Area
		{
			ID:        "mock-res-dhanmondi-01",
			Title:     "Beautiful Lake View Flat Dhanmondi",
			Address:   "Road 8/A, Dhanmondi, Dhaka",
			Price:     55000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/db6726f48a0bae50917980327e8ff5eb40ae871e.png"},
			Badges:    []string{"To-let", "Verified", "Residential", "Lake View"},
			Bedrooms:  3,
			Bathrooms: 3,
			Area:      1900,
			Parking:   2,
		},
		{
			ID:        "mock-res-dhanmondi-02",
			Title:     "Spacious Family Apartment Dhanmondi 15",
			Address:   "Road 15, Dhanmondi, Dhaka",
			Price:     12000000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/2f8fe8dfbde9fb83f633da9c0e8bdff775034700.png"},
			Badges:    []string{"For Sale", "Verified", "Residential"},
			Bedrooms:  4,
			Bathrooms: 3,
			Area:      2100,
			Parking:   2,
		},
		{
			ID:        "mock-com-dhanmondi-01",
			Title:     "Commercial Space Satmasjid Road",
			Address:   "Satmasjid Road, Dhanmondi, Dhaka",
			Price:     85000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/d466fbc3c6a3829176f4bf45c88ed96204288a39.png"},
			Badges:    []string{"To-let", "Verified", "Commercial", "Retail"},
			Bedrooms:  0,
			Bathrooms: 2,
			Area:      1500,
			Parking:   1,
		},
		// Mirpur Area
		{
			ID:        "mock-res-mirpur-01",
			Title:     "Affordable Family Flat Mirpur 10",
			Address:   "Road 12, Mirpur 10, Dhaka",
			Price:     22000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/1f002be890c252fab41bc52a14801210d4fa2535.png"},
			Badges:    []string{"To-let", "Verified", "Residential"},
			Bedrooms:  3,
			Bathrooms: 2,
			Area:      1200,
			Parking:   1,
		},
		{
			ID:        "mock-res-mirpur-02",
			Title:     "Budget Friendly 2BR Mirpur 11",
			Address:   "Section 11, Mirpur, Dhaka",
			Price:     16000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/8abeccd3fd2f4096a7b4a66a184c5ae36074637a.png"},
			Badges:    []string{"To-let", "Verified", "Residential"},
			Bedrooms:  2,
			Bathrooms: 1,
			Area:      900,
			Parking:   0,
		},
		// Bashundhara Area
		{
			ID:        "mock-res-bashundhara-01",
			Title:     "Modern Apartment Bashundhara R/A",
			Address:   "Block G, Road 5, Bashundhara R/A, Dhaka",
			Price:     48000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/db6726f48a0bae50917980327e8ff5eb40ae871e.png"},
			Badges:    []string{"To-let", "Verified", "Residential", "Semi-Furnished"},
			Bedrooms:  3,
			Bathrooms: 3,
			Area:      1700,
			Parking:   2,
		},
		{
			ID:        "mock-res-bashundhara-02",
			Title:     "Luxury Villa Bashundhara",
			Address:   "Block D, Bashundhara R/A, Dhaka",
			Price:     25000000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/2f8fe8dfbde9fb83f633da9c0e8bdff775034700.png"},
			Badges:    []string{"For Sale", "Verified", "Residential", "Luxury"},
			Bedrooms:  6,
			Bathrooms: 6,
			Area:      4500,
			Parking:   4,
		},
		// Mohammadpur Area
		{
			ID:        "mock-res-mohammadpur-01",
			Title:     "Comfortable Flat Mohammadpur",
			Address:   "Nobodoy Housing, Mohammadpur, Dhaka",
			Price:     20000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/1f002be890c252fab41bc52a14801210d4fa2535.png"},
			Badges:    []string{"To-let", "Verified", "Residential"},
			Bedrooms:  2,
			Bathrooms: 2,
			Area:      1000,
			Parking:   1,
		},
		// Hostel/Shared Properties
		{
			ID:        "mock-hostel-01",
			Title:     "Student Hostel Near NSU Bashundhara",
			Address:   "Near NSU, Bashundhara, Dhaka",
			Price:     8000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/8abeccd3fd2f4096a7b4a66a184c5ae36074637a.png"},
			Badges:    []string{"To-let", "Verified", "Hostel", "Shared"},
			Bedrooms:  1,
			Bathrooms: 1,
			Area:      250,
			Parking:   0,
		},
		{
			ID:        "mock-hostel-02",
			Title:     "Working Professional Hostel Uttara",
			Address:   "Sector 9, Uttara, Dhaka",
			Price:     12000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/1f002be890c252fab41bc52a14801210d4fa2535.png"},
			Badges:    []string{"To-let", "Verified", "Hostel", "Furnished"},
			Bedrooms:  1,
			Bathrooms: 1,
			Area:      350,
			Parking:   0,
		},
		// Short Term Rentals
		{
			ID:        "mock-str-01",
			Title:     "Service Apartment Banani (Daily/Monthly)",
			Address:   "Road 17, Banani, Dhaka",
			Price:     3500,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/db6726f48a0bae50917980327e8ff5eb40ae871e.png"},
			Badges:    []string{"To-let", "Verified", "Short Term Rental", "Fully Furnished"},
			Bedrooms:  1,
			Bathrooms: 1,
			Area:      550,
			Parking:   0,
		},
		{
			ID:        "mock-str-02",
			Title:     "Serviced Studio Gulshan 2",
			Address:   "Road 86, Gulshan 2, Dhaka",
			Price:     4500,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/2f8fe8dfbde9fb83f633da9c0e8bdff775034700.png"},
			Badges:    []string{"To-let", "Verified", "Short Term Rental", "Luxury"},
			Bedrooms:  1,
			Bathrooms: 1,
			Area:      600,
			Parking:   1,
		},
	}
}

// matchesMockFilters checks if a property matches the given filter criteria
func matchesMockFilters(prop Property, q url.Values) bool {
	// Text search (q parameter) - searches in title, address, and badges
	if searchQuery := cleanAnyValue(q.Get("q")); searchQuery != "" {
		searchLower := strings.ToLower(searchQuery)
		if !contains(prop.Title, searchLower) &&
			!contains(prop.Address, searchLower) &&
			!containsAny(prop.Badges, searchLower) {
			return false
		}
	}

	// Location filter
	if location := cleanAnyValue(q.Get("location")); location != "" {
		if !contains(prop.Address, location) && !contains(prop.Title, location) {
			return false
		}
	}

	// Area/Neighborhood filter
	if area := cleanAnyValue(q.Get("area")); area != "" {
		if !contains(prop.Address, area) && !contains(prop.Title, area) {
			return false
		}
	}
	if neighborhood := cleanAnyValue(q.Get("neighborhood")); neighborhood != "" {
		if !contains(prop.Address, neighborhood) && !contains(prop.Title, neighborhood) {
			return false
		}
	}

	// Property type filter
	if types := cleanAnyValue(q.Get("types")); types != "" {
		typeList := strings.Split(types, ",")
		found := false
		for _, t := range typeList {
			if containsAny(prop.Badges, strings.TrimSpace(t)) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if propertyType := cleanAnyValue(q.Get("type")); propertyType != "" {
		if !containsAny(prop.Badges, propertyType) {
			return false
		}
	}

	// Status filter
	if status := cleanAnyValue(q.Get("status")); status != "" {
		statusList := strings.Split(status, ",")
		found := false
		for _, s := range statusList {
			normalized := normalizeMockStatus(strings.TrimSpace(s))
			if containsAny(prop.Badges, normalized) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Price filters
	if minPrice := parseFloatParam(q.Get("price_min"), 0); minPrice > 0 {
		if prop.Price < minPrice {
			return false
		}
	}
	if maxPrice := parseFloatParam(q.Get("price_max"), 0); maxPrice > 0 {
		if prop.Price > maxPrice {
			return false
		}
	}

	// Parking filter
	if parking := parseIntParam(q.Get("parking"), -1); parking >= 0 {
		if parking >= 10 {
			if prop.Parking < parking {
				return false
			}
		} else if parking >= 3 {
			if prop.Parking < parking {
				return false
			}
		} else if prop.Parking != parking {
			return false
		}
	}

	// Area/Sqft filter
	if areaMin := parseFloatParam(q.Get("area_min"), 0); areaMin > 0 && prop.Area > 0 {
		if float64(prop.Area) < areaMin {
			return false
		}
	}
	if areaMax := parseFloatParam(q.Get("area_max"), 0); areaMax > 0 && prop.Area > 0 {
		if float64(prop.Area) > areaMax {
			return false
		}
	}

	// Bedrooms filter
	if bedrooms := parseIntParam(q.Get("bedrooms"), 0); bedrooms > 0 {
		if prop.Bedrooms < bedrooms {
			return false
		}
	}

	// Bathrooms filter
	if bathrooms := parseIntParam(q.Get("bathrooms"), 0); bathrooms > 0 {
		if prop.Bathrooms < bathrooms {
			return false
		}
	}

	// Furnished filter
	if furnished := cleanAnyValue(q.Get("furnished")); furnished != "" {
		furnishedLower := strings.ToLower(furnished)
		if furnishedLower == "yes" || furnishedLower == "true" || furnishedLower == "1" {
			if !containsAny(prop.Badges, "Furnished") {
				return false
			}
		} else if furnishedLower == "no" || furnishedLower == "false" || furnishedLower == "0" {
			if containsAny(prop.Badges, "Furnished") {
				return false
			}
		}
	}

	// Serviced filter (basic heuristic on badges/title)
	if serviced := cleanAnyValue(q.Get("serviced")); serviced != "" {
		isServiced := containsAny(prop.Badges, "Serviced") || contains(prop.Title, "serviced")
		if serviced == "true" || serviced == "yes" || serviced == "1" {
			if !isServiced {
				return false
			}
		} else if serviced == "false" || serviced == "no" || serviced == "0" {
			if isServiced {
				return false
			}
		}
	}

	// Shared room filter (hostel heuristic)
	if shared := cleanAnyValue(firstNonEmpty(q.Get("shared_room"), q.Get("sharedRoom"))); shared != "" {
		isShared := contains(prop.Title, "shared") || containsAny(prop.Badges, "Shared", "Shared Room")
		if shared == "true" || shared == "yes" || shared == "1" {
			if !isShared {
				return false
			}
		} else if shared == "false" || shared == "no" || shared == "0" {
			if isShared {
				return false
			}
		}
	}

	return true
}

// normalizeMockStatus normalizes backend status values to display values
func normalizeMockStatus(status string) string {
	statusLower := strings.ToLower(strings.TrimSpace(status))
	switch statusLower {
	case "listed_rental":
		return "To-let"
	case "listed_sale":
		return "For Sale"
	case "ready_for_listing", "active", "available":
		return "To-let"
	case "for_sale", "sale":
		return "For Sale"
	case "leased", "rented":
		return "Leased"
	default:
		return titleize(statusLower)
	}
}

func (c *Client) GetProperty(id string) (Property, error) {
	var out Property
	if id == "" {
		return out, fmt.Errorf("property id required")
	}

	// If mock mode is enabled, search in mock data
	if c.mockEnabled {
		if prop, ok := mockPropertyByID(id); ok {
			log.Printf("🎭 Mock: Found property with ID: %s", id)
			return finalizeProperty(prop), nil
		}
		return out, fmt.Errorf("property not found: %s", id)
	}

	res, err := c.doGet(fmt.Sprintf("/assets/%s", id), nil)
	if err != nil {
		if prop, ok := mockPropertyByID(id); ok {
			log.Printf("API: falling back to mock property for id=%s after error: %v", id, err)
			return finalizeProperty(prop), nil
		}
		return out, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		if prop, ok := mockPropertyByID(id); ok {
			log.Printf("API: status %s for id=%s; using mock data", res.Status, id)
			return finalizeProperty(prop), nil
		}
		return out, fmt.Errorf("api: %s", res.Status)
	}
	var payload map[string]any
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		if prop, ok := mockPropertyByID(id); ok {
			log.Printf("API: decode failed for id=%s: %v; using mock data", id, err)
			return finalizeProperty(prop), nil
		}
		return out, err
	}
	prop := mapAssetToProperty(payload)
	if prop.ID == "" {
		prop.ID = id
	}
	return prop, nil
}

func (c *Client) GetRequiredDocuments(assetType string) ([]Document, error) {
	assetType = strings.TrimSpace(strings.ToLower(assetType))
	if assetType == "" {
		assetType = "default"
	}

	// Mock path
	if c.mockEnabled {
		return mockRequiredDocuments(assetType), nil
	}

	endpoint := fmt.Sprintf("/config/asset/%s/documents", assetType)
	res, err := c.doGet(endpoint, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("documents: %s", res.Status)
	}

	var payload any
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return nil, err
	}

	// Payload can be an array or { data: [] }
	var rows []any
	switch v := payload.(type) {
	case []any:
		rows = v
	case map[string]any:
		if data, ok := v["data"]; ok {
			if arr, ok := data.([]any); ok {
				rows = arr
			}
		}
	}

	docs := make([]Document, 0, len(rows))
	for _, row := range rows {
		if m, ok := row.(map[string]any); ok {
			doc := mapToDocument(m)
			if doc.Label == "" {
				continue
			}
			if doc.ID == "" {
				doc.ID = doc.Label
			}
			docs = append(docs, doc)
		}
	}
	return docs, nil
}

type LeadReq struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	PropertyID   string `json:"propertyId"`
	Message      string `json:"message,omitempty"`
	ContactEmail string `json:"contactEmail,omitempty"`
	UTMSource    string `json:"utmSource,omitempty"`
	UTMCampaign  string `json:"utmCampaign,omitempty"`
	CaptchaToken string `json:"captchaToken,omitempty"`
}

type NestloLeadClientInfo struct {
	Name                   string `json:"name"`
	Email                  string `json:"email,omitempty"`
	Phone                  string `json:"phone,omitempty"`
	PreferredContactMethod string `json:"preferred_contact_method,omitempty"`
}

type NestloLeadRequirements struct {
	PropertyTypes []string `json:"property_types,omitempty"`
	Locations     []string `json:"locations,omitempty"`
	BudgetMin     float64  `json:"budget_min,omitempty"`
	BudgetMax     float64  `json:"budget_max,omitempty"`
	Bedrooms      int      `json:"bedrooms,omitempty"`
	Bathrooms     int      `json:"bathrooms,omitempty"`
	Amenities     []string `json:"amenities,omitempty"`
	MoveInDate    string   `json:"move_in_date,omitempty"`
}

type NestloLeadPayload struct {
	LeadType     string                  `json:"lead_type"`
	Source       string                  `json:"source"`
	ClientInfo   NestloLeadClientInfo    `json:"client_info"`
	Requirements *NestloLeadRequirements `json:"requirements,omitempty"`
	Notes        string                  `json:"notes,omitempty"`
	AssetID      string                  `json:"asset_id,omitempty"`
}

func (c *Client) SubmitLead(in LeadReq) error {
	endp := c.buildURL("/leads", nil)
	b, _ := json.Marshal(in)
	req, err := http.NewRequest(http.MethodPost, endp, bytes.NewReader(b))
	if err != nil {
		return err
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.HC.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("lead: %s", res.Status)
	}
	return nil
}

func (c *Client) CreateNestloLead(in NestloLeadPayload) error {
	if strings.TrimSpace(in.LeadType) == "" {
		in.LeadType = "tenant"
	}
	if strings.TrimSpace(in.Source) == "" {
		in.Source = "web"
	}

	endp := c.buildURL("/admin/leads", nil)
	body, _ := json.Marshal(in)
	req, err := http.NewRequest(http.MethodPost, endp, bytes.NewReader(body))
	if err != nil {
		return err
	}
	c.decorateRequest(req)
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	res, err := c.HC.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return fmt.Errorf("nestlo lead: %s %s", res.Status, strings.TrimSpace(string(detail)))
	}

	log.Printf("Nestlo lead created for asset %s in %dms", in.AssetID, time.Since(start).Milliseconds())
	return nil
}
