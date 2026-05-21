package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/envdrift/internal/drift"
)

// Format represents the output format for drift reports.
type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

// Summary holds aggregated statistics for a drift report.
type Summary struct {
	Total    int
	Missing  int
	Extra    int
	Mismatch int
	Match    int
}

// ComputeSummary calculates summary statistics from a slice of DriftResult.
func ComputeSummary(results []drift.DriftResult) Summary {
	s := Summary{Total: len(results)}
	for _, r := range results {
		switch r.Status {
		case drift.StatusMissing:
			s.Missing++
		case drift.StatusExtra:
			s.Extra++
		case drift.StatusMismatch:
			s.Mismatch++
		case drift.StatusMatch:
			s.Match++
		}
	}
	return s
}

// WriteText writes a human-readable drift report to w.
func WriteText(w io.Writer, results []drift.DriftResult, summary Summary) error {
	if len(results) == 0 {
		_, err := fmt.Fprintln(w, "No drift detected. Environment matches .env file.")
		return err
	}

	fmt.Fprintln(w, strings.Repeat("-", 60))
	fmt.Fprintln(w, "Drift Report")
	fmt.Fprintln(w, strings.Repeat("-", 60))

	for _, r := range results {
		switch r.Status {
		case drift.StatusMissing:
			fmt.Fprintf(w, "[MISSING]  %s (expected: %q)\n", r.Key, r.Expected)
		case drift.StatusExtra:
			fmt.Fprintf(w, "[EXTRA]    %s (got: %q)\n", r.Key, r.Got)
		case drift.StatusMismatch:
			fmt.Fprintf(w, "[MISMATCH] %s (expected: %q, got: %q)\n", r.Key, r.Expected, r.Got)
		}
	}

	fmt.Fprintln(w, strings.Repeat("-", 60))
	fmt.Fprintf(w, "Total: %d | Match: %d | Missing: %d | Extra: %d | Mismatch: %d\n",
		summary.Total, summary.Match, summary.Missing, summary.Extra, summary.Mismatch)
	return nil
}
