package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/BohoBytes/dhakahome-web/internal/api"
)

type Option struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type SearchDropdowns struct {
	TypeOptions         []Option
	CityOptions         []Option
	AreaOptions         []Option
	MaxPriceOptions     []Option
	ListingTypeOptions  []Option
	BedroomOptions      []Option
	BathroomOptions     []Option
	ParkingOptions      []Option
	CommercialParkings  []Option
	SquareFootOptions   []Option
	SelectedType        string
	SelectedCity        string
	SelectedArea        string
	SelectedMaxPrice    string
	SelectedPriceMin    string
	SelectedListingType string
	SelectedBedrooms    string
	SelectedBathrooms   string
	SelectedParking     string
	SelectedServiced    string
	SelectedSharedRoom  string
	SelectedAreaMin     string
	SelectedAreaMax     string
}

func withSearchData(r *http.Request, data map[string]any) map[string]any {
	if data == nil {
		data = map[string]any{}
	}
	data["Search"] = buildSearchDropdowns(r.URL.Query())
	if _, ok := data["Query"]; !ok {
		data["Query"] = r.URL.Query()
	}
	return data
}

func buildSearchDropdowns(q url.Values) SearchDropdowns {
	selectedType := sanitizeSelection(firstNonEmpty(q.Get("type"), q.Get("types")))
	selectedCity := sanitizeSelection(q.Get("city"))
	selectedArea := sanitizeSelection(firstNonEmpty(q.Get("neighborhood"), q.Get("area")))
	selectedMaxPrice := normalizePriceValue(firstNonEmpty(q.Get("price_max"), q.Get("maxPrice")))
	selectedPriceMin := normalizePriceValue(firstNonEmpty(q.Get("price_min"), q.Get("minPrice")))
	selectedListingType := sanitizeSelection(firstNonEmpty(q.Get("listing_type"), q.Get("listingType"), deriveListingTypeFromStatus(q.Get("status"))))
	selectedBedrooms := sanitizeSelection(q.Get("bedrooms"))
	selectedBathrooms := sanitizeSelection(q.Get("bathrooms"))
	selectedParking := sanitizeSelection(q.Get("parking"))
	selectedServiced := sanitizeSelection(q.Get("serviced"))
	selectedShared := sanitizeSelection(firstNonEmpty(q.Get("shared_room"), q.Get("sharedRoom")))
	selectedAreaMin := normalizePriceValue(q.Get("area_min"))
	selectedAreaMax := normalizePriceValue(q.Get("area_max"))

	cl := api.New()

	cityOptions := []Option{{Label: "Any", Value: ""}}
	if cities, err := cl.GetCities(); err == nil && len(cities) > 0 {
		for _, city := range cities {
			cityOptions = append(cityOptions, Option{Value: city, Label: city})
		}
	} else if err != nil {
		log.Printf("search dropdowns: cities fallback: %v", err)
	}

	areaOptions := []Option{{Label: "Any", Value: ""}}
	if selectedCity != "" {
		if areas, err := cl.GetNeighborhoods(selectedCity); err == nil && len(areas) > 0 {
			for _, area := range areas {
				areaOptions = append(areaOptions, Option{Value: area, Label: area})
			}
		} else if err != nil {
			log.Printf("search dropdowns: areas fallback for city=%s: %v", selectedCity, err)
		}
	}

	return SearchDropdowns{
		TypeOptions:         typeOptions(),
		CityOptions:         cityOptions,
		AreaOptions:         areaOptions,
		MaxPriceOptions:     maxPriceOptions(),
		ListingTypeOptions:  listingTypeOptions(),
		BedroomOptions:      bedroomOptions(),
		BathroomOptions:     bathroomOptions(),
		ParkingOptions:      parkingOptions(),
		CommercialParkings:  commercialParkingOptions(),
		SquareFootOptions:   squareFootOptions(),
		SelectedType:        selectedType,
		SelectedCity:        selectedCity,
		SelectedArea:        selectedArea,
		SelectedMaxPrice:    selectedMaxPrice,
		SelectedPriceMin:    selectedPriceMin,
		SelectedListingType: selectedListingType,
		SelectedBedrooms:    selectedBedrooms,
		SelectedBathrooms:   selectedBathrooms,
		SelectedParking:     selectedParking,
		SelectedServiced:    selectedServiced,
		SelectedSharedRoom:  selectedShared,
		SelectedAreaMin:     selectedAreaMin,
		SelectedAreaMax:     selectedAreaMax,
	}
}

func CitiesJSON(w http.ResponseWriter, r *http.Request) {
	cl := api.New()
	cities, err := cl.GetCities()
	if err != nil {
		log.Printf("cities endpoint: %v", err)
	}
	writeJSON(w, map[string]any{"data": cities})
}

func NeighborhoodsJSON(w http.ResponseWriter, r *http.Request) {
	city := sanitizeSelection(r.URL.Query().Get("city"))
	if city == "" {
		http.Error(w, "city is required", http.StatusBadRequest)
		return
	}

	cl := api.New()
	areas, err := cl.GetNeighborhoods(city)
	if err != nil {
		log.Printf("neighborhoods endpoint: %v", err)
	}
	writeJSON(w, map[string]any{"data": areas})
}

func writeJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func typeOptions() []Option {
	return []Option{
		{Label: "Any", Value: ""},
		{Label: "Residential", Value: "Residential"},
		{Label: "Commercial", Value: "Commercial"},
		{Label: "Hostel", Value: "Hostel"},
		{Label: "Land", Value: "Land"},
	}
}

func maxPriceOptions() []Option {
	return []Option{
		{Label: "Any", Value: ""},
		{Label: "৳১০,০০০", Value: "10000"},
		{Label: "৳৫০,০০০", Value: "50000"},
		{Label: "৳১,০০,০০০", Value: "100000"},
		{Label: "৳১০,০০,০০০", Value: "1000000"},
		{Label: "৳৫০,০০,০০০", Value: "5000000"},
		{Label: "৳১,০০,০০,০০০", Value: "10000000"},
		{Label: "৳১০,০০,০০,০০০", Value: "100000000"},
	}
}

func listingTypeOptions() []Option {
	return []Option{
		{Label: "Rent or Sale", Value: ""},
		{Label: "Rent", Value: "listed_rental"},
		{Label: "Sale", Value: "listed_sale"},
	}
}

func bedroomOptions() []Option {
	return []Option{
		{Label: "Any", Value: ""},
		{Label: "Studio / 0", Value: "0"},
		{Label: "1+", Value: "1"},
		{Label: "2+", Value: "2"},
		{Label: "3+", Value: "3"},
		{Label: "4+", Value: "4"},
		{Label: "5+", Value: "5"},
	}
}

func bathroomOptions() []Option {
	return []Option{
		{Label: "Any", Value: ""},
		{Label: "1+", Value: "1"},
		{Label: "2+", Value: "2"},
		{Label: "3+", Value: "3"},
		{Label: "4+", Value: "4"},
		{Label: "5+", Value: "5"},
	}
}

func parkingOptions() []Option {
	return []Option{
		{Label: "Any", Value: ""},
		{Label: "0", Value: "0"},
		{Label: "1", Value: "1"},
		{Label: "2", Value: "2"},
		{Label: "3+", Value: "3"},
	}
}

func commercialParkingOptions() []Option {
	return []Option{
		{Label: "Any", Value: ""},
		{Label: "0", Value: "0"},
		{Label: "1", Value: "1"},
		{Label: "2", Value: "2"},
		{Label: "3", Value: "3"},
		{Label: "4", Value: "4"},
		{Label: "5", Value: "5"},
		{Label: "6", Value: "6"},
		{Label: "7", Value: "7"},
		{Label: "8", Value: "8"},
		{Label: "9", Value: "9"},
		{Label: "10+", Value: "10"},
	}
}

func squareFootOptions() []Option {
	return []Option{
		{Label: "Any", Value: ""},
		{Label: "500+", Value: "500"},
		{Label: "800+", Value: "800"},
		{Label: "1,000+", Value: "1000"},
		{Label: "1,200+", Value: "1200"},
		{Label: "1,500+", Value: "1500"},
		{Label: "2,000+", Value: "2000"},
		{Label: "2,500+", Value: "2500"},
		{Label: "3,000+", Value: "3000"},
		{Label: "4,000+", Value: "4000"},
	}
}

func sanitizeSelection(v string) string {
	v = strings.TrimSpace(v)
	if strings.EqualFold(v, "any") {
		return ""
	}
	return v
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func normalizePriceValue(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
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
		return ""
	}

	return b.String()
}

func deriveListingTypeFromStatus(status string) string {
	clean := strings.TrimSpace(strings.ToLower(status))
	switch clean {
	case "listed_rental":
		return "listed_rental"
	case "listed_sale":
		return "listed_sale"
	}
	return ""
}
