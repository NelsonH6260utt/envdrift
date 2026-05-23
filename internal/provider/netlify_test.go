package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeNetlifyServer(t *testing.T, payload map[string]map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload)
	}))
}

func TestNewNetlifyProvider_Validation(t *testing.T) {
	if _, err := NewNetlifyProvider("", "site-123"); err == nil {
		t.Fatal("expected error for missing token")
	}
	if _, err := NewNetlifyProvider("tok", ""); err == nil {
		t.Fatal("expected error for missing site_id")
	}
	if _, err := NewNetlifyProvider("tok", "site-123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNetlifyProvider_Name(t *testing.T) {
	p, _ := NewNetlifyProvider("tok", "site-123")
	if p.Name() != "netlify" {
		t.Fatalf("expected netlify, got %s", p.Name())
	}
}

func TestNetlifyProvider_FetchEnv_AllKeys(t *testing.T) {
	payload := map[string]map[string]string{
		"DB_URL":  {"value": "postgres://localhost"},
		"API_KEY": {"value": "secret"},
	}
	srv := makeNetlifyServer(t, payload)
	defer srv.Close()

	p, _ := NewNetlifyProvider("tok", "site-abc")
	p.baseURL = srv.URL

	env, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["DB_URL"] != "postgres://localhost" {
		t.Errorf("expected postgres://localhost, got %s", env["DB_URL"])
	}
	if env["API_KEY"] != "secret" {
		t.Errorf("expected secret, got %s", env["API_KEY"])
	}
}

func TestNetlifyProvider_FetchEnv_FilterKeys(t *testing.T) {
	payload := map[string]map[string]string{
		"DB_URL":  {"value": "postgres://localhost"},
		"API_KEY": {"value": "secret"},
		"EXTRA":   {"value": "ignored"},
	}
	srv := makeNetlifyServer(t, payload)
	defer srv.Close()

	p, _ := NewNetlifyProvider("tok", "site-abc")
	p.baseURL = srv.URL

	env, err := p.FetchEnv(context.Background(), []string{"DB_URL", "API_KEY"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := env["EXTRA"]; ok {
		t.Error("EXTRA should have been filtered out")
	}
	if len(env) != 2 {
		t.Errorf("expected 2 keys, got %d", len(env))
	}
}

func TestNetlifyProvider_FetchEnv_BadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	p, _ := NewNetlifyProvider("bad-tok", "site-abc")
	p.baseURL = srv.URL

	if _, err := p.FetchEnv(context.Background(), nil); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}
