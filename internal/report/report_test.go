package report_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/envdrift/internal/drift"
	"github.com/envdrift/internal/report"
)

var sampleResults = []drift.DriftResult{
	{Key: "DB_HOST", Status: drift.StatusMatch, Expected: "localhost", Got: "localhost"},
	{Key: "DB_PASS", Status: drift.StatusMismatch, Expected: "secret", Got: "wrong"},
	{Key: "API_KEY", Status: drift.StatusMissing, Expected: "abc123", Got: ""},
	{Key: "EXTRA_VAR", Status: drift.StatusExtra, Expected: "", Got: "surprise"},
}

func TestComputeSummary(t *testing.T) {
	s := report.ComputeSummary(sampleResults)
	if s.Total != 4 {
		t.Errorf("expected Total=4, got %d", s.Total)
	}
	if s.Match != 1 {
		t.Errorf("expected Match=1, got %d", s.Match)
	}
	if s.Missing != 1 {
		t.Errorf("expected Missing=1, got %d", s.Missing)
	}
	if s.Extra != 1 {
		t.Errorf("expected Extra=1, got %d", s.Extra)
	}
	if s.Mismatch != 1 {
		t.Errorf("expected Mismatch=1, got %d", s.Mismatch)
	}
}

func TestWriteText_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	s := report.ComputeSummary(nil)
	if err := report.WriteText(&buf, nil, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "No drift detected") {
		t.Errorf("expected no-drift message, got: %s", buf.String())
	}
}

func TestWriteText_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	s := report.ComputeSummary(sampleResults)
	if err := report.WriteText(&buf, sampleResults, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"[MISSING]", "[EXTRA]", "[MISMATCH]", "Drift Report"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output", want)
		}
	}
}

func TestWriteJSON_Structure(t *testing.T) {
	var buf bytes.Buffer
	s := report.ComputeSummary(sampleResults)
	if err := report.WriteJSON(&buf, sampleResults, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := out["summary"]; !ok {
		t.Error("expected 'summary' key in JSON output")
	}
	if _, ok := out["results"]; !ok {
		t.Error("expected 'results' key in JSON output")
	}
}
