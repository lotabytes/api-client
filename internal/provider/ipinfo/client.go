// Package ipinfo provides a client for the ipinfo.io geolocation service.
package ipinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"api-client/internal/model"
	"api-client/internal/provider"
)

const (
	// ProviderName identifies this provider in reports.
	ProviderName = "ipinfo"

	// BaseURL is the API endpoint.
	BaseURL = "https://ipinfo.io/"
)

var _ provider.Provider = &Client{}

// response represents the JSON structure returned by ipinfo.io.
type response struct {
	IP       string `json:"ip"`
	City     string `json:"city"`
	Region   string `json:"region"`
	Country  string `json:"country"` // Two-letter country code
	Loc      string `json:"loc"`     // "latitude,longitude"
	Org      string `json:"org"`     // "AS12345 Organization Name"
	Timezone string `json:"timezone"`
	// Error response fields
	Error *errorResponse `json:"error,omitempty"`
}

func (r response) toGeoLocation(ip model.IPAddress) model.Geolocation {
	geo := model.Geolocation{
		IP:          ip,
		CountryCode: r.Country,
		Region:      r.Region,
		City:        r.City,
	}

	// Parse location "lat,lon"
	if r.Loc != "" {
		lat, lon, err := parseLocation(r.Loc)
		if err == nil {
			geo.Latitude = lat
			geo.Longitude = lon
		}
	}

	if r.Org != "" {
		asn, org := parseOrg(r.Org)
		geo.ASN = asn
		geo.Org = org
		// ipinfo.io doesn't distinguish ISP from Org, so we use Org for both
		geo.ISP = org
	}

	return geo
}

type errorResponse struct {
	Title   string `json:"title"`
	Message string `json:"message"`
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
	url := c.baseURL + ip.String() + "/json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return model.Geolocation{}, fmt.Errorf("creating request: %w", err)
	}

	// ipinfo.io recommends setting Accept header
	req.Header.Set("Accept", "application/json")

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

	if apiResp.Error != nil {
		return model.Geolocation{}, fmt.Errorf("API error: %s - %s", apiResp.Error.Title, apiResp.Error.Message)
	}

	return apiResp.toGeoLocation(ip), nil
}

// parseLocation parses "latitude,longitude" string.
func parseLocation(loc string) (lat, lon float64, err error) {
	parts := strings.Split(loc, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid location format: %s", loc)
	}

	lat, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing latitude: %w", err)
	}

	lon, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("parsing longitude: %w", err)
	}

	return lat, lon, nil
}

// parseOrg parses "AS12345 Organization Name" into ASN and org name.
func parseOrg(org string) (asn, name string) {
	parts := strings.SplitN(org, " ", 2)
	// Check if first part looks like an ASN
	if strings.HasPrefix(parts[0], "AS") {
		asn = parts[0]
		if len(parts) > 1 {
			name = parts[1]
		}
		return asn, name
	}

	return "", org
}
