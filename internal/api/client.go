package api

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "time"
)

type Client struct {
    Base string
    HC   *http.Client
}

func New() *Client {
    return &Client{
        Base: getenv("API_BASE_URL", "http://localhost:3000/api/v1"),
        HC:   &http.Client{Timeout: 10 * time.Second},
    }
}

func getenv(k, d string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return d
}

type Property struct {
    ID       string   `json:"id"`
    Title    string   `json:"title"`
    Address  string   `json:"address"`
    Price    float64  `json:"price"`
    Currency string   `json:"currency"`
    Images   []string `json:"images"`
    Badges   []string `json:"badges"`
}

type PropertyList struct {
    Items []Property `json:"items"`
    Page  int        `json:"page"`
    Pages int        `json:"pages"`
    Total int        `json:"total"`
}

func (c *Client) SearchProperties(q url.Values) (PropertyList, error) {
    var out PropertyList
    endp := fmt.Sprintf("%s/properties?%s", c.Base, q.Encode())
    res, err := c.HC.Get(endp)
    if err != nil {
        return out, err
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
        return out, fmt.Errorf("api: %s", res.Status)
    }
    return out, json.NewDecoder(res.Body).Decode(&out)
}

func (c *Client) GetProperty(id string) (Property, error) {
    var out Property
    endp := fmt.Sprintf("%s/properties/%s", c.Base, id)
    res, err := c.HC.Get(endp)
    if err != nil {
        return out, err
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
        return out, fmt.Errorf("api: %s", res.Status)
    }
    return out, json.NewDecoder(res.Body).Decode(&out)
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
    endp := fmt.Sprintf("%s/leads", c.Base)
    b, _ := json.Marshal(in)
    res, err := c.HC.Post(endp, "application/json", bytes.NewReader(b))
    if err != nil {
        return err
    }
    defer res.Body.Close()
    if res.StatusCode >= 300 {
        return fmt.Errorf("lead: %s", res.Status)
    }
    return nil
}
