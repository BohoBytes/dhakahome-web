package handlers

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/BohoBytes/dhakahome-web/internal/api"
	"github.com/go-chi/chi/v5"
)

type FeaturedArea struct {
	Neighborhood string
	City         string
	Count        int
	Image        string
	SearchURL    string
}

// render parses ONLY the base layout + the requested page (+ partials as needed),
// so each page can define its own "content" without collisions.
func render(w http.ResponseWriter, topLevelTemplate string, pageFile string, data any) {
	log.Printf("Rendering template: %s with page: %s", topLevelTemplate, pageFile)
	if m, ok := data.(map[string]any); ok {
		if _, exists := m["GetStartedURL"]; !exists {
			m["GetStartedURL"] = getStartedURL()
		}
		data = m
	}
	t := template.Must(template.New(pageFile).Funcs(template.FuncMap{
		"eq":          func(a, b any) bool { return a == b },
		"formatPrice": formatPrice,
		"add":         add,
		"sub":         sub,
		"seq":         seq,
		"dict":        dict,
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/"+pageFile,
		"internal/views/partials/page-header.html",
		"internal/views/partials/header.html",
		"internal/views/partials/hero.html",
		"internal/views/partials/search-box.html",
		"internal/views/partials/search-results-list.html",
		"internal/views/partials/property-card.html",
		"internal/views/partials/property-badge.html",
		"internal/views/partials/property-stats.html",
		"internal/views/partials/pagination.html",
		"internal/views/partials/common-sections.html",
		"internal/views/partials/services.html",
		"internal/views/partials/why-dhakahome.html",
		"internal/views/partials/properties-by-area.html",
		"internal/views/partials/testimonials.html",
		"internal/views/partials/faq.html",
	))
	log.Printf("Templates parsed successfully")
	if err := t.ExecuteTemplate(w, topLevelTemplate, data); err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("Template executed successfully")
}

func Home(w http.ResponseWriter, r *http.Request) {
	log.Printf("Home handler called")
	w.Header().Set("Content-Type", "text/html")
	data := withSearchData(r, map[string]any{
		"List":             api.PropertyList{},
		"ShowResults":      false,
		"ActivePage":       "home",
		"ShortlistEnabled": true,
	})
	data["GetStartedURL"] = getStartedURL()
	data = withTopAreas(data)
	render(w, "pages/home.html", "home.html", data)
}

func SearchPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	cl := api.New()
	list, _ := cl.SearchProperties(q) // TODO: handle error, flash message
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.New("pages/search-results.html").Funcs(template.FuncMap{
		"eq":          func(a, b any) bool { return a == b },
		"formatPrice": formatPrice,
		"add":         add,
		"sub":         sub,
		"seq":         seq,
		"dict":        dict,
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/search-results.html",
		"internal/views/partials/page-header.html",
		"internal/views/partials/header.html",
		"internal/views/partials/hero.html",
		"internal/views/partials/search-box.html",
		"internal/views/partials/search-advanced-box.html",
		"internal/views/partials/common-sections.html",
		"internal/views/partials/services.html",
		"internal/views/partials/why-dhakahome.html",
		"internal/views/partials/properties-by-area.html",
		"internal/views/partials/testimonials.html",
		"internal/views/partials/faq.html",
		"internal/views/partials/search-results-list.html",
		"internal/views/partials/property-card.html",
		"internal/views/partials/property-badge.html",
		"internal/views/partials/pagination.html",
	))
	data := withSearchData(r, map[string]any{
		"List":             list,
		"Query":            q,
		"ActivePage":       "search",
		"ShowResults":      true,
		"ShortlistEnabled": true,
	})
	data["GetStartedURL"] = getStartedURL()
	data = withTopAreas(data)
	if err := t.ExecuteTemplate(w, "pages/search-results.html", data); err != nil {
		log.Printf("search page template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func PropertiesPage(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if strings.TrimSpace(q.Get("limit")) == "" {
		q.Set("limit", "24")
	}
	if strings.TrimSpace(q.Get("sort_by")) == "" {
		q.Set("sort_by", "price")
	}
	if strings.TrimSpace(q.Get("order")) == "" {
		q.Set("order", "desc")
	}
	cl := api.New()
	list, _ := cl.SearchProperties(q) // mock-backed in dev
	sortBy := strings.ToLower(strings.TrimSpace(q.Get("sort_by")))
	order := strings.ToLower(strings.TrimSpace(q.Get("order")))
	if sortBy == "price" && len(list.Items) > 1 {
		sort.SliceStable(list.Items, func(i, j int) bool {
			if order == "asc" {
				return list.Items[i].Price < list.Items[j].Price
			}
			return list.Items[i].Price > list.Items[j].Price
		})
	}
	w.Header().Set("Content-Type", "text/html")
	mapToken := strings.TrimSpace(os.Getenv("MAPBOX_PUBLIC_TOKEN"))
	mapStyle := strings.TrimSpace(os.Getenv("MAPBOX_STYLE_URL"))
	if mapStyle == "" {
		mapStyle = "mapbox://styles/mapbox/streets-v12"
	}
	data := withSearchData(r, map[string]any{
		"ActivePage":     "properties",
		"List":           list,
		"Query":          q,
		"MapEnabled":     mapToken != "",
		"MapboxToken":    mapToken,
		"MapboxStyle":    mapStyle,
		"MapDefaultLat":  envFloat("MAP_DEFAULT_LAT", 23.810332),
		"MapDefaultLng":  envFloat("MAP_DEFAULT_LNG", 90.412521),
		"MapDefaultZoom": envFloat("MAP_DEFAULT_ZOOM", 11.2),
	})
	data["GetStartedURL"] = getStartedURL()
	render(w, "pages/properties.html", "properties.html", data)
}

func PropertyPage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	cl := api.New()
	p, _ := cl.GetProperty(id) // TODO: handle error

	docs, _ := cl.GetRequiredDocuments(p.Type)

	enquiryEmail := strings.TrimSpace(os.Getenv("PROPERY_ENQUIRY_EMAIL"))
	if enquiryEmail == "" {
		enquiryEmail = "enquiry@dhakahome.com"
	}

	contactEmail := enquiryEmail
	if candidate := strings.TrimSpace(p.ContactEmail); candidate != "" {
		contactEmail = candidate
	}

	contactPhone := defaultContactPhone(p.ListingType)
	if contactPhone == "" {
		contactPhone = strings.TrimSpace(p.ContactPhone)
		if contactPhone != "" {
			if normalized, err := normalizeBDPhone(contactPhone); err == nil {
				contactPhone = normalized
			}
		}
	}
	if contactPhone == "" {
		if normalized, err := normalizeBDPhone("01877-721-579"); err == nil {
			contactPhone = normalized
		}
	}

	similarQuery := url.Values{}
	if p.Type != "" {
		similarQuery.Set("type", p.Type)
	}
	// grab a few extra so filtering by listing type still yields rows
	similarQuery.Set("limit", "12")

	similar := api.PropertyList{}
	if list, err := cl.SearchProperties(similarQuery); err == nil {
		filtered := make([]api.Property, 0, len(list.Items))
		for _, item := range list.Items {
			if item.ID == p.ID {
				continue
			}
			if p.ListingType != "" && !strings.EqualFold(item.ListingType, p.ListingType) {
				continue
			}
			filtered = append(filtered, item)
		}

		if len(filtered) == 0 {
			for _, item := range list.Items {
				if item.ID == p.ID {
					continue
				}
				filtered = append(filtered, item)
			}
		}

		if len(filtered) > 6 {
			filtered = filtered[:6]
		}

		similar = api.PropertyList{
			Items: filtered,
			Page:  1,
			Pages: 1,
			Total: len(filtered),
		}
	}

	data := withSearchData(r, map[string]any{
		"P":               p,
		"Similar":         similar,
		"SearchBoxLayout": "static",
		"ShowSimilar":     len(similar.Items) > 0,
		"ActivePage":      "home",
		"SimilarType":     p.Type,
		"SimilarListing":  p.ListingType,
		"Documents":       docs,
		"ContactEmail":    contactEmail,
		"ContactPhone":    contactPhone,
	})
	data["GetStartedURL"] = getStartedURL()
	render(w, "pages/property.html", "property.html", data)
}

func FAQPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("FAQ handler called")
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.New("pages/faq.html").Funcs(template.FuncMap{
		"eq": func(a, b any) bool { return a == b },
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/faq.html",
		"internal/views/partials/page-header.html",
		"internal/views/partials/header.html",
	))
	data := map[string]any{
		"ActivePage":    "faq",
		"GetStartedURL": getStartedURL(),
	}
	if err := t.ExecuteTemplate(w, "pages/faq.html", data); err != nil {
		log.Printf("FAQ template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func AboutUsPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("About Us handler called")
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.New("pages/about-us.html").Funcs(template.FuncMap{
		"eq": func(a, b any) bool { return a == b },
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/about-us.html",
		"internal/views/partials/page-header.html",
		"internal/views/partials/header.html",
	))
	data := map[string]any{
		"ActivePage":    "about",
		"GetStartedURL": getStartedURL(),
	}
	if err := t.ExecuteTemplate(w, "pages/about-us.html", data); err != nil {
		log.Printf("About Us template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func HotelsPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Hotels page handler called")
	w.Header().Set("Content-Type", "text/html")
	t := template.Must(template.New("pages/hotels.html").Funcs(template.FuncMap{
		"eq": func(a, b any) bool { return a == b },
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/hotels.html",
		"internal/views/partials/page-header.html",
		"internal/views/partials/header.html",
	))
	data := map[string]any{
		"ActivePage":    "hotels",
		"GetStartedURL": getStartedURL(),
	}
	if err := t.ExecuteTemplate(w, "pages/hotels.html", data); err != nil {
		log.Printf("Hotels template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func ContactUsPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Contact Us page handler called")
	w.Header().Set("Content-Type", "text/html")
	contactEmail := defaultContactEmail()
	t := template.Must(template.New("pages/contact-us.html").Funcs(template.FuncMap{
		"eq": func(a, b any) bool { return a == b },
	}).ParseFiles(
		"internal/views/layouts/base.html",
		"internal/views/pages/contact-us.html",
		"internal/views/partials/page-header.html",
		"internal/views/partials/header.html",
	))
	data := map[string]any{
		"ActivePage":   "contact",
		"ContactEmail": contactEmail,
	}
	data["GetStartedURL"] = getStartedURL()
	if err := t.ExecuteTemplate(w, "pages/contact-us.html", data); err != nil {
		log.Printf("Contact Us template execution error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func withTopAreas(data map[string]any) map[string]any {
	if data == nil {
		data = map[string]any{}
	}
	if _, exists := data["TopAreas"]; exists {
		return data
	}

	if areas := loadTopAreas(); len(areas) >= 4 {
		data["TopAreas"] = areas
	}

	return data
}

func loadTopAreas() []FeaturedArea {
	cl := api.New()
	stats, err := cl.GetTopNeighborhoods(10, defaultTopAreasCity())
	if err != nil {
		log.Printf("top areas: %v", err)
	}

	filtered := make([]api.NeighborhoodStat, 0, len(stats))
	for _, stat := range stats {
		if strings.TrimSpace(stat.Neighborhood) == "" {
			continue
		}
		filtered = append(filtered, stat)
	}

	if len(filtered) < 4 {
		log.Printf("top areas: insufficient data to render section (got %d)", len(filtered))
		return nil
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(filtered), func(i, j int) {
		filtered[i], filtered[j] = filtered[j], filtered[i]
	})

	selected := filtered
	if len(selected) > 4 {
		selected = selected[:4]
	}

	images := areaImagePool()
	areas := make([]FeaturedArea, 0, len(selected))
	for _, stat := range selected {
		city := stat.City
		if strings.TrimSpace(city) == "" {
			city = defaultTopAreasCity()
		}
		areas = append(areas, FeaturedArea{
			Neighborhood: stat.Neighborhood,
			City:         city,
			Count:        stat.Count,
			Image:        pickAreaImage(stat.Neighborhood, images),
			SearchURL:    buildAreaSearchURL(city, stat.Neighborhood),
		})
	}

	if len(areas) < 4 {
		log.Printf("top areas: unable to build 4 featured areas (got %d)", len(areas))
		return nil
	}

	return areas
}

func buildAreaSearchURL(city, neighborhood string) string {
	params := url.Values{}
	if strings.TrimSpace(city) != "" {
		params.Set("city", city)
	}
	if strings.TrimSpace(neighborhood) != "" {
		params.Set("area", neighborhood)
		params.Set("neighborhood", neighborhood)
	}
	if len(params) == 0 {
		return "/search"
	}
	return "/search?" + params.Encode()
}

func pickAreaImage(name string, pool []string) string {
	key := normalizeAreaKey(name)
	if img := areaImageByName()[key]; img != "" {
		return img
	}

	if len(pool) == 0 {
		return ""
	}
	sum := 0
	for _, r := range strings.ToLower(name) {
		sum += int(r)
	}
	return pool[sum%len(pool)]
}

func areaImagePool() []string {
	return []string{
		"/assets/images/areas/area1.png",
		"/assets/images/areas/area2.png",
		"/assets/images/areas/area3.png",
		"/assets/images/areas/area4.png",
	}
}

func defaultTopAreasCity() string {
	return "Dhaka"
}

func normalizeAreaKey(name string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(name)) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func areaImageByName() map[string]string {
	return map[string]string{
		"gulshan":     "/assets/images/areas/gulshan.png",
		"banani":      "/assets/images/areas/banani.png",
		"bashundhara": "/assets/images/areas/bashundhara.png",
		"dhanmondi":   "/assets/images/areas/dhanmondi.png",
		"mirpur":      "/assets/images/areas/mirpur.png",
		"uttara":      "/assets/images/areas/uttara.png",
		"baridhara":   "/assets/images/areas/baridhara.png",
		"niketon":     "/assets/images/areas/niketon.png",
		"mohakhali":   "/assets/images/areas/mohakhali.png",
		"motijheel":   "/assets/images/areas/motijheel.png",
		"agargaon":    "/assets/images/areas/agargaon.png",
	}
}

func formatPrice(price float64) string {
	if price == 0 {
		return "0"
	}
	s := strconv.FormatInt(int64(price), 10)
	if len(s) <= 3 {
		return s
	}

	parts := []string{s[len(s)-3:]}
	prefix := s[:len(s)-3]

	for len(prefix) > 0 {
		if len(prefix) <= 2 {
			parts = append(parts, prefix)
			break
		}
		parts = append(parts, prefix[len(prefix)-2:])
		prefix = prefix[:len(prefix)-2]
	}

	var out strings.Builder
	for i := len(parts) - 1; i >= 0; i-- {
		out.WriteString(parts[i])
		if i > 0 {
			out.WriteRune(',')
		}
	}
	return out.String()
}

func add(a, b int) int { return a + b }
func sub(a, b int) int { return a - b }

func dict(values ...any) map[string]any {
	result := make(map[string]any, len(values)/2)
	for i := 0; i+1 < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			continue
		}
		result[key] = values[i+1]
	}
	return result
}

func seq(start, end int) []int {
	if end < start {
		return []int{}
	}
	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}

func getStartedURL() string {
	val := strings.TrimSpace(os.Getenv("PORTAL_BASE_URL"))
	if val == "" {
		return "/"
	}
	return val
}

func envFloat(key string, def float64) float64 {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return def
	}
	if parsed, err := strconv.ParseFloat(val, 64); err == nil {
		return parsed
	}
	return def
}
