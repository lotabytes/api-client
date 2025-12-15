// Package main is the entry point for the ipintel CLI application.
package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"api-client/internal/aggregator"
	"api-client/internal/cli"
	"api-client/internal/model"
	"api-client/internal/provider"
	"api-client/internal/provider/ipapi"
	"api-client/internal/provider/ipinfo"
	"api-client/internal/provider/ipwhois"
)

// Version is set at build time via -ldflags.
var Version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	parser := cli.NewParser()

	cfg, err := parser.Parse(args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if cfg.ShowHelp {
		parser.PrintUsage()
		return 0
	}

	if cfg.ShowVersion {
		parser.PrintVersion(Version)
		return 0
	}

	if cfg.IPAddress == "-" {
		scanner := bufio.NewScanner(os.Stdin)
		if !scanner.Scan() {
			if scanErr := scanner.Err(); scanErr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", scanErr)
			} else {
				_, _ = fmt.Fprintf(os.Stderr, "Error: no input provided on stdin\n")
			}
			return 1
		}
		cfg.IPAddress = strings.TrimSpace(scanner.Text())
		// Force JSON output for stdin mode as per requirement
		cfg.Format = cli.FormatJSON
	}

	if err := cfg.Validate(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		_, _ = fmt.Fprintf(os.Stderr, "Use --help for usage information.\n")
		return 1
	}

	// Parse and validate the IP address
	ip, err := model.ParseAddr(cfg.IPAddress)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Warn if IP is not globally routable
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: %s is not a globally routable address. Results may be limited.\n\n", ip)
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.Timeout) * time.Second,
	}

	// Create providers
	providers := []provider.Provider{
		ipapi.New(httpClient),
		ipinfo.New(httpClient),
		ipwhois.New(httpClient),
	}

	// Create aggregator and perform lookup
	agg := aggregator.New(providers...)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
	defer cancel()

	report := agg.Lookup(ctx, ip)

	// Format and output the report
	formatter := cli.NewFormatter(os.Stdout)
	if err := formatter.Format(report, cfg.Format); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		return 1
	}

	// Return non-zero if all checkers failed
	if report.SuccessCount() == 0 && len(report.Results) > 0 {
		return 1
	}

	return 0
}
