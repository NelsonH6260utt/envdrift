package provider

import (
	"context"
	"testing"
)

func TestNewGCPRuntimeProvider_MissingProject(t *testing.T) {
	_, err := NewGCPRuntimeProvider("", "")
	if err == nil {
		t.Fatal("expected error for empty project, got nil")
	}
}

func TestGCPRuntimeProvider_Name(t *testing.T) {
	p, err := NewGCPRuntimeProvider("my-project", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := p.Name(); got != "gcp-runtime(my-project)" {
		t.Errorf("Name() = %q, want %q", got, "gcp-runtime(my-project)")
	}
}

func TestGCPRuntimeProvider_FetchEnv_SpecificKeys(t *testing.T) {
	t.Setenv("GCPRT_FOO", "bar")
	t.Setenv("GCPRT_BAZ", "qux")

	p, _ := NewGCPRuntimeProvider("proj", "")
	got, err := p.FetchEnv(context.Background(), []string{"GCPRT_FOO", "GCPRT_MISSING"})
	if err != nil {
		t.Fatalf("FetchEnv error: %v", err)
	}
	if got["GCPRT_FOO"] != "bar" {
		t.Errorf("expected GCPRT_FOO=bar, got %q", got["GCPRT_FOO"])
	}
	if _, ok := got["GCPRT_MISSING"]; ok {
		t.Error("expected GCPRT_MISSING to be absent")
	}
	if _, ok := got["GCPRT_BAZ"]; ok {
		t.Error("expected GCPRT_BAZ to be absent when not in keys list")
	}
}

func TestGCPRuntimeProvider_FetchEnv_PrefixFilter(t *testing.T) {
	t.Setenv("APP_HOST", "localhost")
	t.Setenv("APP_PORT", "8080")
	t.Setenv("OTHER_VAR", "ignore")

	p, _ := NewGCPRuntimeProvider("proj", "APP_")
	got, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("FetchEnv error: %v", err)
	}
	if got["HOST"] != "localhost" {
		t.Errorf("expected HOST=localhost after prefix strip, got %q", got["HOST"])
	}
	if got["PORT"] != "8080" {
		t.Errorf("expected PORT=8080 after prefix strip, got %q", got["PORT"])
	}
	if _, ok := got["OTHER_VAR"]; ok {
		t.Error("OTHER_VAR should be excluded by prefix filter")
	}
}

func TestGCPRuntimeProvider_FetchEnv_NoPrefix(t *testing.T) {
	t.Setenv("ENVDRIFT_UNIQUE_KEY", "unique_value")

	p, _ := NewGCPRuntimeProvider("proj", "")
	got, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("FetchEnv error: %v", err)
	}
	if got["ENVDRIFT_UNIQUE_KEY"] != "unique_value" {
		t.Errorf("expected ENVDRIFT_UNIQUE_KEY=unique_value, got %q", got["ENVDRIFT_UNIQUE_KEY"])
	}
}
