package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeVercelServer(envs []vercelEnvVar) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(vercelEnvResponse{Envs: envs})
	}))
}

func TestNewVercelProvider_Validation(t *testing.T) {
	if _, err := NewVercelProvider("", "proj", ""); err == nil {
		t.Fatal("expected error for missing token")
	}
	if _, err := NewVercelProvider("tok", "", ""); err == nil {
		t.Fatal("expected error for missing project_id")
	}
	if _, err := NewVercelProvider("tok", "proj", ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVercelProvider_Name(t *testing.T) {
	p, _ := NewVercelProvider("tok", "proj", "")
	if p.Name() != "vercel" {
		t.Fatalf("expected 'vercel', got %q", p.Name())
	}
}

func TestVercelProvider_FetchEnv_AllKeys(t *testing.T) {
	envs := []vercelEnvVar{
		{Key: "APP_ENV", Value: "production"},
		{Key: "DB_URL", Value: "postgres://localhost/db"},
	}
	srv := makeVercelServer(envs)
	defer srv.Close()

	p, _ := NewVercelProvider("tok", "proj", "")
	p.baseURL = srv.URL

	got, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["APP_ENV"] != "production" {
		t.Errorf("expected APP_ENV=production, got %q", got["APP_ENV"])
	}
	if got["DB_URL"] != "postgres://localhost/db" {
		t.Errorf("expected DB_URL set, got %q", got["DB_URL"])
	}
}

func TestVercelProvider_FetchEnv_FilterKeys(t *testing.T) {
	envs := []vercelEnvVar{
		{Key: "APP_ENV", Value: "production"},
		{Key: "SECRET", Value: "s3cr3t"},
	}
	srv := makeVercelServer(envs)
	defer srv.Close()

	p, _ := NewVercelProvider("tok", "proj", "")
	p.baseURL = srv.URL

	got, err := p.FetchEnv(context.Background(), []string{"APP_ENV"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := got["SECRET"]; ok {
		t.Error("SECRET should have been filtered out")
	}
	if got["APP_ENV"] != "production" {
		t.Errorf("expected APP_ENV=production, got %q", got["APP_ENV"])
	}
}
