package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeHerokuServer(t *testing.T, vars map[string]string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") == "" {
			t.Error("expected Accept header")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(vars)
	}))
}

func TestNewHerokuProvider_Validation(t *testing.T) {
	if _, err := NewHerokuProvider("", "token"); err == nil {
		t.Error("expected error for missing appID")
	}
	if _, err := NewHerokuProvider("myapp", ""); err == nil {
		t.Error("expected error for missing token")
	}
	if _, err := NewHerokuProvider("myapp", "token"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHerokuProvider_Name(t *testing.T) {
	p, _ := NewHerokuProvider("myapp", "tok")
	if p.Name() != "heroku" {
		t.Errorf("expected heroku, got %s", p.Name())
	}
}

func TestHerokuProvider_FetchEnv_AllKeys(t *testing.T) {
	vars := map[string]string{"DATABASE_URL": "postgres://", "PORT": "5000"}
	srv := makeHerokuServer(t, vars, http.StatusOK)
	defer srv.Close()

	p, _ := NewHerokuProvider("myapp", "tok")
	p.client = srv.Client()
	herokuAPIBase = srv.URL // override for test

	got, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["DATABASE_URL"] != "postgres://" {
		t.Errorf("expected postgres://, got %s", got["DATABASE_URL"])
	}
	if len(got) != 2 {
		t.Errorf("expected 2 keys, got %d", len(got))
	}
}

func TestHerokuProvider_FetchEnv_FilterKeys(t *testing.T) {
	vars := map[string]string{"DATABASE_URL": "postgres://", "PORT": "5000", "SECRET": "abc"}
	srv := makeHerokuServer(t, vars, http.StatusOK)
	defer srv.Close()

	p, _ := NewHerokuProvider("myapp", "tok")
	p.client = srv.Client()
	herokuAPIBase = srv.URL

	got, err := p.FetchEnv(context.Background(), []string{"PORT", "SECRET"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 filtered keys, got %d", len(got))
	}
	if _, ok := got["DATABASE_URL"]; ok {
		t.Error("DATABASE_URL should have been filtered out")
	}
}

func TestHerokuProvider_FetchEnv_ErrorStatus(t *testing.T) {
	srv := makeHerokuServer(t, nil, http.StatusUnauthorized)
	defer srv.Close()

	p, _ := NewHerokuProvider("myapp", "bad-token")
	p.client = srv.Client()
	herokuAPIBase = srv.URL

	if _, err := p.FetchEnv(context.Background(), nil); err == nil {
		t.Error("expected error for non-200 status")
	}
}
