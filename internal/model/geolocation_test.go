package model

import (
	"encoding/json"
	"testing"
)

func TestGeolocation_HasLocation(t *testing.T) {
	tests := []struct {
		name string
		geo  Geolocation
		want bool
	}{
		{
			name: "has latitude and longitude",
			geo:  Geolocation{Latitude: 37.7749, Longitude: -122.4194},
			want: true,
		},
		{
			name: "has only latitude",
			geo:  Geolocation{Latitude: 37.7749},
			want: true,
		},
		{
			name: "has only longitude",
			geo:  Geolocation{Longitude: -122.4194},
			want: true,
		},
		{
			name: "zero coordinates",
			geo:  Geolocation{},
			want: false,
		},
		{
			name: "explicit zero coordinates",
			geo:  Geolocation{Latitude: 0, Longitude: 0},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.geo.HasLocation(); got != tt.want {
				t.Errorf("HasLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeolocation_HasNetworkInfo(t *testing.T) {
	tests := []struct {
		name string
		geo  Geolocation
		want bool
	}{
		{
			name: "has ISP",
			geo:  Geolocation{ISP: "Google LLC"},
			want: true,
		},
		{
			name: "has Org",
			geo:  Geolocation{Org: "Google"},
			want: true,
		},
		{
			name: "has ASN",
			geo:  Geolocation{ASN: "AS15169"},
			want: true,
		},
		{
			name: "has all network info",
			geo:  Geolocation{ISP: "Google LLC", Org: "Google", ASN: "AS15169"},
			want: true,
		},
		{
			name: "no network info",
			geo:  Geolocation{Country: "United States"},
			want: false,
		},
		{
			name: "empty geolocation",
			geo:  Geolocation{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.geo.HasNetworkInfo(); got != tt.want {
				t.Errorf("HasNetworkInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeolocation_IsEmpty(t *testing.T) {
	tests := []struct {
		name string
		geo  Geolocation
		want bool
	}{
		{
			name: "completely empty",
			geo:  Geolocation{},
			want: true,
		},
		{
			name: "only IP set",
			geo:  Geolocation{IP: MustParseAddr("8.8.8.8")},
			want: true, // IP doesn't count
		},
		{
			name: "has country",
			geo:  Geolocation{Country: "US"},
			want: false,
		},
		{
			name: "has coordinates",
			geo:  Geolocation{Latitude: 1.0},
			want: false,
		},
		{
			name: "has ISP",
			geo:  Geolocation{ISP: "Test"},
			want: false,
		},
		{
			name: "fully populated",
			geo: Geolocation{
				IP:          MustParseAddr("8.8.8.8"),
				Country:     "United States",
				CountryCode: "US",
				Region:      "California",
				City:        "Mountain View",
				Latitude:    37.386,
				Longitude:   -122.084,
				ISP:         "Google LLC",
				Org:         "Google",
				ASN:         "AS15169",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.geo.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeolocation_JSONMarshal(t *testing.T) {
	geo := Geolocation{
		IP:          MustParseAddr("8.8.8.8"),
		Country:     "United States",
		CountryCode: "US",
		Region:      "California",
		City:        "Mountain View",
		Latitude:    37.386,
		Longitude:   -122.084,
		ISP:         "Google LLC",
		Org:         "Google",
		ASN:         "AS15169",
	}

	data, err := json.Marshal(geo)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Verify it's valid JSON by unmarshalling into a map
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	// Check a few key fields
	if m["ip"] != "8.8.8.8" {
		t.Errorf("ip = %v, want 8.8.8.8", m["ip"])
	}
	if m["country"] != "United States" {
		t.Errorf("country = %v, want United States", m["country"])
	}
	if m["country_code"] != "US" {
		t.Errorf("country_code = %v, want US", m["country_code"])
	}
}

func TestGeolocation_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"ip": "8.8.8.8",
		"country": "United States",
		"country_code": "US",
		"region": "California",
		"city": "Mountain View",
		"latitude": 37.386,
		"longitude": -122.084,
		"isp": "Google LLC",
		"org": "Google",
		"asn": "AS15169"
	}`

	var geo Geolocation
	if err := json.Unmarshal([]byte(jsonData), &geo); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if geo.IP.String() != "8.8.8.8" {
		t.Errorf("IP = %v, want 8.8.8.8", geo.IP)
	}
	if geo.Country != "United States" {
		t.Errorf("Country = %v, want United States", geo.Country)
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
	if geo.ISP != "Google LLC" {
		t.Errorf("ISP = %v, want Google LLC", geo.ISP)
	}
	if geo.Org != "Google" {
		t.Errorf("Org = %v, want Google", geo.Org)
	}
	if geo.ASN != "AS15169" {
		t.Errorf("ASN = %v, want AS15169", geo.ASN)
	}
}

func TestGeolocation_JSONRoundTrip(t *testing.T) {
	original := Geolocation{
		IP:          MustParseAddr("2001:4860:4860::8888"),
		Country:     "United States",
		CountryCode: "US",
		Region:      "California",
		City:        "Mountain View",
		Latitude:    37.386,
		Longitude:   -122.084,
		ISP:         "Google LLC",
		Org:         "Google",
		ASN:         "AS15169",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Geolocation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Compare all fields
	if original.IP.Compare(decoded.IP) != 0 {
		t.Errorf("IP mismatch: got %v, want %v", decoded.IP, original.IP)
	}
	if original.Country != decoded.Country {
		t.Errorf("Country mismatch: got %v, want %v", decoded.Country, original.Country)
	}
	if original.CountryCode != decoded.CountryCode {
		t.Errorf("CountryCode mismatch: got %v, want %v", decoded.CountryCode, original.CountryCode)
	}
	if original.Region != decoded.Region {
		t.Errorf("Region mismatch: got %v, want %v", decoded.Region, original.Region)
	}
	if original.City != decoded.City {
		t.Errorf("City mismatch: got %v, want %v", decoded.City, original.City)
	}
	if original.Latitude != decoded.Latitude {
		t.Errorf("Latitude mismatch: got %v, want %v", decoded.Latitude, original.Latitude)
	}
	if original.Longitude != decoded.Longitude {
		t.Errorf("Longitude mismatch: got %v, want %v", decoded.Longitude, original.Longitude)
	}
	if original.ISP != decoded.ISP {
		t.Errorf("ISP mismatch: got %v, want %v", decoded.ISP, original.ISP)
	}
	if original.Org != decoded.Org {
		t.Errorf("Org mismatch: got %v, want %v", decoded.Org, original.Org)
	}
	if original.ASN != decoded.ASN {
		t.Errorf("ASN mismatch: got %v, want %v", decoded.ASN, original.ASN)
	}
}

func TestGeolocation_JSONEmptyValues(t *testing.T) {
	// Test that empty/zero values serialize correctly
	geo := Geolocation{
		IP:      MustParseAddr("8.8.8.8"),
		Country: "United States",
		// All other fields are zero values
	}

	data, err := json.Marshal(geo)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var decoded Geolocation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if decoded.City != "" {
		t.Errorf("City should be empty, got %q", decoded.City)
	}
	if decoded.Latitude != 0 {
		t.Errorf("Latitude should be 0, got %v", decoded.Latitude)
	}
}
