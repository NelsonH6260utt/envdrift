package provider_test

import (
	"os"
	"testing"

	"github.com/yourorg/envdrift/internal/provider"
)

func TestEnvProvider_Name(t *testing.T) {
	p := provider.NewEnvProvider()
	if p.Name() != "env" {
		t.Errorf("expected name 'env', got %q", p.Name())
	}
}

func TestEnvProvider_FetchEnv_SpecificKeys(t *testing.T) {
	os.Setenv("ENVDRIFT_TEST_FOO", "bar")
	os.Setenv("ENVDRIFT_TEST_BAZ", "qux")
	t.Cleanup(func() {
		os.Unsetenv("ENVDRIFT_TEST_FOO")
		os.Unsetenv("ENVDRIFT_TEST_BAZ")
	})

	p := provider.NewEnvProvider()
	got, err := p.FetchEnv([]string{"ENVDRIFT_TEST_FOO", "ENVDRIFT_TEST_BAZ"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["ENVDRIFT_TEST_FOO"] != "bar" {
		t.Errorf("expected 'bar', got %q", got["ENVDRIFT_TEST_FOO"])
	}
	if got["ENVDRIFT_TEST_BAZ"] != "qux" {
		t.Errorf("expected 'qux', got %q", got["ENVDRIFT_TEST_BAZ"])
	}
}

func TestEnvProvider_FetchEnv_MissingKeyOmitted(t *testing.T) {
	p := provider.NewEnvProvider()
	got, err := p.FetchEnv([]string{"ENVDRIFT_DEFINITELY_NOT_SET_XYZ"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := got["ENVDRIFT_DEFINITELY_NOT_SET_XYZ"]; ok {
		t.Error("expected missing key to be omitted from result")
	}
}

func TestEnvProvider_FetchEnv_AllKeys(t *testing.T) {
	os.Setenv("ENVDRIFT_TEST_ALL", "yes")
	t.Cleanup(func() { os.Unsetenv("ENVDRIFT_TEST_ALL") })

	p := provider.NewEnvProvider()
	got, err := p.FetchEnv(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["ENVDRIFT_TEST_ALL"] != "yes" {
		t.Errorf("expected 'yes', got %q", got["ENVDRIFT_TEST_ALL"])
	}
	if len(got) == 0 {
		t.Error("expected at least one environment variable")
	}
}
