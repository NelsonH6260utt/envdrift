package drift

import "fmt"

// DriftType describes the kind of configuration drift detected.
type DriftType string

const (
	DriftMissing  DriftType = "missing"   // key in .env but not in deployed env
	DriftExtra    DriftType = "extra"     // key in deployed env but not in .env
	DriftMismatch DriftType = "mismatch" // key present in both but values differ
)

// DriftEntry represents a single detected drift between a local and deployed env.
type DriftEntry struct {
	Key        string
	Type       DriftType
	LocalValue string
	LiveValue  string
}

// String returns a human-readable description of the drift entry.
func (d DriftEntry) String() string {
	switch d.Type {
	case DriftMissing:
		return fmt.Sprintf("[missing] %q is in .env but not deployed (local=%q)", d.Key, d.LocalValue)
	case DriftExtra:
		return fmt.Sprintf("[extra]   %q is deployed but not in .env (live=%q)", d.Key, d.LiveValue)
	case DriftMismatch:
		return fmt.Sprintf("[mismatch] %q differs: local=%q live=%q", d.Key, d.LocalValue, d.LiveValue)
	}
	return fmt.Sprintf("[unknown] %q", d.Key)
}

// Report holds all drift entries found during a comparison.
type Report struct {
	Entries []DriftEntry
}

// HasDrift returns true when at least one drift entry exists.
func (r *Report) HasDrift() bool {
	return len(r.Entries) > 0
}

// Detect compares a local env map (from .env file) against a live env map
// (from a cloud provider). Keys present in ignoreKeys are skipped.
func Detect(local, live map[string]string, ignoreKeys map[string]struct{}) *Report {
	report := &Report{}

	for key, localVal := range local {
		if _, ignored := ignoreKeys[key]; ignored {
			continue
		}
		liveVal, exists := live[key]
		if !exists {
			report.Entries = append(report.Entries, DriftEntry{
				Key:        key,
				Type:       DriftMissing,
				LocalValue: localVal,
			})
			continue
		}
		if localVal != liveVal {
			report.Entries = append(report.Entries, DriftEntry{
				Key:        key,
				Type:       DriftMismatch,
				LocalValue: localVal,
				LiveValue:  liveVal,
			})
		}
	}

	for key, liveVal := range live {
		if _, ignored := ignoreKeys[key]; ignored {
			continue
		}
		if _, exists := local[key]; !exists {
			report.Entries = append(report.Entries, DriftEntry{
				Key:       key,
				Type:      DriftExtra,
				LiveValue: liveVal,
			})
		}
	}

	return report
}
