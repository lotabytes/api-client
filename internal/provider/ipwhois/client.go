// Package ipwhois provides a client for the ipwhois.app geolocation service.
package ipwhois

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"api-client/internal/model"
)

const (
	// ProviderName identifies this provider in reports.
	ProviderName = "ipwhois"

	// BaseURL is the API endpoint.
	BaseURL = "https://ipwhois.app/json/"
)

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

// Client is an ipwhois.app geolocation provider.
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(client *Client) {
		client.httpClient = c
	}
}

// WithBaseURL sets a custom base URL (useful for testing).
func WithBaseURL(url string) Option {
	return func(client *Client) {
		client.baseURL = url
	}
}

// New creates a new ipwhois.app client.
func New(opts ...Option) *Client {
	c := &Client{
		httpClient: http.DefaultClient,
		baseURL:    BaseURL,
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

	resp, err := c.httpClient.Do(req)
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

	return model.Geolocation{
		IP:          ip,
		Country:     apiResp.Country,
		CountryCode: apiResp.CountryCode,
		Region:      apiResp.Region,
		City:        apiResp.City,
		Latitude:    apiResp.Latitude,
		Longitude:   apiResp.Longitude,
		ISP:         apiResp.ISP,
		Org:         apiResp.Org,
		ASN:         apiResp.ASN,
	}, nil
}
