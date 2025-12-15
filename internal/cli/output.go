package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"api-client/internal/model"
)

// Formatter formats and outputs reports.
type Formatter struct {
	w io.Writer
}

// NewFormatter creates a new output formatter.
func NewFormatter(w io.Writer) *Formatter {
	return &Formatter{w: w}
}

// Format outputs the report in the specified format.
func (f *Formatter) Format(report model.Report, format OutputFormat) error {
	switch format {
	case FormatJSON:
		return f.formatJSON(report)
	case FormatText:
		return f.formatText(report)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func (f *Formatter) formatJSON(report model.Report) error {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func (f *Formatter) formatText(report model.Report) error {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("IP Intelligence Report for %s\n", report.IP))
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")

	// Consensus results
	consensus := report.Consensus()
	sb.WriteString("CONSENSUS (aggregated from all providers):\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	if consensus.Country != "" {
		sb.WriteString(fmt.Sprintf("  Country:      %s", consensus.Country))
		if consensus.CountryCode != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", consensus.CountryCode))
		}
		sb.WriteString("\n")
	}

	if consensus.Region != "" {
		sb.WriteString(fmt.Sprintf("  Region:       %s\n", consensus.Region))
	}

	if consensus.City != "" {
		sb.WriteString(fmt.Sprintf("  City:         %s\n", consensus.City))
	}

	if consensus.HasLocation() {
		sb.WriteString(fmt.Sprintf("  Coordinates:  %.4f, %.4f\n", consensus.Latitude, consensus.Longitude))
	}

	if consensus.ISP != "" {
		sb.WriteString(fmt.Sprintf("  ISP:          %s\n", consensus.ISP))
	}

	if consensus.Org != "" {
		sb.WriteString(fmt.Sprintf("  Organization: %s\n", consensus.Org))
	}

	if consensus.ASN != "" {
		sb.WriteString(fmt.Sprintf("  ASN:          %s\n", consensus.ASN))
	}

	sb.WriteString("\n")

	// Individual provider results
	sb.WriteString("PROVIDER DETAILS:\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	for _, result := range report.Results {
		sb.WriteString(fmt.Sprintf("\n[%s] ", result.Provider))
		if result.Success() {
			sb.WriteString(fmt.Sprintf("(%.0fms)\n", float64(result.Duration.Milliseconds())))
			f.formatGeolocation(&sb, result.Result)
		} else {
			sb.WriteString("FAILED\n")
			sb.WriteString(fmt.Sprintf("  Error: %s\n", result.Error))
		}
	}

	// Summary
	sb.WriteString("\n" + strings.Repeat("-", 40) + "\n")
	sb.WriteString(fmt.Sprintf("Total: %d/%d providers succeeded in %dms\n",
		report.SuccessCount(),
		len(report.Results),
		report.TotalDuration.Milliseconds()))

	_, err := f.w.Write([]byte(sb.String()))
	return err
}

func (f *Formatter) formatGeolocation(sb *strings.Builder, geo *model.Geolocation) {
	if geo == nil {
		return
	}

	if geo.Country != "" {
		sb.WriteString(fmt.Sprintf("  Country: %s", geo.Country))
		if geo.CountryCode != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", geo.CountryCode))
		}
		sb.WriteString("\n")
	}

	if geo.Region != "" {
		sb.WriteString(fmt.Sprintf("  Region:  %s\n", geo.Region))
	}

	if geo.City != "" {
		sb.WriteString(fmt.Sprintf("  City:    %s\n", geo.City))
	}

	if geo.HasLocation() {
		sb.WriteString(fmt.Sprintf("  Coords:  %.4f, %.4f\n", geo.Latitude, geo.Longitude))
	}

	if geo.ISP != "" {
		sb.WriteString(fmt.Sprintf("  ISP:     %s\n", geo.ISP))
	}

	if geo.Org != "" {
		sb.WriteString(fmt.Sprintf("  Org:     %s\n", geo.Org))
	}

	if geo.ASN != "" {
		sb.WriteString(fmt.Sprintf("  ASN:     %s\n", geo.ASN))
	}
}
