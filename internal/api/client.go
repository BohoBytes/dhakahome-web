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
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// PropertyService defines the interface for property operations
// This allows both real API client and mock service to implement the same interface
type PropertyService interface {
	SearchProperties(q url.Values) (PropertyList, error)
	GetProperty(id string) (Property, error)
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
	mockEnabled bool

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

	if useMock {
		log.Printf("🎭 API Client: MOCK MODE ENABLED - All API calls will use mock data")
		return &Client{
			mockEnabled: true,
			HC:          &http.Client{Timeout: 10 * time.Second},
		}
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

	log.Printf("API Client initialized:")
	log.Printf("  Base URL: %s", base)
	log.Printf("  Static Token: %v (length: %d)", staticToken != "", len(staticToken))
	log.Printf("  OAuth Client ID: %s", clientID)
	log.Printf("  OAuth Token URL: %s", tokenURL)

	return &Client{
		Base:         base,
		Token:        staticToken,
		HC:           &http.Client{Timeout: 10 * time.Second},
		tokenURL:     tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		scope:        scope,
		mockEnabled:  false,
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
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Address     string   `json:"address"`
	Price       float64  `json:"price"`
	Currency    string   `json:"currency"`
	Type        string   `json:"type"`
	ListingType string   `json:"listingType"`
	Images      []string `json:"images"`
	Badges      []string `json:"badges"`
	Bedrooms    int      `json:"bedrooms"`
	Bathrooms   int      `json:"bathrooms"`
	Area        int      `json:"area"` // in square feet
	Parking     int      `json:"parking"`
	Gallery     []string `json:"-"`
	HasImages   bool     `json:"-"`
}

type PropertyList struct {
	Items []Property `json:"items"`
	Page  int        `json:"page"`
	Pages int        `json:"pages"`
	Total int        `json:"total"`
}

type assetListResponse struct {
	Data  []map[string]any `json:"data"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

func (c *Client) SearchProperties(q url.Values) (PropertyList, error) {
	// If mock mode is enabled, use mock data
	if c.mockEnabled {
		return c.getMockSearchResults(q), nil
	}

	params := buildAssetSearchParams(q)

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
		return c.getMockSearchResults(q), nil
	}
	defer res.Body.Close()

	c.LastResponseStatus = res.StatusCode

	if res.StatusCode != http.StatusOK {
		log.Printf("API: Status %d after %dms - using mock data", res.StatusCode, c.LastRequestDuration.Milliseconds())
		return c.getMockSearchResults(q), nil
	}

	var payload assetListResponse
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		log.Printf("API: JSON decode failed: %v - using mock data", err)
		return c.getMockSearchResults(q), nil
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

	status := strings.TrimSpace(q.Get("status"))
	if status == "" {
		// Nestlo backend statuses: ready_for_listing, active, leased, etc.
		status = "ready_for_listing,active"
	}
	params.Set("status", status)

	if rawQ := strings.TrimSpace(q.Get("q")); rawQ != "" {
		params.Set("q", rawQ)
	} else {
		if loc := strings.TrimSpace(q.Get("location")); loc != "" {
			params.Set("q", loc)
		} else if area := strings.TrimSpace(q.Get("area")); area != "" {
			params.Set("q", area)
		}
	}

	if loc := strings.TrimSpace(q.Get("location")); loc != "" {
		params.Set("location", loc)
	}

	if neighborhood := strings.TrimSpace(q.Get("neighborhood")); neighborhood != "" {
		params.Set("neighborhood", neighborhood)
	} else if area := strings.TrimSpace(q.Get("area")); area != "" {
		params.Set("neighborhood", area)
	}

	if rawTypes := strings.TrimSpace(q.Get("types")); rawTypes != "" {
		params.Set("types", rawTypes)
	} else if t := normalizeTypeValue(q.Get("type")); t != "" {
		params.Set("types", t)
	}

	if priceMax := strings.TrimSpace(q.Get("price_max")); priceMax != "" {
		params.Set("price_max", priceMax)
	} else if maxPrice, ok := parsePriceField(q.Get("maxPrice")); ok {
		params.Set("price_max", strconv.FormatFloat(maxPrice, 'f', -1, 64))
	}

	for _, key := range []string{
		"city",
		"bedrooms",
		"bathrooms",
		"price_min",
		"furnished",
		"subunit_type",
		"exclude_leased",
		"sort_by",
		"order",
	} {
		if val := strings.TrimSpace(q.Get(key)); val != "" {
			params.Set(key, val)
		}
	}

	return params
}

func normalizeTypeValue(v string) string {
	clean := strings.TrimSpace(strings.ToLower(v))
	switch clean {
	case "":
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
		if unicode.IsDigit(r) || r == '.' {
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

	prop.Address = firstNonEmpty(
		firstString(raw, "Address", "address"),
		buildAddress(location),
		firstString(location, "raw"),
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

	// Provide a placeholder only for display; keep HasImages for actual photo presence
	if len(prop.Images) == 0 {
		if len(prop.Gallery) > 0 {
			prop.Images = prop.Gallery
		} else {
			prop.Images = []string{"/assets/images/placeholders/property-placeholder.svg"}
		}
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

func containsAny(slice []string, val string) bool {
	valLower := strings.ToLower(val)
	for _, s := range slice {
		if strings.ToLower(s) == valLower || strings.Contains(strings.ToLower(s), valLower) {
			return true
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
			ID:        "mock-res-uttara-01",
			Title:     "Luxury Apartment in Uttara Sec 7",
			Address:   "House 12, Road 7, Sector 7, Uttara, Dhaka",
			Price:     45000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/db6726f48a0bae50917980327e8ff5eb40ae871e.png"},
			Badges:    []string{"To-let", "Verified", "Residential", "Fully Furnished"},
			Bedrooms:  3,
			Bathrooms: 3,
			Area:      1800,
			Parking:   2,
		},
		{
			ID:        "mock-res-uttara-02",
			Title:     "Modern Family Home Uttara Sec 10",
			Address:   "Plot 25, Uttara Sec 10, Dhaka",
			Price:     8500000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/8abeccd3fd2f4096a7b4a66a184c5ae36074637a.png"},
			Badges:    []string{"For Sale", "Verified", "Residential"},
			Bedrooms:  4,
			Bathrooms: 4,
			Area:      2200,
			Parking:   2,
		},
		{
			ID:        "mock-res-uttara-03",
			Title:     "Cozy Studio Apartment Uttara South",
			Address:   "Uttara South, Sector 3, Dhaka",
			Price:     18000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/1f002be890c252fab41bc52a14801210d4fa2535.png"},
			Badges:    []string{"To-let", "Verified", "Residential", "Semi-Furnished"},
			Bedrooms:  1,
			Bathrooms: 1,
			Area:      650,
			Parking:   1,
		},
		{
			ID:        "mock-res-uttara-04",
			Title:     "Spacious 4BR Apartment Uttara Sec 12",
			Address:   "Road 15, Sector 12, Uttara, Dhaka",
			Price:     55000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/2f8fe8dfbde9fb83f633da9c0e8bdff775034700.png"},
			Badges:    []string{"To-let", "Verified", "Residential", "Fully Furnished"},
			Bedrooms:  4,
			Bathrooms: 3,
			Area:      2000,
			Parking:   2,
		},
		// Commercial Properties
		{
			ID:        "mock-com-uttara-01",
			Title:     "Premium Office Space Uttara Sec 11",
			Address:   "Building: Crystal Tower, Sector 11, Uttara, Dhaka",
			Price:     120000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/d466fbc3c6a3829176f4bf45c88ed96204288a39.png"},
			Badges:    []string{"To-let", "Verified", "Commercial", "Office Space"},
			Bedrooms:  0,
			Bathrooms: 2,
			Area:      2500,
			Parking:   3,
		},
		{
			ID:        "mock-com-uttara-02",
			Title:     "Retail Shop Space Uttara Sec 4",
			Address:   "Shop 5, Ground Floor, Uttara Sec 4, Dhaka",
			Price:     3500000,
			Currency:  "৳",
			Images:    []string{"/assets/images/mock-properties/8abeccd3fd2f4096a7b4a66a184c5ae36074637a.png"},
			Badges:    []string{"For Sale", "Verified", "Commercial", "Retail"},
			Bedrooms:  0,
			Bathrooms: 1,
			Area:      800,
			Parking:   0,
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
	if searchQuery := strings.TrimSpace(q.Get("q")); searchQuery != "" {
		searchLower := strings.ToLower(searchQuery)
		if !contains(prop.Title, searchLower) &&
			!contains(prop.Address, searchLower) &&
			!containsAny(prop.Badges, searchLower) {
			return false
		}
	}

	// Location filter
	if location := strings.TrimSpace(q.Get("location")); location != "" {
		if !contains(prop.Address, location) && !contains(prop.Title, location) {
			return false
		}
	}

	// Area/Neighborhood filter
	if area := strings.TrimSpace(q.Get("area")); area != "" {
		if !contains(prop.Address, area) && !contains(prop.Title, area) {
			return false
		}
	}
	if neighborhood := strings.TrimSpace(q.Get("neighborhood")); neighborhood != "" {
		if !contains(prop.Address, neighborhood) && !contains(prop.Title, neighborhood) {
			return false
		}
	}

	// Property type filter
	if types := strings.TrimSpace(q.Get("types")); types != "" {
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
	if propertyType := strings.TrimSpace(q.Get("type")); propertyType != "" {
		if !containsAny(prop.Badges, propertyType) {
			return false
		}
	}

	// Status filter
	if status := strings.TrimSpace(q.Get("status")); status != "" {
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
	if furnished := strings.TrimSpace(q.Get("furnished")); furnished != "" {
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

	return true
}

// normalizeMockStatus normalizes backend status values to display values
func normalizeMockStatus(status string) string {
	statusLower := strings.ToLower(strings.TrimSpace(status))
	switch statusLower {
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
		mockList := c.getMockSearchResults(url.Values{})
		for _, prop := range mockList.Items {
			if prop.ID == id {
				log.Printf("🎭 Mock: Found property with ID: %s", id)
				return prop, nil
			}
		}
		return out, fmt.Errorf("property not found: %s", id)
	}

	res, err := c.doGet(fmt.Sprintf("/assets/%s", id), nil)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return out, fmt.Errorf("api: %s", res.Status)
	}
	var payload map[string]any
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return out, err
	}
	prop := mapAssetToProperty(payload)
	if prop.ID == "" {
		prop.ID = id
	}
	return prop, nil
}

type LeadReq struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	PropertyID   string `json:"propertyId"`
	UTMSource    string `json:"utmSource,omitempty"`
	UTMCampaign  string `json:"utmCampaign,omitempty"`
	CaptchaToken string `json:"captchaToken,omitempty"`
}

func (c *Client) SubmitLead(in LeadReq) error {
	// If mock mode is enabled, just log and return success
	if c.mockEnabled {
		log.Printf("🎭 Mock: Lead submitted - Name: %s, Email: %s, Phone: %s, PropertyID: %s",
			in.Name, in.Email, in.Phone, in.PropertyID)
		return nil
	}

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
