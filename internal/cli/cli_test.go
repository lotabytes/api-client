package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestParser_Parse_Defaults(t *testing.T) {
	p := NewParser()
	cfg, err := p.Parse([]string{"8.8.8.8"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.IPAddress != "8.8.8.8" {
		t.Errorf("IPAddress = %q, want '8.8.8.8'", cfg.IPAddress)
	}

	if cfg.Format != FormatText {
		t.Errorf("Format = %v, want FormatText", cfg.Format)
	}

	if cfg.Timeout != DefaultTimeout {
		t.Errorf("Timeout = %d, want %d", cfg.Timeout, DefaultTimeout)
	}

	if cfg.ShowHelp {
		t.Error("ShowHelp should be false by default")
	}

	if cfg.ShowVersion {
		t.Error("ShowVersion should be false by default")
	}
}

func TestParser_Parse_FormatText(t *testing.T) {
	tests := []struct {
		args []string
	}{
		{[]string{"-f", "text", "8.8.8.8"}},
		{[]string{"--format", "text", "8.8.8.8"}},
		{[]string{"-format", "text", "8.8.8.8"}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			p := NewParser()
			cfg, err := p.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if cfg.Format != FormatText {
				t.Errorf("Format = %v, want FormatText", cfg.Format)
			}
		})
	}
}

func TestParser_Parse_FormatJSON(t *testing.T) {
	tests := []struct {
		args []string
	}{
		{[]string{"-f", "json", "8.8.8.8"}},
		{[]string{"--format", "json", "8.8.8.8"}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			p := NewParser()
			cfg, err := p.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if cfg.Format != FormatJSON {
				t.Errorf("Format = %v, want FormatJSON", cfg.Format)
			}
		})
	}
}

func TestParser_Parse_InvalidFormat(t *testing.T) {
	p := NewParser()
	var stderr bytes.Buffer
	p.SetOutput(&bytes.Buffer{}, &stderr)

	_, err := p.Parse([]string{"-f", "xml", "8.8.8.8"})
	if err == nil {
		t.Fatal("Parse() expected error for invalid format")
	}

	if !strings.Contains(err.Error(), "invalid format") {
		t.Errorf("error = %v, should mention 'invalid format'", err)
	}
}

func TestParser_Parse_Timeout(t *testing.T) {
	tests := []struct {
		args        []string
		wantTimeout time.Duration
	}{
		{[]string{"-t", "5s", "8.8.8.8"}, MustParseDuration("5s")},
		{[]string{"--timeout", "30s", "8.8.8.8"}, MustParseDuration("30s")},
		{[]string{"-timeout", "15s", "8.8.8.8"}, MustParseDuration("15s")},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			p := NewParser()
			cfg, err := p.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if cfg.Timeout != tt.wantTimeout {
				t.Errorf("Timeout = %d, want %d", cfg.Timeout, tt.wantTimeout)
			}
		})
	}
}

func TestParser_Parse_Help(t *testing.T) {
	tests := []struct {
		args []string
	}{
		{[]string{"-h"}},
		{[]string{"--help"}},
		{[]string{"-help"}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			p := NewParser()
			cfg, err := p.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if !cfg.ShowHelp {
				t.Error("ShowHelp should be true")
			}
		})
	}
}

func TestParser_Parse_Version(t *testing.T) {
	tests := []struct {
		args []string
	}{
		{[]string{"-v"}},
		{[]string{"--version"}},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			p := NewParser()
			cfg, err := p.Parse(tt.args)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if !cfg.ShowVersion {
				t.Error("ShowVersion should be true")
			}
		})
	}
}

func TestParser_Parse_CombinedFlags(t *testing.T) {
	p := NewParser()
	cfg, err := p.Parse([]string{"-f", "json", "-t", "5s", "1.1.1.1"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.IPAddress != "1.1.1.1" {
		t.Errorf("IPAddress = %q, want '1.1.1.1'", cfg.IPAddress)
	}

	if cfg.Format != FormatJSON {
		t.Errorf("Format = %v, want FormatJSON", cfg.Format)
	}

	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout = %d, want 5", cfg.Timeout)
	}
}

func TestParser_PrintUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	p := NewParser()
	p.SetOutput(&stdout, &stderr)

	p.PrintUsage()

	output := stderr.String()

	// Check key sections are present
	if !strings.Contains(output, "USAGE:") {
		t.Error("Usage should contain 'USAGE:' section")
	}

	if !strings.Contains(output, "OPTIONS:") {
		t.Error("Usage should contain 'OPTIONS:' section")
	}

	if !strings.Contains(output, "EXAMPLES:") {
		t.Error("Usage should contain 'EXAMPLES:' section")
	}

	if !strings.Contains(output, "PROVIDERS:") {
		t.Error("Usage should contain 'PROVIDERS:' section")
	}

	// Check specific flags are documented
	if !strings.Contains(output, "--format") {
		t.Error("Usage should document --format flag")
	}

	if !strings.Contains(output, "--timeout") {
		t.Error("Usage should document --timeout flag")
	}
}

func TestParser_PrintVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	p := NewParser()
	p.SetOutput(&stdout, &stderr)

	p.PrintVersion("1.2.3")

	output := stdout.String()
	if !strings.Contains(output, "1.2.3") {
		t.Errorf("Version output should contain version number, got: %s", output)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			cfg:     Config{IPAddress: "8.8.8.8", Timeout: 10 * time.Second},
			wantErr: false,
		},
		{
			name:    "missing IP address",
			cfg:     Config{Timeout: 10},
			wantErr: true,
			errMsg:  "IP address is required",
		},
		{
			name:    "help flag skips validation",
			cfg:     Config{ShowHelp: true},
			wantErr: false,
		},
		{
			name:    "version flag skips validation",
			cfg:     Config{ShowVersion: true},
			wantErr: false,
		},
		{
			name:    "timeout too low",
			cfg:     Config{IPAddress: "8.8.8.8", Timeout: 10 * time.Millisecond},
			wantErr: true,
			errMsg:  "timeout must be at least 100 millisecond",
		},
		{
			name:    "timeout too high",
			cfg:     Config{IPAddress: "8.8.8.8", Timeout: 100 * time.Second},
			wantErr: true,
			errMsg:  "timeout must not exceed 60 seconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %v, should contain %q", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error = %v", err)
			}
		})
	}
}

func MustParseDuration(duration string) time.Duration {
	d, err := time.ParseDuration(duration)
	if err != nil {
		panic(err)
	}

	return d
}
