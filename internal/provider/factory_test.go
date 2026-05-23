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
		t.Errorf("expected name 'env', got %q", p.Name())
	}
}

func TestNew_UnknownProvider(t *testing.T) {
	_, err := New(Config{Kind: "unknown"})
	if err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestNew_AWSMissingPrefix(t *testing.T) {
	_, err := New(Config{Kind: "aws"})
	if err == nil {
		t.Fatal("expected error when AWSPrefix is missing")
	}
}

func TestNew_AWSWithPrefix(t *testing.T) {
	p, err := New(Config{Kind: "aws", AWSPrefix: "/myapp/"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestNew_GCPMissingProject(t *testing.T) {
	_, err := New(Config{Kind: "gcp"})
	if err == nil {
		t.Fatal("expected error when GCPProject is missing")
	}
}

func TestNew_GitHubMissingOwner(t *testing.T) {
	_, err := New(Config{Kind: "github", GitHubRepo: "repo", GitHubToken: "tok"})
	if err == nil {
		t.Fatal("expected error when GitHubOwner is missing")
	}
}

func TestNew_GitHubMissingRepo(t *testing.T) {
	_, err := New(Config{Kind: "github", GitHubOwner: "owner", GitHubToken: "tok"})
	if err == nil {
		t.Fatal("expected error when GitHubRepo is missing")
	}
}

func TestNew_GitHubMissingToken(t *testing.T) {
	_, err := New(Config{Kind: "github", GitHubOwner: "owner", GitHubRepo: "repo"})
	if err == nil {
		t.Fatal("expected error when GitHubToken is missing")
	}
}

func TestNew_GitHubValid(t *testing.T) {
	p, err := New(Config{
		Kind:        "github",
		GitHubOwner: "acme",
		GitHubRepo:  "myrepo",
		GitHubToken: "ghp_test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Name() != "github(acme/myrepo)" {
		t.Errorf("unexpected provider name: %s", p.Name())
	}
}
