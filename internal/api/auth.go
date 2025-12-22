package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type AuthUser struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	PhoneNumber     string `json:"phone_number"`
	Role            string `json:"role"`
	Status          string `json:"status"`
	ImageURL        string `json:"image_url"`
	SelectedAssetID string `json:"selected_asset_id"`
}

type LoginResponse struct {
	Token string   `json:"token"`
	User  AuthUser `json:"user"`
}

type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// LoginUser authenticates a Nestlo user via email/password and returns the JWT + user profile.
func (c *Client) LoginUser(email, password string) (LoginResponse, error) {
	// Support mock mode to avoid hitting real API in local development.
	if c.mockEnabled && c.mockAuthEnabled {
		now := time.Now().Unix()
		return LoginResponse{
			Token: fmt.Sprintf("mock-token-%d", now),
			User: AuthUser{
				ID:              fmt.Sprintf("mock-user-%d", now),
				Name:            strings.Split(email, "@")[0],
				Email:           email,
				Role:            "tenant",
				Status:          "active",
				PhoneNumber:     "+8801700000000",
				SelectedAssetID: "",
			},
		}, nil
	}

	body, _ := json.Marshal(map[string]string{
		"email":    strings.TrimSpace(email),
		"password": password,
	})

	endp := c.buildURL("/auth/login", nil)
	req, err := http.NewRequest(http.MethodPost, endp, bytes.NewReader(body))
	if err != nil {
		return LoginResponse{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	start := time.Now()
	res, err := c.HC.Do(req)
	if err != nil {
		return LoginResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		detail, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		msg := strings.TrimSpace(string(detail))
		if msg == "" {
			msg = http.StatusText(res.StatusCode)
		}
		return LoginResponse{}, &APIError{
			StatusCode: res.StatusCode,
			Message:    msg,
		}
	}

	var payload LoginResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return LoginResponse{}, err
	}

	if payload.Token == "" {
		return LoginResponse{}, fmt.Errorf("login: missing token in response after %dms", time.Since(start).Milliseconds())
	}

	return payload, nil
}
