package report

import (
	"encoding/json"
	"io"

	"github.com/envdrift/internal/drift"
)

// jsonReport is the top-level structure serialized as JSON output.
type jsonReport struct {
	Summary jsonSummary      `json:"summary"`
	Results []jsonDriftEntry `json:"results"`
}

type jsonSummary struct {
	Total    int `json:"total"`
	Match    int `json:"match"`
	Missing  int `json:"missing"`
	Extra    int `json:"extra"`
	Mismatch int `json:"mismatch"`
}

type jsonDriftEntry struct {
	Key      string `json:"key"`
	Status   string `json:"status"`
	Expected string `json:"expected,omitempty"`
	Got      string `json:"got,omitempty"`
}

// WriteJSON writes a machine-readable JSON drift report to w.
func WriteJSON(w io.Writer, results []drift.DriftResult, summary Summary) error {
	entries := make([]jsonDriftEntry, 0, len(results))
	for _, r := range results {
		entry := jsonDriftEntry{
			Key:    r.Key,
			Status: string(r.Status),
		}
		if r.Expected != "" {
			entry.Expected = r.Expected
		}
		if r.Got != "" {
			entry.Got = r.Got
		}
		entries = append(entries, entry)
	}

	report := jsonReport{
		Summary: jsonSummary{
			Total:    summary.Total,
			Match:    summary.Match,
			Missing:  summary.Missing,
			Extra:    summary.Extra,
			Mismatch: summary.Mismatch,
		},
		Results: entries,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}
