package model

import (
	"encoding/json"
	"testing"
)

func TestParseIPAddress(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantV4  bool
		wantV6  bool
		wantStr string
	}{
		// Valid IPv4
		{
			name:    "valid IPv4",
			input:   "192.168.1.1",
			wantErr: false,
			wantV4:  true,
			wantV6:  false,
			wantStr: "192.168.1.1",
		},
		{
			name:    "valid IPv4 loopback",
			input:   "127.0.0.1",
			wantErr: false,
			wantV4:  true,
			wantV6:  false,
			wantStr: "127.0.0.1",
		},
		{
			name:    "valid IPv4 all zeros",
			input:   "0.0.0.0",
			wantErr: false,
			wantV4:  true,
			wantV6:  false,
			wantStr: "0.0.0.0",
		},
		{
			name:    "valid IPv4 broadcast",
			input:   "255.255.255.255",
			wantErr: false,
			wantV4:  true,
			wantV6:  false,
			wantStr: "255.255.255.255",
		},
		// Valid IPv6
		{
			name:    "valid IPv6 full",
			input:   "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			wantErr: false,
			wantV4:  false,
			wantV6:  true,
			wantStr: "2001:db8:85a3::8a2e:370:7334", // netip normalises
		},
		{
			name:    "valid IPv6 compressed",
			input:   "2001:db8::1",
			wantErr: false,
			wantV4:  false,
			wantV6:  true,
			wantStr: "2001:db8::1",
		},
		{
			name:    "valid IPv6 loopback",
			input:   "::1",
			wantErr: false,
			wantV4:  false,
			wantV6:  true,
			wantStr: "::1",
		},
		{
			name:    "valid IPv6 unspecified",
			input:   "::",
			wantErr: false,
			wantV4:  false,
			wantV6:  true,
			wantStr: "::",
		},
		// Invalid inputs
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "not-an-ip",
			wantErr: true,
		},
		{
			name:    "IPv4 with port",
			input:   "192.168.1.1:8080",
			wantErr: true,
		},
		{
			name:    "IPv4 octet out of range",
			input:   "256.1.1.1",
			wantErr: true,
		},
		{
			name:    "IPv4 too many octets",
			input:   "192.168.1.1.1",
			wantErr: true,
		},
		{
			name:    "IPv4 too few octets",
			input:   "192.168.1",
			wantErr: true,
		},
		{
			name:    "IPv6 with zone",
			input:   "fe80::1%eth0",
			wantErr: true,
		},
		{
			name:    "CIDR notation",
			input:   "192.168.1.0/24",
			wantErr: true,
		},
		{
			name:    "whitespace",
			input:   " 192.168.1.1 ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, err := ParseIPAddress(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseIPAddress(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseIPAddress(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got := ip.IsV4(); got != tt.wantV4 {
				t.Errorf("IsV4() = %v, want %v", got, tt.wantV4)
			}

			if got := ip.IsV6(); got != tt.wantV6 {
				t.Errorf("IsV6() = %v, want %v", got, tt.wantV6)
			}

			if got := ip.String(); got != tt.wantStr {
				t.Errorf("String() = %q, want %q", got, tt.wantStr)
			}

			if !ip.IsValid() {
				t.Error("IsValid() = false, want true")
			}
		})
	}
}

func TestIPAddress_ZeroValue(t *testing.T) {
	var ip IPAddress

	if ip.IsValid() {
		t.Error("zero value should not be valid")
	}

	if ip.String() != "" {
		t.Errorf("zero value String() = %q, want empty string", ip.String())
	}

	if ip.IsV4() {
		t.Error("zero value should not be IPv4")
	}

	if ip.IsV6() {
		t.Error("zero value should not be IPv6")
	}
}

func TestIPAddress_Properties(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		isLoopback      bool
		isPrivate       bool
		isGlobalUnicast bool
		isUnspecified   bool
	}{
		{
			name:            "IPv4 loopback",
			input:           "127.0.0.1",
			isLoopback:      true,
			isPrivate:       false,
			isGlobalUnicast: false,
			isUnspecified:   false,
		},
		{
			name:            "IPv6 loopback",
			input:           "::1",
			isLoopback:      true,
			isPrivate:       false,
			isGlobalUnicast: false,
			isUnspecified:   false,
		},
		{
			name:            "IPv4 private class A",
			input:           "10.0.0.1",
			isLoopback:      false,
			isPrivate:       true,
			isGlobalUnicast: true,
			isUnspecified:   false,
		},
		{
			name:            "IPv4 private class B",
			input:           "172.16.0.1",
			isLoopback:      false,
			isPrivate:       true,
			isGlobalUnicast: true,
			isUnspecified:   false,
		},
		{
			name:            "IPv4 private class C",
			input:           "192.168.1.1",
			isLoopback:      false,
			isPrivate:       true,
			isGlobalUnicast: true,
			isUnspecified:   false,
		},
		{
			name:            "IPv4 public",
			input:           "8.8.8.8",
			isLoopback:      false,
			isPrivate:       false,
			isGlobalUnicast: true,
			isUnspecified:   false,
		},
		{
			name:            "IPv4 unspecified",
			input:           "0.0.0.0",
			isLoopback:      false,
			isPrivate:       false,
			isGlobalUnicast: false,
			isUnspecified:   true,
		},
		{
			name:            "IPv6 unspecified",
			input:           "::",
			isLoopback:      false,
			isPrivate:       false,
			isGlobalUnicast: false,
			isUnspecified:   true,
		},
		{
			name:            "IPv6 global unicast",
			input:           "2001:4860:4860::8888",
			isLoopback:      false,
			isPrivate:       false,
			isGlobalUnicast: true,
			isUnspecified:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := MustParseIPAddress(tt.input)

			if got := ip.IsLoopback(); got != tt.isLoopback {
				t.Errorf("IsLoopback() = %v, want %v", got, tt.isLoopback)
			}

			if got := ip.IsPrivate(); got != tt.isPrivate {
				t.Errorf("IsPrivate() = %v, want %v", got, tt.isPrivate)
			}

			if got := ip.IsGlobalUnicast(); got != tt.isGlobalUnicast {
				t.Errorf("IsGlobalUnicast() = %v, want %v", got, tt.isGlobalUnicast)
			}

			if got := ip.IsUnspecified(); got != tt.isUnspecified {
				t.Errorf("IsUnspecified() = %v, want %v", got, tt.isUnspecified)
			}
		})
	}
}

func TestIPAddress_Compare(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want int
	}{
		{
			name: "equal IPv4",
			a:    "192.168.1.1",
			b:    "192.168.1.1",
			want: 0,
		},
		{
			name: "a less than b IPv4",
			a:    "192.168.1.1",
			b:    "192.168.1.2",
			want: -1,
		},
		{
			name: "a greater than b IPv4",
			a:    "192.168.1.2",
			b:    "192.168.1.1",
			want: 1,
		},
		{
			name: "IPv4 less than IPv6",
			a:    "192.168.1.1",
			b:    "::1",
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := MustParseIPAddress(tt.a)
			b := MustParseIPAddress(tt.b)

			if got := a.Compare(b); got != tt.want {
				t.Errorf("Compare() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIPAddress_Equal(t *testing.T) {
	a := MustParseIPAddress("192.168.1.1")
	b := MustParseIPAddress("192.168.1.1")
	c := MustParseIPAddress("192.168.1.2")

	if !a.Equal(b) {
		t.Error("Equal() should return true for same address")
	}

	if a.Equal(c) {
		t.Error("Equal() should return false for different addresses")
	}
}

func TestIPAddress_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		ip       IPAddress
		wantJSON string
	}{
		{
			name:     "valid IPv4",
			ip:       MustParseIPAddress("192.168.1.1"),
			wantJSON: `"192.168.1.1"`,
		},
		{
			name:     "valid IPv6",
			ip:       MustParseIPAddress("2001:db8::1"),
			wantJSON: `"2001:db8::1"`,
		},
		{
			name:     "zero value",
			ip:       IPAddress{},
			wantJSON: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.ip)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			if string(got) != tt.wantJSON {
				t.Errorf("Marshal() = %s, want %s", got, tt.wantJSON)
			}
		})
	}
}

func TestIPAddress_JSONUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantIP  string
		wantErr bool
	}{
		{
			name:   "valid IPv4",
			json:   `"192.168.1.1"`,
			wantIP: "192.168.1.1",
		},
		{
			name:   "valid IPv6",
			json:   `"2001:db8::1"`,
			wantIP: "2001:db8::1",
		},
		{
			name:   "null",
			json:   `null`,
			wantIP: "",
		},
		{
			name:    "invalid IP",
			json:    `"not-an-ip"`,
			wantErr: true,
		},
		{
			name:    "number instead of string",
			json:    `12345`,
			wantErr: true,
		},
		{
			name:    "empty string",
			json:    `""`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ip IPAddress
			err := json.Unmarshal([]byte(tt.json), &ip)

			if tt.wantErr {
				if err == nil {
					t.Error("Unmarshal() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unmarshal() unexpected error: %v", err)
			}

			if ip.String() != tt.wantIP {
				t.Errorf("Unmarshal() ip = %q, want %q", ip.String(), tt.wantIP)
			}
		})
	}
}

func TestIPAddress_JSONRoundTrip(t *testing.T) {
	tests := []string{
		"192.168.1.1",
		"8.8.8.8",
		"127.0.0.1",
		"::1",
		"2001:db8::1",
		"2001:4860:4860::8888",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			original := MustParseIPAddress(input)

			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			var decoded IPAddress
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if !original.Equal(decoded) {
				t.Errorf("round trip failed: original = %v, decoded = %v", original, decoded)
			}
		})
	}
}

func TestIPAddress_JSONInStruct(t *testing.T) {
	type wrapper struct {
		Address IPAddress `json:"address"`
		Name    string    `json:"Name"`
	}

	original := wrapper{
		Address: MustParseIPAddress("8.8.8.8"),
		Name:    "Google DNS",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	wantJSON := `{"address":"8.8.8.8","Name":"Google DNS"}`
	if string(data) != wantJSON {
		t.Errorf("Marshal() = %s, want %s", data, wantJSON)
	}

	var decoded wrapper
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if !original.Address.Equal(decoded.Address) {
		t.Errorf("Address mismatch: got %v, want %v", decoded.Address, original.Address)
	}

	if original.Name != decoded.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, original.Name)
	}
}

func TestIPAddress_TextMarshal(t *testing.T) {
	ip := MustParseIPAddress("192.168.1.1")

	data, err := ip.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error = %v", err)
	}

	if string(data) != "192.168.1.1" {
		t.Errorf("MarshalText() = %s, want 192.168.1.1", data)
	}
}

func TestIPAddress_TextUnmarshal(t *testing.T) {
	var ip IPAddress

	if err := ip.UnmarshalText([]byte("192.168.1.1")); err != nil {
		t.Fatalf("UnmarshalText() error = %v", err)
	}

	if ip.String() != "192.168.1.1" {
		t.Errorf("UnmarshalText() ip = %s, want 192.168.1.1", ip.String())
	}

	// Empty should result in zero value
	var ip2 IPAddress
	if err := ip2.UnmarshalText([]byte("")); err != nil {
		t.Fatalf("UnmarshalText() error = %v", err)
	}

	if ip2.IsValid() {
		t.Error("UnmarshalText() with empty should result in zero value")
	}
}

func TestMustParseIPAddress_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParseIPAddress() should panic on invalid input")
		}
	}()

	MustParseIPAddress("not-an-ip")
}
