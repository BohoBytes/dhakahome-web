package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BohoBytes/dhakahome-web/internal/api"
)

func getProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if filepath.Base(wd) == "web" {
		return filepath.Join(wd, "..", "..")
	}
	return wd
}

func SubmitLead(w http.ResponseWriter, r *http.Request) {
	respondJSON := wantsJSON(r)

	in, err := parseLeadPayload(r)
	if err != nil {
		writeLeadError(w, respondJSON, http.StatusBadRequest, map[string]any{
			"error": "invalid request body",
		})
		return
	}

	clean, errs := validateLead(in)
	if len(errs) > 0 {
		writeLeadError(w, respondJSON, http.StatusBadRequest, map[string]any{"errors": errs})
		return
	}

	client := api.New()
	contactEmail := strings.TrimSpace(clean.ContactEmail)
	if contactEmail == "" {
		contactEmail = defaultContactEmail()
	}
	req := api.LeadReq{
		Name:         clean.Name,
		Email:        clean.Email,
		Phone:        clean.Phone,
		Message:      clean.Message,
		PropertyID:   clean.PropertyID,
		ContactEmail: contactEmail,
	}

	if err := client.SubmitLead(req); err != nil {
		log.Printf("lead submission failed: %v", err)
		writeLeadError(w, respondJSON, http.StatusBadGateway, map[string]any{
			"error": "could not submit lead",
		})
		return
	}

	if respondJSON {
		writeLeadJSON(w, http.StatusOK, map[string]any{"status": "ok"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type leadPayload struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Message      string `json:"message"`
	PropertyID   string `json:"propertyId"`
	ContactEmail string `json:"contactEmail"`
}

var emailRegex = regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)

func wantsJSON(r *http.Request) bool {
	accept := strings.ToLower(r.Header.Get("Accept"))
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	return strings.Contains(accept, "application/json") ||
		strings.Contains(ct, "application/json") ||
		strings.EqualFold(r.Header.Get("HX-Request"), "true") ||
		strings.EqualFold(r.Header.Get("X-Requested-With"), "xmlhttprequest")
}

func parseLeadPayload(r *http.Request) (leadPayload, error) {
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.Contains(ct, "application/json") {
		defer r.Body.Close()
		var payload leadPayload
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&payload); err != nil { // limit to 1MB
			return leadPayload{}, err
		}
		return payload, nil
	}

	if err := r.ParseForm(); err != nil {
		return leadPayload{}, err
	}

	return leadPayload{
		Name:         r.FormValue("name"),
		Email:        r.FormValue("email"),
		Phone:        r.FormValue("phone"),
		Message:      r.FormValue("message"),
		PropertyID:   r.FormValue("propertyId"),
		ContactEmail: r.FormValue("contactEmail"),
	}, nil
}

func validateLead(in leadPayload) (leadPayload, map[string]string) {
	errs := make(map[string]string)

	in.Name = strings.TrimSpace(in.Name)
	in.Email = strings.TrimSpace(in.Email)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Message = strings.TrimSpace(in.Message)
	in.PropertyID = strings.TrimSpace(in.PropertyID)
	in.ContactEmail = strings.TrimSpace(in.ContactEmail)

	if in.ContactEmail == "" {
		in.ContactEmail = defaultContactEmail()
	} else if !emailRegex.MatchString(in.ContactEmail) {
		errs["contactEmail"] = "Please provide a valid contact email."
	}

	if len(in.Name) < 2 {
		errs["name"] = "Please enter your name."
	}
	if !emailRegex.MatchString(in.Email) {
		errs["email"] = "Please enter a valid email."
	}

	phone, err := normalizeBDPhone(in.Phone)
	if err != nil {
		errs["phone"] = err.Error()
	} else {
		in.Phone = phone
	}

	if in.Message == "" {
		errs["message"] = "Please include a message."
	}

	return in, errs
}

func normalizeBDPhone(phone string) (string, error) {
	clean := strings.TrimSpace(strings.ToLower(phone))
	if clean == "" {
		return "", fmt.Errorf("Please provide your phone number.")
	}

	// remove common separators
	replacer := strings.NewReplacer(" ", "", "-", "", "(", "", ")", "", ".", "", "tel:", "")
	clean = replacer.Replace(clean)

	if strings.HasPrefix(clean, "+") {
		clean = strings.TrimPrefix(clean, "+")
	}

	if strings.HasPrefix(clean, "88") {
		clean = strings.TrimPrefix(clean, "88")
	}

	if !strings.HasPrefix(clean, "01") {
		return "", fmt.Errorf("Use a Bangladesh number starting with 01.")
	}

	if len(clean) != 11 {
		return "", fmt.Errorf("Bangladesh numbers must be 11 digits.")
	}

	// second digit (index 2) must be 3-9 (01X)
	if clean[2] < '3' || clean[2] > '9' {
		return "", fmt.Errorf("Use a valid Bangladesh mobile operator code.")
	}

	return "+880" + clean[1:], nil
}

func writeLeadError(w http.ResponseWriter, respondJSON bool, status int, payload map[string]any) {
	if respondJSON {
		writeLeadJSON(w, status, payload)
		return
	}

	http.Error(w, "Lead submission failed", status)
}

func writeLeadJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
