package ipinfo

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"api-client/internal/model"
)

func TestClient_Check_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/8.8.8.8/json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("missing or wrong Accept header: %s", r.Header.Get("Accept"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"ip": "8.8.8.8",
			"city": "Mountain View",
			"region": "California",
			"country": "US",
			"loc": "37.386,-122.084",
			"org": "AS15169 Google LLC",
			"timezone": "America/Los_Angeles"
		}`))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("8.8.8.8")

	geo, err := client.Check(context.Background(), ip)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if geo.CountryCode != "US" {
		t.Errorf("CountryCode = %v, want US", geo.CountryCode)
	}
	if geo.Region != "California" {
		t.Errorf("Region = %v, want California", geo.Region)
	}
	if geo.City != "Mountain View" {
		t.Errorf("City = %v, want Mountain View", geo.City)
	}
	if geo.Latitude != 37.386 {
		t.Errorf("Latitude = %v, want 37.386", geo.Latitude)
	}
	if geo.Longitude != -122.084 {
		t.Errorf("Longitude = %v, want -122.084", geo.Longitude)
	}
	if geo.ASN != "AS15169" {
		t.Errorf("ASN = %v, want AS15169", geo.ASN)
	}
	if geo.Org != "Google LLC" {
		t.Errorf("Org = %v, want Google LLC", geo.Org)
	}
}

func TestClient_Check_IPv6(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/2001:4860:4860::8888/json" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"ip": "2001:4860:4860::8888",
			"city": "Mountain View",
			"region": "California",
			"country": "US",
			"loc": "37.386,-122.084",
			"org": "AS15169 Google LLC"
		}`))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("2001:4860:4860::8888")

	geo, err := client.Check(context.Background(), ip)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if geo.CountryCode != "US" {
		t.Errorf("CountryCode = %v, want US", geo.CountryCode)
	}
}

func TestClient_Check_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"error": {
				"title": "Wrong ip",
				"message": "Please provide a valid IP address"
			}
		}`))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("127.0.0.1")

	_, err := client.Check(context.Background(), ip)
	if err == nil {
		t.Fatal("Check() expected error")
	}

	expected := "API error: Wrong ip - Please provide a valid IP address"
	if err.Error() != expected {
		t.Errorf("error = %v, want %v", err, expected)
	}
}

func TestClient_Check_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("8.8.8.8")

	_, err := client.Check(context.Background(), ip)
	if err == nil {
		t.Fatal("Check() expected error for HTTP 429")
	}
}

func TestClient_Check_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid`))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("8.8.8.8")

	_, err := client.Check(context.Background(), ip)
	if err == nil {
		t.Fatal("Check() expected error for invalid JSON")
	}
}

func TestClient_Check_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("8.8.8.8")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.Check(ctx, ip)
	if err == nil {
		t.Fatal("Check() expected error due to context timeout")
	}
}

func TestParseLocation(t *testing.T) {
	tests := []struct {
		name    string
		loc     string
		wantLat float64
		wantLon float64
		wantErr bool
	}{
		{
			name:    "valid location",
			loc:     "37.386,-122.084",
			wantLat: 37.386,
			wantLon: -122.084,
		},
		{
			name:    "with spaces",
			loc:     "37.386, -122.084",
			wantLat: 37.386,
			wantLon: -122.084,
		},
		{
			name:    "negative latitude",
			loc:     "-33.8688,151.2093",
			wantLat: -33.8688,
			wantLon: 151.2093,
		},
		{
			name:    "missing comma",
			loc:     "37.386 -122.084",
			wantErr: true,
		},
		{
			name:    "invalid latitude",
			loc:     "abc,-122.084",
			wantErr: true,
		},
		{
			name:    "invalid longitude",
			loc:     "37.386,xyz",
			wantErr: true,
		},
		{
			name:    "empty string",
			loc:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lon, err := parseLocation(tt.loc)

			if tt.wantErr {
				if err == nil {
					t.Error("parseLocation() expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseLocation() error = %v", err)
			}

			if lat != tt.wantLat {
				t.Errorf("latitude = %v, want %v", lat, tt.wantLat)
			}
			if lon != tt.wantLon {
				t.Errorf("longitude = %v, want %v", lon, tt.wantLon)
			}
		})
	}
}

func TestParseOrg(t *testing.T) {
	tests := []struct {
		name    string
		org     string
		wantASN string
		wantOrg string
	}{
		{
			name:    "standard format",
			org:     "AS15169 Google LLC",
			wantASN: "AS15169",
			wantOrg: "Google LLC",
		},
		{
			name:    "no ASN",
			org:     "Google LLC",
			wantASN: "",
			wantOrg: "Google LLC",
		},
		{
			name:    "ASN only",
			org:     "AS15169",
			wantASN: "AS15169",
			wantOrg: "",
		},
		{
			name:    "empty string",
			org:     "",
			wantASN: "",
			wantOrg: "",
		},
		{
			name:    "org with multiple spaces",
			org:     "AS12345 Some Long Organization Name",
			wantASN: "AS12345",
			wantOrg: "Some Long Organization Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asn, org := parseOrg(tt.org)

			if asn != tt.wantASN {
				t.Errorf("ASN = %v, want %v", asn, tt.wantASN)
			}
			if org != tt.wantOrg {
				t.Errorf("org = %v, want %v", org, tt.wantOrg)
			}
		})
	}
}
