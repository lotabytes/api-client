package model

import (
	"encoding/json"
	"testing"
)

func TestIPAddress_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		ip       IPAddress
		wantJSON string
	}{
		{
			name:     "valid IPv4",
			ip:       MustParseAddr("192.168.1.1"),
			wantJSON: `"192.168.1.1"`,
		},
		{
			name:     "valid IPv6",
			ip:       MustParseAddr("2001:db8::1"),
			wantJSON: `"2001:db8::1"`,
		},
		{
			name:     "zero value",
			ip:       IPAddress{},
			wantJSON: `""`,
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
		wantIP  IPAddress
		wantErr bool
	}{
		{
			name:   "valid IPv4",
			json:   `"192.168.1.1"`,
			wantIP: MustParseAddr("192.168.1.1"),
		},
		{
			name:   "valid IPv6",
			json:   `"2001:db8::1"`,
			wantIP: MustParseAddr("2001:db8::1"),
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
			name:   "empty string",
			json:   `""`,
			wantIP: IPAddress{},
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

			if ip.Compare(tt.wantIP) != 0 {
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
			original := MustParseAddr(input)

			data, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			var decoded IPAddress
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Unmarshal() error = %v", err)
			}

			if original.Compare(decoded) != 0 {
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
		Address: MustParseAddr("8.8.8.8"),
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

	if original.Address.Compare(decoded.Address) != 0 {
		t.Errorf("Address mismatch: got %v, want %v", decoded.Address, original.Address)
	}

	if original.Name != decoded.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, original.Name)
	}
}
