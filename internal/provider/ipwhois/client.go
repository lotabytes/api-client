// Package ipwhois provides a client for the ipwhois.app geolocation service.
package ipwhois

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"api-client/internal/model"
	"api-client/internal/provider"
)

const (
	// ProviderName identifies this provider in reports.
	ProviderName = "ipwhois"

	// BaseURL is the API endpoint.
	BaseURL = "https://ipwhois.app/json/"
)

var _ provider.Provider = &Client{}

// response represents the JSON structure returned by ipwhois.app.
type response struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message,omitempty"`
	IP          string  `json:"ip"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	ASN         string  `json:"asn"`
}

func (r response) toGeoLocation(ip model.IPAddress) model.Geolocation {
	return model.Geolocation{
		IP:          ip,
		Country:     r.Country,
		CountryCode: r.CountryCode,
		Region:      r.Region,
		City:        r.City,
		Latitude:    r.Latitude,
		Longitude:   r.Longitude,
		ISP:         r.ISP,
		Org:         r.Org,
		ASN:         r.ASN,
	}
}

type Client struct {
	requester provider.HttpRequester
	baseURL   string
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL (useful for testing).
func WithBaseURL(url string) Option {
	return func(client *Client) {
		client.baseURL = url
	}
}

// New creates a new ip-api.com client.
func New(requester provider.HttpRequester, opts ...Option) *Client {
	c := &Client{
		requester: requester,
		baseURL:   BaseURL,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Name returns the provider name.
func (c *Client) Name() string {
	return ProviderName
}

// Check looks up geolocation data for the given IP address.
func (c *Client) Check(ctx context.Context, ip model.IPAddress) (model.Geolocation, error) {
	url := c.baseURL + ip.String()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.Geolocation{}, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.requester.Do(req)
	if err != nil {
		return model.Geolocation{}, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return model.Geolocation{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apiResp response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return model.Geolocation{}, fmt.Errorf("decoding response: %w", err)
	}

	if !apiResp.Success {
		msg := apiResp.Message
		if msg == "" {
			msg = "unknown error"
		}
		return model.Geolocation{}, fmt.Errorf("API error: %s", msg)
	}

	return apiResp.toGeoLocation(ip), nil
}
