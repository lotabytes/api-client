// Package ipapi provides a client for the ip-api.com geolocation service.
package ipapi

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
	ProviderName = "ip-api"

	// BaseURL is the API endpoint. HTTP is used for the free tier.
	BaseURL = "http://ip-api.com/json/"
)

// response represents the JSON structure returned by ip-api.com.
type response struct {
	Status      string  `json:"status"`
	Message     string  `json:"message,omitempty"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

func (r response) toGeoLocation(ip model.IPAddress) model.Geolocation {
	return model.Geolocation{
		IP:          ip,
		Country:     r.Country,
		CountryCode: r.CountryCode,
		Region:      r.RegionName,
		City:        r.City,
		Latitude:    r.Lat,
		Longitude:   r.Lon,
		ISP:         r.ISP,
		Org:         r.Org,
		ASN:         r.AS,
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

	if apiResp.Status != "success" {
		msg := apiResp.Message
		if msg == "" {
			msg = "unknown error"
		}
		return model.Geolocation{}, fmt.Errorf("API error: %s", msg)
	}

	return apiResp.toGeoLocation(ip), nil
}
