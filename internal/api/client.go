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
}

func New() *Client {
	base := getenv("API_BASE_URL", "http://localhost:3000/api/v1")
	scope := strings.TrimSpace(getenv("API_TOKEN_SCOPE", "assets.read"))
	if scope == "" {
		scope = "assets.read"
	}
	return &Client{
		Base:         base,
		Token:        strings.TrimSpace(getenv("API_AUTH_TOKEN", "")),
		HC:           &http.Client{Timeout: 10 * time.Second},
		tokenURL:     strings.TrimSpace(getenv("API_AUTH_URL", deriveTokenURL(base))),
		clientID:     strings.TrimSpace(os.Getenv("API_CLIENT_ID")),
		clientSecret: strings.TrimSpace(os.Getenv("API_CLIENT_SECRET")),
		scope:        scope,
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
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Address   string   `json:"address"`
	Price     float64  `json:"price"`
	Currency  string   `json:"currency"`
	Images    []string `json:"images"`
	Badges    []string `json:"badges"`
	Bedrooms  int      `json:"bedrooms"`
	Bathrooms int      `json:"bathrooms"`
	Area      int      `json:"area"` // in square feet
	Parking   int      `json:"parking"`
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
	params := buildAssetSearchParams(q)
	res, err := c.doGet("/assets", params)
	if err != nil {
		return c.getMockSearchResults(q), nil
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return c.getMockSearchResults(q), nil
	}

	var payload assetListResponse
	dec := json.NewDecoder(res.Body)
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return c.getMockSearchResults(q), nil
	}

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
		status = "listed_rental,listed_sale"
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
	if c.Token != "" {
		return fmt.Sprintf("Bearer %s", c.Token)
	}
	token, err := c.getOAuthToken()
	if err != nil {
		log.Printf("api: oauth token error: %v", err)
		return ""
	}
	if token == "" {
		return ""
	}
	return fmt.Sprintf("Bearer %s", token)
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
		return c.cachedToken, nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	if c.scope != "" {
		data.Set("scope", c.scope)
	}

	req, err := http.NewRequest(http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.HC.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return "", fmt.Errorf("oauth token: %s %s", res.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", fmt.Errorf("oauth token: empty access_token")
	}

	expiresIn := time.Duration(payload.ExpiresIn) * time.Second
	if expiresIn <= 0 {
		expiresIn = 30 * time.Minute
	}
	refreshBefore := 30 * time.Second
	if expiresIn <= refreshBefore {
		refreshBefore = expiresIn / 10
	}

	c.cachedToken = payload.AccessToken
	c.tokenExpiry = time.Now().Add(expiresIn - refreshBefore)

	return payload.AccessToken, nil
}

func mapAssetToProperty(raw map[string]any) Property {
	if raw == nil {
		return Property{}
	}

	details := pickMap(raw, "Details", "details")
	location := pickMap(raw, "Location", "location")

	prop := Property{
		ID:       firstString(raw, "ID", "id"),
		Currency: "৳",
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

	prop.Images = selectPhotoURLs(pickSlice(raw, "photos", "Photos"))
	if len(prop.Images) == 0 {
		prop.Images = []string{"/assets/hero-image.png"}
	}

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
	}

	if prop.Price == 0 {
		if v, ok := floatFrom(raw, "rent_price", "RentPrice", "monthly_rent"); ok {
			prop.Price = v
		}
	}

	badges := []string{
		titleize(firstString(raw, "Type", "type")),
		titleize(firstString(raw, "Status", "status")),
		titleize(firstString(location, "city")),
		titleize(firstString(location, "neighborhood")),
		titleize(firstString(details, "furnishingStatus", "furnishing_status")),
	}
	prop.Badges = dedupStrings(badges)

	return prop
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
	location := q.Get("location")
	propertyType := q.Get("type")
	area := q.Get("area")

	mockProperties := []Property{
		{
			ID:        "1",
			Title:     "Service Apt. Uttara South",
			Address:   "Uttara South, Dhaka",
			Price:     105000,
			Currency:  "৳",
			Images:    []string{"/assets/db6726f48a0bae50917980327e8ff5eb40ae871e.png"},
			Badges:    []string{"For Sale", "Verified", "Residential"},
			Bedrooms:  3,
			Bathrooms: 2,
			Area:      1200,
			Parking:   1,
		},
		{
			ID:        "2",
			Title:     "Residental Apt. Uttara South Sec 10",
			Address:   "Uttara South Sec 10, Dhaka",
			Price:     120000,
			Currency:  "৳",
			Images:    []string{"/assets/8abeccd3fd2f4096a7b4a66a184c5ae36074637a.png"},
			Badges:    []string{"To-let", "Verified", "Hostel"},
			Bedrooms:  3,
			Bathrooms: 2,
			Area:      1200,
			Parking:   1,
		},
		{
			ID:        "3",
			Title:     "Residental Apt. Uttara South Sec 9",
			Address:   "Uttara South Sec 9, Dhaka",
			Price:     50000,
			Currency:  "৳",
			Images:    []string{"/assets/1f002be890c252fab41bc52a14801210d4fa2535.png"},
			Badges:    []string{"To-let", "Verified", "Short Term Rental"},
			Bedrooms:  3,
			Bathrooms: 2,
			Area:      1200,
			Parking:   1,
		},
		{
			ID:        "4",
			Title:     "Office Space Uttara South Sec 12",
			Address:   "Uttara South Sec 12, Dhaka",
			Price:     5500000,
			Currency:  "৳",
			Images:    []string{"/assets/d466fbc3c6a3829176f4bf45c88ed96204288a39.png"},
			Badges:    []string{"For Sale", "Verified", "Office Space"},
			Bedrooms:  3,
			Bathrooms: 2,
			Area:      1200,
			Parking:   1,
		},
		{
			ID:        "5",
			Title:     "Furnished Apt. Uttara North Sec 12",
			Address:   "Uttara North Sec 12, Dhaka",
			Price:     5500000,
			Currency:  "৳",
			Images:    []string{"/assets/2f8fe8dfbde9fb83f633da9c0e8bdff775034700.png"},
			Badges:    []string{"For Sale", "Verified", "Long Term Rental"},
			Bedrooms:  3,
			Bathrooms: 2,
			Area:      1200,
			Parking:   1,
		},
	}

	filtered := []Property{}
	for _, prop := range mockProperties {
		if location != "" && !contains(prop.Address, location) && !contains(prop.Title, location) {
			continue
		}
		if propertyType != "" && !containsAny(prop.Badges, propertyType) {
			continue
		}
		if area != "" && !contains(prop.Address, area) && !contains(prop.Title, area) {
			continue
		}
		filtered = append(filtered, prop)
	}

	if len(filtered) == 0 {
		filtered = mockProperties
	}

	return PropertyList{
		Items: filtered,
		Page:  1,
		Pages: 1,
		Total: len(filtered),
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

func (c *Client) GetProperty(id string) (Property, error) {
	var out Property
	if id == "" {
		return out, fmt.Errorf("property id required")
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
