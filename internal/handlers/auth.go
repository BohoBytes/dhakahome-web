package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/BohoBytes/dhakahome-web/internal/api"
)

type loginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	in, err := parseLoginPayload(r)
	if err != nil {
		writeAuthJSON(w, http.StatusBadRequest, map[string]any{
			"error": "Invalid request payload.",
		})
		return
	}

	errs := validateLoginPayload(in)
	if len(errs) > 0 {
		writeAuthJSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return
	}

	client := api.New()
	auth, err := client.LoginUser(in.Email, in.Password)
	if err != nil {
		status := http.StatusBadGateway
		msg := "Login failed. Please try again."

		var nestErr *api.APIError
		if errors.As(err, &nestErr) {
			if nestErr.StatusCode > 0 {
				status = nestErr.StatusCode
			}

			if status == http.StatusUnauthorized {
				msg = "Invalid email or password."
			} else if strings.TrimSpace(nestErr.Message) != "" {
				msg = nestErr.Message
			}
		} else {
			log.Printf("nestlo login error: %v", err)
		}

		writeAuthJSON(w, status, map[string]any{
			"error": msg,
		})
		return
	}

	expiresAt := time.Now().Add(24 * time.Hour).UTC()

	writeAuthJSON(w, http.StatusOK, map[string]any{
		"token":     auth.Token,
		"user":      auth.User,
		"expiresAt": expiresAt.Format(time.RFC3339),
	})
}

func parseLoginPayload(r *http.Request) (loginPayload, error) {
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	if strings.Contains(ct, "application/json") {
		defer r.Body.Close()
		var payload loginPayload
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&payload); err != nil {
			return loginPayload{}, err
		}
		return payload, nil
	}

	if err := r.ParseForm(); err != nil {
		return loginPayload{}, err
	}

	return loginPayload{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}, nil
}

func validateLoginPayload(in loginPayload) map[string]string {
	errs := make(map[string]string)

	in.Email = strings.TrimSpace(in.Email)
	in.Password = strings.TrimSpace(in.Password)

	if !emailRegex.MatchString(in.Email) {
		errs["email"] = "Enter a valid email address."
	}

	if in.Password == "" {
		errs["password"] = "Password is required."
	}

	return errs
}

func writeAuthJSON(w http.ResponseWriter, status int, payload map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
