package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"api-client/internal/model"
)

func makeTestReport() model.Report {
	ip := model.MustParseAddr("8.8.8.8")
	return model.Report{
		IP:        ip,
		Timestamp: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Results: []model.ProviderResult{
			{
				Provider: "provider1",
				Result: &model.Geolocation{
					IP:          ip,
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
				Duration: 100 * time.Millisecond,
			},
			{
				Provider: "provider2",
				Result: &model.Geolocation{
					IP:          ip,
					Country:     "United States",
					CountryCode: "US",
					Region:      "California",
					City:        "Mountain View",
					Latitude:    37.4,
					Longitude:   -122.1,
					ISP:         "Google",
					Org:         "Google Inc",
					ASN:         "AS15169",
				},
				Duration: 150 * time.Millisecond,
			},
		},
		TotalDuration: 180 * time.Millisecond,
	}
}

func makeTestReportWithError() model.Report {
	ip := model.MustParseAddr("8.8.8.8")
	return model.Report{
		IP:        ip,
		Timestamp: time.Now(),
		Results: []model.ProviderResult{
			{
				Provider: "success",
				Result: &model.Geolocation{
					IP:      ip,
					Country: "United States",
				},
				Duration: 100 * time.Millisecond,
			},
			{
				Provider: "failure",
				Error:    "connection timeout",
				Duration: 5000 * time.Millisecond,
			},
		},
		TotalDuration: 200 * time.Millisecond,
	}
}

func TestFormatter_FormatJSON(t *testing.T) {
	report := makeTestReport()

	var buf bytes.Buffer
	f := NewFormatter(&buf)

	err := f.Format(report, FormatJSON)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Verify it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}

	// Check key fields
	if parsed["ip"] != "8.8.8.8" {
		t.Errorf("ip = %v, want 8.8.8.8", parsed["ip"])
	}

	results, ok := parsed["results"].([]interface{})
	if !ok {
		t.Fatal("results should be an array")
	}

	if len(results) != 2 {
		t.Errorf("results length = %d, want 2", len(results))
	}
}

func TestFormatter_FormatText(t *testing.T) {
	report := makeTestReport()

	var buf bytes.Buffer
	f := NewFormatter(&buf)

	err := f.Format(report, FormatText)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()

	// Check header
	if !strings.Contains(output, "IP Intelligence Report for 8.8.8.8") {
		t.Error("output should contain header with IP address")
	}

	// Check consensus section
	if !strings.Contains(output, "CONSENSUS") {
		t.Error("output should contain CONSENSUS section")
	}

	// Check country appears
	if !strings.Contains(output, "United States") {
		t.Error("output should contain country name")
	}

	// Check provider details section
	if !strings.Contains(output, "PROVIDER DETAILS") {
		t.Error("output should contain PROVIDER DETAILS section")
	}

	// Check provider names appear
	if !strings.Contains(output, "[provider1]") {
		t.Error("output should contain provider1")
	}
	if !strings.Contains(output, "[provider2]") {
		t.Error("output should contain provider2")
	}

	// Check summary
	if !strings.Contains(output, "2/2 providers succeeded") {
		t.Error("output should contain success count")
	}
}

func TestFormatter_FormatText_WithError(t *testing.T) {
	report := makeTestReportWithError()

	var buf bytes.Buffer
	f := NewFormatter(&buf)

	err := f.Format(report, FormatText)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()

	// Check failed provider
	if !strings.Contains(output, "[failure] FAILED") {
		t.Error("output should show failed provider")
	}

	if !strings.Contains(output, "connection timeout") {
		t.Error("output should contain error message")
	}

	// Check summary shows partial success
	if !strings.Contains(output, "1/2 providers succeeded") {
		t.Error("output should show 1/2 success count")
	}
}

func TestFormatter_FormatText_EmptyReport(t *testing.T) {
	ip := model.MustParseAddr("8.8.8.8")
	report := model.Report{
		IP:            ip,
		Timestamp:     time.Now(),
		Results:       []model.ProviderResult{},
		TotalDuration: 10 * time.Millisecond,
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf)

	err := f.Format(report, FormatText)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()

	// Should still have header
	if !strings.Contains(output, "8.8.8.8") {
		t.Error("output should contain IP address")
	}

	// Summary should show 0/0
	if !strings.Contains(output, "0/0 providers succeeded") {
		t.Error("output should show 0/0 success count")
	}
}

func TestFormatter_Format_InvalidFormat(t *testing.T) {
	report := makeTestReport()

	var buf bytes.Buffer
	f := NewFormatter(&buf)

	err := f.Format(report, "invalid")
	if err == nil {
		t.Fatal("Format() should error for invalid format")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("error = %v, should mention 'unsupported format'", err)
	}
}

func TestFormatter_FormatText_Coordinates(t *testing.T) {
	ip := model.MustParseAddr("8.8.8.8")
	report := model.Report{
		IP:        ip,
		Timestamp: time.Now(),
		Results: []model.ProviderResult{
			{
				Provider: "test",
				Result: &model.Geolocation{
					IP:        ip,
					Latitude:  37.38605,
					Longitude: -122.08385,
				},
				Duration: 100 * time.Millisecond,
			},
		},
		TotalDuration: 100 * time.Millisecond,
	}

	var buf bytes.Buffer
	f := NewFormatter(&buf)

	err := f.Format(report, FormatText)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	output := buf.String()

	// Coordinates should be formatted to 4 decimal places
	if !strings.Contains(output, "37.3860") || !strings.Contains(output, "-122.0838") {
		t.Errorf("coordinates not formatted correctly, got: %s", output)
	}
}

func TestFormatter_FormatJSON_Duration(t *testing.T) {
	report := makeTestReport()

	var buf bytes.Buffer
	f := NewFormatter(&buf)

	err := f.Format(report, FormatJSON)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Duration should be in milliseconds
	if !strings.Contains(buf.String(), `"total_duration_ms": 180`) {
		t.Errorf("JSON should contain duration in ms, got: %s", buf.String())
	}
}
