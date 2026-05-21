package envparser

import (
	"os"
	"testing"
)

func TestFromEnvironment_SpecificKeys(t *testing.T) {
	os.Setenv("ENVDRIFT_TEST_KEY", "hello")
	defer os.Unsetenv("ENVDRIFT_TEST_KEY")

	got := FromEnvironment([]string{"ENVDRIFT_TEST_KEY", "MISSING_KEY"})
	if got["ENVDRIFT_TEST_KEY"] != "hello" {
		t.Errorf("expected 'hello', got %q", got["ENVDRIFT_TEST_KEY"])
	}
	if _, ok := got["MISSING_KEY"]; ok {
		t.Error("MISSING_KEY should not be present")
	}
}

func TestFromEnvironment_AllKeys(t *testing.T) {
	os.Setenv("ENVDRIFT_SAMPLE", "42")
	defer os.Unsetenv("ENVDRIFT_SAMPLE")

	got := FromEnvironment(nil)
	if got["ENVDRIFT_SAMPLE"] != "42" {
		t.Errorf("expected '42', got %q", got["ENVDRIFT_SAMPLE"])
	}
}

func TestMerge_OverrideWins(t *testing.T) {
	base := EnvMap{"A": "1", "B": "2"}
	override := EnvMap{"B": "99", "C": "3"}
	result := Merge(base, override)

	if result["A"] != "1" {
		t.Errorf("A: got %q, want '1'", result["A"])
	}
	if result["B"] != "99" {
		t.Errorf("B: got %q, want '99'", result["B"])
	}
	if result["C"] != "3" {
		t.Errorf("C: got %q, want '3'", result["C"])
	}
}

func TestMerge_DoesNotMutateBase(t *testing.T) {
	base := EnvMap{"X": "original"}
	override := EnvMap{"X": "changed"}
	Merge(base, override)
	if base["X"] != "original" {
		t.Error("Merge mutated base map")
	}
}

func TestSubset(t *testing.T) {
	e := EnvMap{"A": "1", "B": "2", "C": "3"}
	sub := e.Subset([]string{"A", "C", "D"})
	if len(sub) != 2 {
		t.Errorf("expected 2 keys, got %d", len(sub))
	}
	if sub["A"] != "1" || sub["C"] != "3" {
		t.Errorf("unexpected subset values: %v", sub)
	}
}
