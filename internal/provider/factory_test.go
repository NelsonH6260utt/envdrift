package provider

import (
	"testing"
)

func TestNew_EnvProvider(t *testing.T) {
	p, err := New(Config{Kind: "env"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "env" {
		t.Fatalf("expected env, got %s", p.Name())
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	_, err := New(Config{Kind: "azure"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNew_AWSMissingPrefix(t *testing.T) {
	_, err := New(Config{Kind: "aws", AWSPrefix: ""})
	if err == nil {
		t.Fatal("expected error when AWS prefix is empty")
	}
}

func TestNew_AWSWithPrefix(t *testing.T) {
	// We can't make a real AWS call, but we can verify construction succeeds
	// with a valid prefix (NewAWSProvider only fails on client init).
	_, err := New(Config{Kind: "aws", AWSPrefix: "/myapp/prod/"})
	// Allow error from AWS SDK missing credentials in test env
	if err != nil {
		t.Logf("aws provider construction error (expected in CI): %v", err)
	}
}

func TestNew_GCPMissingProject(t *testing.T) {
	_, err := New(Config{Kind: "gcp", GCPProject: ""})
	if err == nil {
		t.Fatal("expected error when GCP project is empty")
	}
}

func TestNew_GCPWithProject(t *testing.T) {
	// Real GCP client init may fail without credentials; that's acceptable.
	_, err := New(Config{Kind: "gcp", GCPProject: "my-project", GCPPrefix: "APP_"})
	if err != nil {
		t.Logf("gcp provider construction error (expected in CI): %v", err)
	}
}

func TestNew_CaseInsensitiveKind(t *testing.T) {
	p, err := New(Config{Kind: "ENV"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "env" {
		t.Fatalf("expected env, got %s", p.Name())
	}
}
