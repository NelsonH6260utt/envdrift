package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeRenderServer(t *testing.T, vars map[string]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var wrappers []renderEnvVarWrapper
		for k, v := range vars {
			wrappers = append(wrappers, renderEnvVarWrapper{EnvVar: renderEnvVar{Key: k, Value: v}})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(wrappers)
	}))
}

func TestNewRenderProvider_Validation(t *testing.T) {
	if _, err := NewRenderProvider("", "key"); err == nil {
		t.Error("expected error for missing serviceID")
	}
	if _, err := NewRenderProvider("svc-123", ""); err == nil {
		t.Error("expected error for missing apiKey")
	}
	if _, err := NewRenderProvider("svc-123", "key"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRenderProvider_Name(t *testing.T) {
	p, _ := NewRenderProvider("svc-123", "key")
	if p.Name() != "render" {
		t.Errorf("expected render, got %s", p.Name())
	}
}

func TestRenderProvider_FetchEnv_AllKeys(t *testing.T) {
	server := makeRenderServer(t, map[string]string{
		"DB_HOST": "localhost",
		"PORT":    "8080",
	})
	defer server.Close()

	p := &renderProvider{
		serviceID: "svc-123",
		client:    newBaseURLClient(server.URL, "test-key"),
	}

	got, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["DB_HOST"] != "localhost" {
		t.Errorf("expected localhost, got %s", got["DB_HOST"])
	}
	if got["PORT"] != "8080" {
		t.Errorf("expected 8080, got %s", got["PORT"])
	}
}

func TestRenderProvider_FetchEnv_FilterKeys(t *testing.T) {
	server := makeRenderServer(t, map[string]string{
		"DB_HOST": "localhost",
		"SECRET":  "topsecret",
		"PORT":    "8080",
	})
	defer server.Close()

	p := &renderProvider{
		serviceID: "svc-123",
		client:    newBaseURLClient(server.URL, "test-key"),
	}

	got, err := p.FetchEnv(context.Background(), []string{"DB_HOST", "PORT"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 keys, got %d", len(got))
	}
	if _, ok := got["SECRET"]; ok {
		t.Error("SECRET should not be present")
	}
}

func TestRenderProvider_FetchEnv_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	p := &renderProvider{
		serviceID: "svc-123",
		client:    newBaseURLClient(server.URL, "bad-key"),
	}

	_, err := p.FetchEnv(context.Background(), nil)
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}
