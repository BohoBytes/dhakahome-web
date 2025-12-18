package handlers

import (
	"os"
	"strings"
)

// defaultContactEmail picks CONTACT_EMAIL first, falls back to PROPERY_ENQUIRY_EMAIL or a sane default.
func defaultContactEmail() string {
	if email := strings.TrimSpace(os.Getenv("CONTACT_EMAIL")); email != "" {
		return email
	}
	if email := strings.TrimSpace(os.Getenv("PROPERY_ENQUIRY_EMAIL")); email != "" {
		return email
	}
	return "info@dhakahome.com"
}

// defaultContactPhone selects a phone number based on listing type and env fallbacks.
func defaultContactPhone(listingType string) string {
	switch normalizeListingTypeValue(listingType) {
	case "listed_rental":
		if phone := envContactPhone("CONTACT_PHONE_RENT"); phone != "" {
			return phone
		}
	case "listed_sale":
		if phone := envContactPhone("CONTACT_PHONE_SALES"); phone != "" {
			return phone
		}
	}
	return envContactPhone("CONTACT_PHONE_RENT")
}

func envContactPhone(key string) string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return ""
	}
	if normalized, err := normalizeBDPhone(raw); err == nil {
		return normalized
	}
	return raw
}

func normalizeListingTypeValue(v string) string {
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
