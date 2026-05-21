package drift

import (
	"strings"
	"testing"
)

func noIgnore() map[string]struct{} {
	return map[string]struct{}{}
}

func TestDetect_NoDrift(t *testing.T) {
	local := map[string]string{"FOO": "bar", "BAZ": "qux"}
	live := map[string]string{"FOO": "bar", "BAZ": "qux"}

	report := Detect(local, live, noIgnore())
	if report.HasDrift() {
		t.Errorf("expected no drift, got %d entries", len(report.Entries))
	}
}

func TestDetect_MissingKey(t *testing.T) {
	local := map[string]string{"FOO": "bar", "MISSING": "val"}
	live := map[string]string{"FOO": "bar"}

	report := Detect(local, live, noIgnore())
	if !report.HasDrift() {
		t.Fatal("expected drift")
	}
	if len(report.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report.Entries))
	}
	if report.Entries[0].Type != DriftMissing {
		t.Errorf("expected DriftMissing, got %s", report.Entries[0].Type)
	}
	if report.Entries[0].Key != "MISSING" {
		t.Errorf("unexpected key %q", report.Entries[0].Key)
	}
}

func TestDetect_ExtraKey(t *testing.T) {
	local := map[string]string{"FOO": "bar"}
	live := map[string]string{"FOO": "bar", "EXTRA": "surprise"}

	report := Detect(local, live, noIgnore())
	if len(report.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report.Entries))
	}
	if report.Entries[0].Type != DriftExtra {
		t.Errorf("expected DriftExtra, got %s", report.Entries[0].Type)
	}
}

func TestDetect_Mismatch(t *testing.T) {
	local := map[string]string{"DB_URL": "localhost"}
	live := map[string]string{"DB_URL": "prod.db.example.com"}

	report := Detect(local, live, noIgnore())
	if len(report.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(report.Entries))
	}
	e := report.Entries[0]
	if e.Type != DriftMismatch {
		t.Errorf("expected DriftMismatch, got %s", e.Type)
	}
	if e.LocalValue != "localhost" || e.LiveValue != "prod.db.example.com" {
		t.Errorf("unexpected values: local=%q live=%q", e.LocalValue, e.LiveValue)
	}
}

func TestDetect_IgnoreKeys(t *testing.T) {
	local := map[string]string{"SECRET": "local", "PORT": "8080"}
	live := map[string]string{"PORT": "8080"}
	ignore := map[string]struct{}{"SECRET": {}}

	report := Detect(local, live, ignore)
	if report.HasDrift() {
		t.Errorf("expected no drift after ignoring SECRET, got %+v", report.Entries)
	}
}

func TestDriftEntry_String(t *testing.T) {
	cases := []struct {
		entry    DriftEntry
		contains string
	}{
		{DriftEntry{Key: "K", Type: DriftMissing, LocalValue: "v"}, "missing"},
		{DriftEntry{Key: "K", Type: DriftExtra, LiveValue: "v"}, "extra"},
		{DriftEntry{Key: "K", Type: DriftMismatch, LocalValue: "a", LiveValue: "b"}, "mismatch"},
	}
	for _, c := range cases {
		if !strings.Contains(c.entry.String(), c.contains) {
			t.Errorf("expected %q in String() output, got %q", c.contains, c.entry.String())
		}
	}
}
