package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProviderResult_Success(t *testing.T) {
	tests := []struct {
		name   string
		result ProviderResult
		want   bool
	}{
		{
			name: "successful result",
			result: ProviderResult{
				Provider: "test",
				Result:   &Geolocation{Country: "US"},
			},
			want: true,
		},
		{
			name: "error result",
			result: ProviderResult{
				Provider: "test",
				Error:    "connection timeout",
			},
			want: false,
		},
		{
			name: "nil result without error",
			result: ProviderResult{
				Provider: "test",
			},
			want: false,
		},
		{
			name: "result with error",
			result: ProviderResult{
				Provider: "test",
				Result:   &Geolocation{Country: "US"},
				Error:    "partial failure",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.Success(); got != tt.want {
				t.Errorf("Success() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderResult_JSONMarshal(t *testing.T) {
	result := ProviderResult{
		Provider: "ip-api",
		Result: &Geolocation{
			IP:      MustParseIPAddress("8.8.8.8"),
			Country: "United States",
		},
		Duration: 150 * time.Millisecond,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if m["provider"] != "ip-api" {
		t.Errorf("provider = %v, want ip-api", m["provider"])
	}

	// Duration should be in milliseconds
	if m["duration_ms"] != float64(150) {
		t.Errorf("duration_ms = %v, want 150", m["duration_ms"])
	}
}

func TestReport_SuccessCount(t *testing.T) {
	report := Report{
		IP: MustParseIPAddress("8.8.8.8"),
		Results: []ProviderResult{
			{Provider: "a", Result: &Geolocation{Country: "US"}},
			{Provider: "b", Error: "timeout"},
			{Provider: "c", Result: &Geolocation{Country: "US"}},
		},
	}

	if got := report.SuccessCount(); got != 2 {
		t.Errorf("SuccessCount() = %v, want 2", got)
	}
}

func TestReport_ErrorCount(t *testing.T) {
	report := Report{
		IP: MustParseIPAddress("8.8.8.8"),
		Results: []ProviderResult{
			{Provider: "a", Result: &Geolocation{Country: "US"}},
			{Provider: "b", Error: "timeout"},
			{Provider: "c", Error: "not found"},
		},
	}

	if got := report.ErrorCount(); got != 2 {
		t.Errorf("ErrorCount() = %v, want 2", got)
	}
}

func TestReport_SuccessfulResults(t *testing.T) {
	report := Report{
		IP: MustParseIPAddress("8.8.8.8"),
		Results: []ProviderResult{
			{Provider: "a", Result: &Geolocation{Country: "US"}},
			{Provider: "b", Error: "timeout"},
			{Provider: "c", Result: &Geolocation{Country: "DE"}},
		},
	}

	successful := report.SuccessfulResults()
	if len(successful) != 2 {
		t.Fatalf("SuccessfulResults() returned %d results, want 2", len(successful))
	}

	if successful[0].Provider != "a" {
		t.Errorf("first result provider = %v, want a", successful[0].Provider)
	}
	if successful[1].Provider != "c" {
		t.Errorf("second result provider = %v, want c", successful[1].Provider)
	}
}

func TestReport_Consensus_AllAgree(t *testing.T) {
	ip := MustParseIPAddress("8.8.8.8")
	report := Report{
		IP: ip,
		Results: []ProviderResult{
			{
				Provider: "a",
				Result: &Geolocation{
					IP: ip, Country: "United States", CountryCode: "US",
					City: "Mountain View", ISP: "Google",
					Latitude: 37.0, Longitude: -122.0,
				},
			},
			{
				Provider: "b",
				Result: &Geolocation{
					IP: ip, Country: "United States", CountryCode: "US",
					City: "Mountain View", ISP: "Google",
					Latitude: 37.0, Longitude: -122.0,
				},
			},
		},
	}

	consensus := report.Consensus()

	if consensus.Country != "United States" {
		t.Errorf("Country = %v, want United States", consensus.Country)
	}
	if consensus.CountryCode != "US" {
		t.Errorf("CountryCode = %v, want US", consensus.CountryCode)
	}
	if consensus.City != "Mountain View" {
		t.Errorf("City = %v, want Mountain View", consensus.City)
	}
	if consensus.ISP != "Google" {
		t.Errorf("ISP = %v, want Google", consensus.ISP)
	}
	if consensus.Latitude != 37.0 {
		t.Errorf("Latitude = %v, want 37.0", consensus.Latitude)
	}
	if consensus.Longitude != -122.0 {
		t.Errorf("Longitude = %v, want -122.0", consensus.Longitude)
	}
}

func TestReport_Consensus_Voting(t *testing.T) {
	ip := MustParseIPAddress("8.8.8.8")
	report := Report{
		IP: ip,
		Results: []ProviderResult{
			{Provider: "a", Result: &Geolocation{IP: ip, Country: "United States", City: "Mountain View"}},
			{Provider: "b", Result: &Geolocation{IP: ip, Country: "United States", City: "San Jose"}},
			{Provider: "c", Result: &Geolocation{IP: ip, Country: "United States", City: "Mountain View"}},
		},
	}

	consensus := report.Consensus()

	// All agree on country
	if consensus.Country != "United States" {
		t.Errorf("Country = %v, want United States", consensus.Country)
	}

	// Mountain View should win (2 vs 1)
	if consensus.City != "Mountain View" {
		t.Errorf("City = %v, want Mountain View", consensus.City)
	}
}

func TestReport_Consensus_AverageCoordinates(t *testing.T) {
	ip := MustParseIPAddress("8.8.8.8")
	report := Report{
		IP: ip,
		Results: []ProviderResult{
			{Provider: "a", Result: &Geolocation{IP: ip, Latitude: 36.0, Longitude: -120.0}},
			{Provider: "b", Result: &Geolocation{IP: ip, Latitude: 38.0, Longitude: -124.0}},
		},
	}

	consensus := report.Consensus()

	// Average of 36 and 38
	if consensus.Latitude != 37.0 {
		t.Errorf("Latitude = %v, want 37.0", consensus.Latitude)
	}

	// Average of -120 and -124
	if consensus.Longitude != -122.0 {
		t.Errorf("Longitude = %v, want -122.0", consensus.Longitude)
	}
}

func TestReport_Consensus_NoResults(t *testing.T) {
	ip := MustParseIPAddress("8.8.8.8")
	report := Report{
		IP:      ip,
		Results: []ProviderResult{},
	}

	consensus := report.Consensus()

	if !consensus.IP.Equal(ip) {
		t.Errorf("IP = %v, want %v", consensus.IP, ip)
	}
	if consensus.Country != "" {
		t.Errorf("Country = %v, want empty", consensus.Country)
	}
}

func TestReport_Consensus_AllErrors(t *testing.T) {
	ip := MustParseIPAddress("8.8.8.8")
	report := Report{
		IP: ip,
		Results: []ProviderResult{
			{Provider: "a", Error: "timeout"},
			{Provider: "b", Error: "not found"},
		},
	}

	consensus := report.Consensus()

	if !consensus.IP.Equal(ip) {
		t.Errorf("IP = %v, want %v", consensus.IP, ip)
	}
	if consensus.Country != "" {
		t.Errorf("Country = %v, want empty", consensus.Country)
	}
}

func TestReport_Consensus_IgnoresErrors(t *testing.T) {
	ip := MustParseIPAddress("8.8.8.8")
	report := Report{
		IP: ip,
		Results: []ProviderResult{
			{Provider: "a", Result: &Geolocation{IP: ip, Country: "Germany"}},
			{Provider: "b", Error: "timeout"},
			{Provider: "c", Result: &Geolocation{IP: ip, Country: "Germany"}},
		},
	}

	consensus := report.Consensus()

	if consensus.Country != "Germany" {
		t.Errorf("Country = %v, want Germany", consensus.Country)
	}
}

func TestReport_JSONMarshal(t *testing.T) {
	report := Report{
		IP:            MustParseIPAddress("8.8.8.8"),
		Timestamp:     time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		TotalDuration: 250 * time.Millisecond,
		Results: []ProviderResult{
			{
				Provider: "test",
				Result:   &Geolocation{Country: "US"},
				Duration: 100 * time.Millisecond,
			},
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}

	if m["ip"] != "8.8.8.8" {
		t.Errorf("ip = %v, want 8.8.8.8", m["ip"])
	}

	if m["total_duration_ms"] != float64(250) {
		t.Errorf("total_duration_ms = %v, want 250", m["total_duration_ms"])
	}
}

func TestMostVoted(t *testing.T) {
	tests := []struct {
		name  string
		votes map[string]int
		want  string
	}{
		{
			name:  "empty map",
			votes: map[string]int{},
			want:  "",
		},
		{
			name:  "single entry",
			votes: map[string]int{"US": 1},
			want:  "US",
		},
		{
			name:  "clear winner",
			votes: map[string]int{"US": 3, "DE": 1},
			want:  "US",
		},
		{
			name:  "tie breaks alphabetically",
			votes: map[string]int{"US": 2, "DE": 2},
			want:  "DE", // D < U
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mostVoted(tt.votes); got != tt.want {
				t.Errorf("mostVoted() = %v, want %v", got, tt.want)
			}
		})
	}
}
