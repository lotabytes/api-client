// Package cli provides command-line interface parsing and output formatting.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"api-client/internal/provider"
)

// OutputFormat specifies how results should be displayed.
type OutputFormat string

const (
	FormatText     OutputFormat = "text"
	FormatJSON     OutputFormat = "json"
	DefaultTimeout              = provider.DefaultRequestTimeout
)

// Config holds the parsed command-line configuration.
type Config struct {
	IPAddress   string
	Format      OutputFormat
	Timeout     time.Duration
	ShowHelp    bool
	ShowVersion bool
}

// Parser handles command-line argument parsing.
type Parser struct {
	fs     *flag.FlagSet
	stdout io.Writer
	stderr io.Writer
}

// NewParser creates a new CLI parser.
func NewParser() *Parser {
	return &Parser{
		fs:     flag.NewFlagSet("ipintel", flag.ContinueOnError),
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

// SetOutput sets the output writers for the parser (useful for testing).
func (p *Parser) SetOutput(stdout, stderr io.Writer) {
	p.stdout = stdout
	p.stderr = stderr
	p.fs.SetOutput(stderr)
}

// Parse parses command-line arguments and returns a Config.
func (p *Parser) Parse(args []string) (Config, error) {
	var cfg Config
	var format string

	p.fs.StringVar(&format, "format", "text", "output format: text or json")
	p.fs.StringVar(&format, "f", "text", "output format: text or json (shorthand)")
	p.fs.DurationVar(&cfg.Timeout, "timeout", DefaultTimeout, "timeout API requests, specified as a duration, eg '1s'")
	p.fs.DurationVar(&cfg.Timeout, "t", DefaultTimeout, "timeout as a duration (shorthand)")
	p.fs.BoolVar(&cfg.ShowHelp, "help", false, "show help message")
	p.fs.BoolVar(&cfg.ShowHelp, "h", false, "show help message (shorthand)")
	p.fs.BoolVar(&cfg.ShowVersion, "version", false, "show version information")
	p.fs.BoolVar(&cfg.ShowVersion, "v", false, "show version (shorthand)")

	p.fs.Usage = func() {
		p.PrintUsage()
	}

	if err := p.fs.Parse(args); err != nil {
		return cfg, err
	}

	// Parse format
	switch format {
	case "text", "":
		cfg.Format = FormatText
	case "json":
		cfg.Format = FormatJSON
	default:
		return cfg, fmt.Errorf("invalid format %q: must be 'text' or 'json'", format)
	}

	// Get positional argument (IP address)
	remaining := p.fs.Args()
	if len(remaining) > 0 {
		cfg.IPAddress = remaining[0]
	}

	return cfg, nil
}

// PrintUsage prints the help message.
func (p *Parser) PrintUsage() {
	usage := `ipintel - IP Intelligence Lookup Tool

USAGE:
    ipintel [OPTIONS] <IP_ADDRESS|->

DESCRIPTION:
    Queries multiple geolocation APIs concurrently to provide comprehensive
    information about an IP address, including location, ISP, and organization.

ARGUMENTS:
    <IP_ADDRESS>    IPv4 or IPv6 address to look up (e.g., 8.8.8.8 or 2001:4860:4860::8888)
    -               Read a single IP address from standard input (forces JSON output)

OPTIONS:
    -f, --format <FORMAT>     Output format: 'text' (default) or 'json'
    -t, --timeout <DURATION>  Timeout for API requests as a duration, e.g. '1s', '500ms' (default: 10 seconds)
    -h, --help                Show this help message
    -v, --version             Show version information

EXAMPLES:
    ipintel 8.8.8.8                 Look up Google's DNS server
    ipintel 2001:4860:4860::8888    Look up IPv6 address
    ipintel -f json 1.1.1.1         Output as JSON
    ipintel --timeout 5s 8.8.8.8    Set 5 second timeout
    echo 8.8.8.8 | ipintel -        Read IP from stdin and output JSON

PROVIDERS:
    Results are aggregated from the following free geolocation APIs:
    - ip-api.com
    - ipinfo.io
    - ipwhois.app

OUTPUT:
    The tool displays consensus results (most agreed-upon values) along with
    individual provider results. When providers disagree, the majority value
    is shown. Coordinates are averaged across providers.

EXIT CODES:
    0    Success
    1    Error (invalid arguments, network failure, etc.)
`
	_, _ = fmt.Fprint(p.stderr, usage)
}

// PrintVersion prints version information.
func (p *Parser) PrintVersion(version string) {
	_, _ = fmt.Fprintf(p.stdout, "ipintel version %s\n", version)
}

// Validate checks that the config has required fields.
func (cfg Config) Validate() error {
	if cfg.ShowHelp || cfg.ShowVersion {
		return nil
	}

	if cfg.IPAddress == "" {
		return fmt.Errorf("IP address is required")
	}

	if cfg.Timeout < 100*time.Millisecond {
		return fmt.Errorf("timeout must be at least 100 milliseconds")
	}

	if cfg.Timeout > 60*time.Second {
		return fmt.Errorf("timeout must not exceed 60 seconds")
	}

	return nil
}
