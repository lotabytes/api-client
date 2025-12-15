package provider

import (
	"context"

	"api-client/internal/model"
	"api-client/internal/provider/ipapi"
	"api-client/internal/provider/ipinfo"
	"api-client/internal/provider/ipwhois"
)

// Checker defines the ability to check an IP address and return geolocation data.
type Checker interface {
	Check(ctx context.Context, ip model.IPAddress) (model.Geolocation, error)
}

type CheckerFunc func(ctx context.Context, ip model.IPAddress) (model.Geolocation, error)

// Check implements the Checker interface by calling the function itself.
func (f CheckerFunc) Check(ctx context.Context, ip model.IPAddress) (model.Geolocation, error) {
	return f(ctx, ip)
}

// Provider is a Checker with a Name.
type Provider interface {
	Checker
	Name() string
}

type TestProvider struct {
	name    string
	checker Checker
}

func (tp TestProvider) Name() string {
	return tp.name
}

func (tp TestProvider) Check(ctx context.Context, ip model.IPAddress) (model.Geolocation, error) {
	return tp.checker.Check(ctx, ip)
}

func NewTestProvider(name string, checker Checker) Provider {
	return TestProvider{name: name, checker: checker}
}

var _ Checker = CheckerFunc(nil)
var _ Provider = &ipapi.Client{}
var _ Provider = &ipinfo.Client{}
var _ Provider = &ipwhois.Client{}
