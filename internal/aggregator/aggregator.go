// Package aggregator provides concurrent IP geolocation lookups across multiple providers.
package aggregator

import (
	"context"
	"sync"
	"time"

	"api-client/internal/model"
	"api-client/internal/provider"
)

// Aggregator coordinates concurrent lookups across multiple Providers.
type Aggregator struct {
	providers []provider.Provider
}

// New creates a new Aggregator with the given providers.
func New(providers ...provider.Provider) *Aggregator {
	return &Aggregator{
		providers: providers,
	}
}

// Lookup queries all providers concurrently and returns an aggregated report.
// The context controls the overall timeout for all lookups.
// Individual provider failures do not cause the entire lookup to fail.
func (a *Aggregator) Lookup(ctx context.Context, ip model.IPAddress) model.Report {
	start := time.Now()

	report := model.Report{
		IP:        ip,
		Timestamp: start,
		Results:   make([]model.ProviderResult, len(a.providers)),
	}

	var wg sync.WaitGroup
	wg.Add(len(a.providers))

	for i, checker := range a.providers {
		go func(idx int, p provider.Provider) {
			defer wg.Done()

			providerStart := time.Now()
			result, err := p.Check(ctx, ip)
			duration := time.Since(providerStart)

			pr := model.ProviderResult{
				Provider: p.Name(),
				Duration: duration,
			}

			if err != nil {
				pr.Error = err.Error()
			} else {
				pr.Result = &result
			}

			report.Results[idx] = pr
		}(i, checker)
	}

	wg.Wait()
	report.TotalDuration = time.Since(start)

	return report
}

// ProviderCount returns the number of configured providers.
func (a *Aggregator) ProviderCount() int {
	return len(a.providers)
}

// ProviderNames returns the names of all configured providers.
func (a *Aggregator) ProviderNames() []string {
	names := make([]string, len(a.providers))
	for i, p := range a.providers {
		names[i] = p.Name()
	}
	return names
}
