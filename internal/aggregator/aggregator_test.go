package aggregator

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"api-client/internal/model"
	"api-client/internal/provider"
)

func TestAggregator_Lookup_AllSuccess(t *testing.T) {
	ip := model.MustParseIPAddress("8.8.8.8")

	p1 := provider.NewTestProvider("provider1", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		return model.Geolocation{IP: ip, Country: "United States"}, nil
	}))
	p2 := provider.NewTestProvider("provider2", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		return model.Geolocation{IP: ip, Country: "United States"}, nil
	}))

	agg := New(p1, p2)
	report := agg.Lookup(context.Background(), ip)

	if !report.IP.Equal(ip) {
		t.Errorf("IP = %v, want %v", report.IP, ip)
	}

	if report.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	if len(report.Results) != 2 {
		t.Fatalf("Results count = %d, want 2", len(report.Results))
	}

	if report.SuccessCount() != 2 {
		t.Errorf("SuccessCount() = %d, want 2", report.SuccessCount())
	}

	if report.ErrorCount() != 0 {
		t.Errorf("ErrorCount() = %d, want 0", report.ErrorCount())
	}

	if report.TotalDuration == 0 {
		t.Error("TotalDuration should not be zero")
	}
}

func TestAggregator_Lookup_PartialFailure(t *testing.T) {
	ip := model.MustParseIPAddress("8.8.8.8")

	p1 := provider.NewTestProvider("success", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		return model.Geolocation{IP: ip, Country: "United States"}, nil
	}))
	p2 := provider.NewTestProvider("failure", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		return model.Geolocation{}, errors.New("connection timeout")
	}))

	agg := New(p1, p2)
	report := agg.Lookup(context.Background(), ip)

	if report.SuccessCount() != 1 {
		t.Errorf("SuccessCount() = %d, want 1", report.SuccessCount())
	}

	if report.ErrorCount() != 1 {
		t.Errorf("ErrorCount() = %d, want 1", report.ErrorCount())
	}

	// Find the failed result
	var failedResult model.ProviderResult
	for _, r := range report.Results {
		if r.Provider == "failure" {
			failedResult = r
			break
		}
	}

	if failedResult.Error != "connection timeout" {
		t.Errorf("Error = %q, want 'connection timeout'", failedResult.Error)
	}

	if failedResult.Result != nil {
		t.Error("Failed result should have nil Result")
	}
}

func TestAggregator_Lookup_AllFailure(t *testing.T) {
	ip := model.MustParseIPAddress("8.8.8.8")

	p1 := provider.NewTestProvider("fail1", provider.CheckerFunc(func(ctx context.Context, ip model.IPAddress) (model.Geolocation,
		error) {
		return model.Geolocation{}, errors.New("error 1")
	}))
	p2 := provider.NewTestProvider("fail2", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		return model.Geolocation{}, errors.New("error 2")
	}))

	agg := New(p1, p2)
	report := agg.Lookup(context.Background(), ip)

	if report.SuccessCount() != 0 {
		t.Errorf("SuccessCount() = %d, want 0", report.SuccessCount())
	}

	if report.ErrorCount() != 2 {
		t.Errorf("ErrorCount() = %d, want 2", report.ErrorCount())
	}
}

func TestAggregator_Lookup_Concurrent(t *testing.T) {
	ip := model.MustParseIPAddress("8.8.8.8")

	// Track call times to verify concurrent execution
	var callCount int32
	var maxConcurrent int32
	var current int32

	makeProvider := func(name string) provider.Provider {
		return provider.NewTestProvider(name, provider.CheckerFunc(func(ctx context.Context,
			ip model.IPAddress) (model.Geolocation, error) {
			atomic.AddInt32(&callCount, 1)
			c := atomic.AddInt32(&current, 1)

			// Update max concurrent
			for {
				maxC := atomic.LoadInt32(&maxConcurrent)
				if c <= maxC || atomic.CompareAndSwapInt32(&maxConcurrent, maxC, c) {
					break
				}
			}

			time.Sleep(50 * time.Millisecond)
			atomic.AddInt32(&current, -1)

			return model.Geolocation{IP: ip}, nil
		}))
	}

	agg := New(makeProvider("p1"), makeProvider("p2"), makeProvider("p3"))
	start := time.Now()
	report := agg.Lookup(context.Background(), ip)
	elapsed := time.Since(start)

	// All 3 providers should have been called
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}

	// Should have had at least 2 running concurrently
	if atomic.LoadInt32(&maxConcurrent) < 2 {
		t.Errorf("maxConcurrent = %d, want at least 2", maxConcurrent)
	}

	// If running sequentially, would take ~150ms. Concurrent should be ~50ms
	if elapsed > 100*time.Millisecond {
		t.Errorf("elapsed = %v, want < 100ms (providers running concurrently)", elapsed)
	}

	if len(report.Results) != 3 {
		t.Errorf("Results count = %d, want 3", len(report.Results))
	}
}

func TestAggregator_Lookup_ContextCancellation(t *testing.T) {
	ip := model.MustParseIPAddress("8.8.8.8")
	p := provider.NewTestProvider("slow", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		select {
		case <-ctx.Done():
			return model.Geolocation{}, ctx.Err()
		case <-time.After(1 * time.Second):
			return model.Geolocation{IP: ip}, nil
		}
	}))

	agg := New(p)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	report := agg.Lookup(ctx, ip)

	if report.SuccessCount() != 0 {
		t.Errorf("SuccessCount() = %d, want 0 (context should cancel)", report.SuccessCount())
	}

	if report.ErrorCount() != 1 {
		t.Errorf("ErrorCount() = %d, want 1", report.ErrorCount())
	}

	// Verify the error is context-related
	if report.Results[0].Error != context.DeadlineExceeded.Error() {
		t.Error("should have 'context exceeded' error message")
	}
}

func TestAggregator_Lookup_NoProviders(t *testing.T) {
	ip := model.MustParseIPAddress("8.8.8.8")

	agg := New() // No providers
	report := agg.Lookup(context.Background(), ip)

	if !report.IP.Equal(ip) {
		t.Errorf("IP = %v, want %v", report.IP, ip)
	}

	if len(report.Results) != 0 {
		t.Errorf("Results count = %d, want 0", len(report.Results))
	}

	if report.SuccessCount() != 0 {
		t.Errorf("SuccessCount() = %d, want 0", report.SuccessCount())
	}
}

func TestAggregator_Lookup_PreservesOrder(t *testing.T) {
	ip := model.MustParseIPAddress("8.8.8.8")

	// Create providers with different delays to test order preservation
	p1 := provider.NewTestProvider("first", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		time.Sleep(30 * time.Millisecond)
		return model.Geolocation{IP: ip}, nil
	}))
	p2 := provider.NewTestProvider("second", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		time.Sleep(10 * time.Millisecond) // Completes first
		return model.Geolocation{IP: ip}, nil
	}))
	p3 := provider.NewTestProvider("third", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		time.Sleep(20 * time.Millisecond)
		return model.Geolocation{IP: ip}, nil
	}))

	agg := New(p1, p2, p3)
	report := agg.Lookup(context.Background(), ip)

	// Order should match provider order, not completion order
	if report.Results[0].Provider != "first" {
		t.Errorf("Results[0].Provider = %q, want 'first'", report.Results[0].Provider)
	}
	if report.Results[1].Provider != "second" {
		t.Errorf("Results[1].Provider = %q, want 'second'", report.Results[1].Provider)
	}
	if report.Results[2].Provider != "third" {
		t.Errorf("Results[2].Provider = %q, want 'third'", report.Results[2].Provider)
	}
}

func TestAggregator_Lookup_Duration(t *testing.T) {
	ip := model.MustParseIPAddress("8.8.8.8")

	p := provider.NewTestProvider("test", provider.CheckerFunc(func(ctx context.Context,
		ip model.IPAddress) (model.Geolocation, error) {
		time.Sleep(50 * time.Millisecond)
		return model.Geolocation{IP: ip}, nil
	}))

	agg := New(p)
	report := agg.Lookup(context.Background(), ip)

	// Checker duration should be around 50ms
	if report.Results[0].Duration < 40*time.Millisecond || report.Results[0].Duration > 100*time.Millisecond {
		t.Errorf("Checker Duration = %v, expected around 50ms", report.Results[0].Duration)
	}

	// Total duration should be similar
	if report.TotalDuration < 40*time.Millisecond || report.TotalDuration > 100*time.Millisecond {
		t.Errorf("TotalDuration = %v, expected around 50ms", report.TotalDuration)
	}
}
