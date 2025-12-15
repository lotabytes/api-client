package model

import (
	"encoding/json"
	"time"
)

// ProviderResult represents the outcome of a single provider lookup.
// It captures either a successful result or an error.
type ProviderResult struct {
	Provider string        `json:"provider"`
	Result   *Geolocation  `json:"result,omitempty"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"-"`
}

// Success reports whether this provider lookup succeeded.
func (pr ProviderResult) Success() bool {
	return pr.Error == "" && pr.Result != nil
}

// MarshalJSON implements custom JSON marshalling to output duration as milliseconds.
func (pr ProviderResult) MarshalJSON() ([]byte, error) {
	type Alias ProviderResult
	return json.Marshal(struct {
		Alias
		Duration int64 `json:"duration_ms"`
	}{
		Alias:    Alias(pr),
		Duration: pr.Duration.Milliseconds(),
	})
}

// Report is the aggregated result of querying multiple providers
// for information about an IP address.
type Report struct {
	// IP is the address that was queried
	IP IPAddress `json:"ip"`

	// Timestamp when the report was generated
	Timestamp time.Time `json:"timestamp"`

	// Results from each provider
	Results []ProviderResult `json:"results"`

	// TotalDuration is how long the entire lookup took
	TotalDuration time.Duration `json:"-"`
}

// MarshalJSON implements custom JSON marshalling for Report.
func (r Report) MarshalJSON() ([]byte, error) {
	type Alias Report
	return json.Marshal(struct {
		Alias
		TotalDuration int64 `json:"total_duration_ms"`
	}{
		Alias:         Alias(r),
		TotalDuration: r.TotalDuration.Milliseconds(),
	})
}

// SuccessCount returns the number of providers that returned successfully.
func (r Report) SuccessCount() int {
	count := 0
	for _, pr := range r.Results {
		if pr.Success() {
			count++
		}
	}
	return count
}

// ErrorCount returns the number of providers that failed.
func (r Report) ErrorCount() int {
	return len(r.Results) - r.SuccessCount()
}

// SuccessfulResults returns only the successful provider results.
func (r Report) SuccessfulResults() []ProviderResult {
	results := make([]ProviderResult, 0, len(r.Results))
	for _, pr := range r.Results {
		if pr.Success() {
			results = append(results, pr)
		}
	}
	return results
}

// Consensus returns the most commonly agreed-upon values across providers.
// This is useful when providers return slightly different data.
func (r Report) Consensus() Geolocation {
	successful := r.SuccessfulResults()
	if len(successful) == 0 {
		return Geolocation{IP: r.IP}
	}

	// For simplicity, we use voting for string fields
	// and averaging for numeric fields
	countryVotes := make(map[string]int)
	countryCodeVotes := make(map[string]int)
	cityVotes := make(map[string]int)
	regionVotes := make(map[string]int)
	ispVotes := make(map[string]int)
	orgVotes := make(map[string]int)
	asnVotes := make(map[string]int)

	var latSum, lonSum float64
	var coordCount int

	for _, pr := range successful {
		if pr.Result == nil {
			continue
		}
		g := pr.Result

		if g.Country != "" {
			countryVotes[g.Country]++
		}
		if g.CountryCode != "" {
			countryCodeVotes[g.CountryCode]++
		}
		if g.City != "" {
			cityVotes[g.City]++
		}
		if g.Region != "" {
			regionVotes[g.Region]++
		}
		if g.ISP != "" {
			ispVotes[g.ISP]++
		}
		if g.Org != "" {
			orgVotes[g.Org]++
		}
		if g.ASN != "" {
			asnVotes[g.ASN]++
		}

		if g.HasLocation() {
			latSum += g.Latitude
			lonSum += g.Longitude
			coordCount++
		}
	}

	consensus := Geolocation{
		IP:          r.IP,
		Country:     mostVoted(countryVotes),
		CountryCode: mostVoted(countryCodeVotes),
		City:        mostVoted(cityVotes),
		Region:      mostVoted(regionVotes),
		ISP:         mostVoted(ispVotes),
		Org:         mostVoted(orgVotes),
		ASN:         mostVoted(asnVotes),
	}

	if coordCount > 0 {
		consensus.Latitude = latSum / float64(coordCount)
		consensus.Longitude = lonSum / float64(coordCount)
	}

	return consensus
}

// mostVoted returns the key with the highest vote count.
// In case of a tie, the result is deterministic but arbitrary.
func mostVoted(votes map[string]int) string {
	var best string
	var bestCount int

	for k, count := range votes {
		if count > bestCount || (count == bestCount && k < best) {
			best = k
			bestCount = count
		}
	}

	return best
}
