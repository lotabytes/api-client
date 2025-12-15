package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
)

// IPAddress represents a validated IPv4 or IPv6 address.
// It wraps netip.Addr to provide immutability, comparability,
// and consistent JSON marshalling as a string.
type IPAddress struct {
	addr netip.Addr
}

// ParseIPAddress parses a string into a validated IPAddress.
// It accepts both IPv4 and IPv6 formats.
// Returns an error if the input is empty, invalid, or contains a zone identifier.
func ParseIPAddress(s string) (IPAddress, error) {
	if s == "" {
		return IPAddress{}, errors.New("empty IP address")
	}

	addr, err := netip.ParseAddr(s)
	if err != nil {
		return IPAddress{}, fmt.Errorf("invalid IP address %q: %w", s, err)
	}

	// Reject addresses with zone identifiers (e.g., "fe80::1%eth0")
	// as they are interface-specific and not meaningful for geolocation
	if addr.Zone() != "" {
		return IPAddress{}, fmt.Errorf("IP address %q contains zone identifier", s)
	}

	return IPAddress{addr: addr}, nil
}

// MustParseIPAddress parses a string into an IPAddress, panicking on error.
// Use only for compile-time constants or tests.
func MustParseIPAddress(s string) IPAddress {
	ip, err := ParseIPAddress(s)
	if err != nil {
		panic(err)
	}
	return ip
}

// String returns the string representation of the IP address.
func (ip IPAddress) String() string {
	if !ip.addr.IsValid() {
		return ""
	}
	return ip.addr.String()
}

// IsValid reports whether the IPAddress holds a valid address.
// The zero value is not valid.
func (ip IPAddress) IsValid() bool {
	return ip.addr.IsValid()
}

// IsV4 reports whether the IP address is an IPv4 address.
func (ip IPAddress) IsV4() bool {
	return ip.addr.Is4()
}

// IsV6 reports whether the IP address is an IPv6 address.
func (ip IPAddress) IsV6() bool {
	return ip.addr.Is6()
}

// IsLoopback reports whether the IP address is a loopback address.
func (ip IPAddress) IsLoopback() bool {
	return ip.addr.IsLoopback()
}

// IsPrivate reports whether the IP address is in a private range.
func (ip IPAddress) IsPrivate() bool {
	return ip.addr.IsPrivate()
}

// IsGlobalUnicast reports whether the IP address is a global unicast address.
// This is typically what you want for geolocation lookups.
func (ip IPAddress) IsGlobalUnicast() bool {
	return ip.addr.IsGlobalUnicast()
}

// IsUnspecified reports whether the IP address is the unspecified address
// (0.0.0.0 for IPv4, :: for IPv6).
func (ip IPAddress) IsUnspecified() bool {
	return ip.addr.IsUnspecified()
}

// Compare returns -1, 0, or 1 depending on whether ip is less than,
// equal to, or greater than other.
func (ip IPAddress) Compare(other IPAddress) int {
	return ip.addr.Compare(other.addr)
}

// Equal reports whether ip and other are the same address.
func (ip IPAddress) Equal(other IPAddress) bool {
	return ip.addr == other.addr
}

// MarshalJSON implements json.Marshaler.
// The IP address is marshalled as a JSON string.
func (ip IPAddress) MarshalJSON() ([]byte, error) {
	if !ip.addr.IsValid() {
		return []byte("null"), nil
	}
	return json.Marshal(ip.addr.String())
}

// UnmarshalJSON implements json.Unmarshaler.
// The IP address is expected to be a JSON string or null.
func (ip *IPAddress) UnmarshalJSON(data []byte) error {
	// Handle null
	if string(data) == "null" {
		*ip = IPAddress{}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("IP address must be a string: %w", err)
	}

	parsed, err := ParseIPAddress(s)
	if err != nil {
		return err
	}

	*ip = parsed
	return nil
}

// MarshalText implements encoding.TextMarshaler.
func (ip IPAddress) MarshalText() ([]byte, error) {
	return []byte(ip.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (ip *IPAddress) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*ip = IPAddress{}
		return nil
	}

	parsed, err := ParseIPAddress(string(data))
	if err != nil {
		return err
	}

	*ip = parsed
	return nil
}
