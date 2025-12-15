package ipapi

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
		if r.URL.Path != "/8.8.8.8" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"status": "success",
			"country": "United States",
			"countryCode": "US",
			"region": "CA",
			"regionName": "California",
			"city": "Mountain View",
			"lat": 37.386,
			"lon": -122.084,
			"isp": "Google LLC",
			"org": "Google Public DNS",
			"as": "AS15169 Google LLC",
			"query": "8.8.8.8"
		}`))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("8.8.8.8")

	geo, err := client.Check(context.Background(), ip)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
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
	if geo.Org != "Google Public DNS" {
		t.Errorf("Org = %v, want Google Public DNS", geo.Org)
	}
	if geo.ASN != "AS15169 Google LLC" {
		t.Errorf("ASN = %v, want AS15169 Google LLC", geo.ASN)
	}
}

func TestClient_Check_IPv6(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// IPv6 address in URL path
		if r.URL.Path != "/2001:4860:4860::8888" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"status": "success",
			"country": "United States",
			"countryCode": "US",
			"regionName": "California",
			"city": "Mountain View",
			"lat": 37.386,
			"lon": -122.084,
			"isp": "Google LLC",
			"org": "Google",
			"as": "AS15169",
			"query": "2001:4860:4860::8888"
		}`))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("2001:4860:4860::8888")

	geo, err := client.Check(context.Background(), ip)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	if geo.Country != "United States" {
		t.Errorf("Country = %v, want United States", geo.Country)
	}
}

func TestClient_Check_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"status": "fail",
			"message": "reserved range",
			"query": "127.0.0.1"
		}`))
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("127.0.0.1")

	_, err := client.Check(context.Background(), ip)
	if err == nil {
		t.Fatal("Check() expected error for reserved range")
	}

	if err.Error() != "API error: reserved range" {
		t.Errorf("error = %v, want 'API error: reserved range'", err)
	}
}

func TestClient_Check_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := New(WithBaseURL(server.URL + "/"))
	ip := model.MustParseIPAddress("8.8.8.8")

	_, err := client.Check(context.Background(), ip)
	if err == nil {
		t.Fatal("Check() expected error for HTTP 500")
	}
}

func TestClient_Check_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
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

func TestClient_Check_ConnectionError(t *testing.T) {
	// Use an invalid URL to simulate connection error
	client := New(WithBaseURL("http://localhost:1/"))
	ip := model.MustParseIPAddress("8.8.8.8")

	_, err := client.Check(context.Background(), ip)
	if err == nil {
		t.Fatal("Check() expected error for connection failure")
	}
}
