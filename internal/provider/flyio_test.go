package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeFlyioServer(t *testing.T, secrets []flyioSecret) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(secrets)
	}))
}

func TestNewFlyioProvider_Validation(t *testing.T) {
	if _, err := NewFlyioProvider("", "myapp"); err == nil {
		t.Fatal("expected error for missing token")
	}
	if _, err := NewFlyioProvider("tok", ""); err == nil {
		t.Fatal("expected error for missing app name")
	}
	if _, err := NewFlyioProvider("tok", "myapp"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFlyioProvider_Name(t *testing.T) {
	p, _ := NewFlyioProvider("tok", "myapp")
	if p.Name() != "flyio" {
		t.Fatalf("expected 'flyio', got %q", p.Name())
	}
}

func TestFlyioProvider_FetchEnv_AllKeys(t *testing.T) {
	secrets := []flyioSecret{
		{Name: "DB_URL", Digest: "abc"},
		{Name: "API_KEY", Digest: "def"},
	}
	srv := makeFlyioServer(t, secrets)
	defer srv.Close()

	p, _ := NewFlyioProvider("mytoken", "myapp")
	p.client = newBaseURLClient(srv.URL)

	env, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(env))
	}
	if env["DB_URL"] != "<fly-secret-present>" {
		t.Errorf("expected sentinel value for DB_URL, got %q", env["DB_URL"])
	}
}

func TestFlyioProvider_FetchEnv_FilterKeys(t *testing.T) {
	secrets := []flyioSecret{
		{Name: "DB_URL", Digest: "abc"},
		{Name: "API_KEY", Digest: "def"},
		{Name: "SECRET_X", Digest: "ghi"},
	}
	srv := makeFlyioServer(t, secrets)
	defer srv.Close()

	p, _ := NewFlyioProvider("mytoken", "myapp")
	p.client = newBaseURLClient(srv.URL)

	env, err := p.FetchEnv(context.Background(), []string{"DB_URL", "API_KEY"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(env))
	}
	if _, ok := env["SECRET_X"]; ok {
		t.Error("SECRET_X should have been filtered out")
	}
}
