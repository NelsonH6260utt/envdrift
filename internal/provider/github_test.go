package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeGitHubServer(secrets []githubSecret, status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(status)
		if status == http.StatusOK {
			_ = json.NewEncoder(w).Encode(githubSecretsResponse{Secrets: secrets})
		}
	}))
}

func TestNewGitHubProvider_Validation(t *testing.T) {
	cases := []struct {
		owner, repo, token string
		wantErr            bool
	}{
		{"", "repo", "tok", true},
		{"owner", "", "tok", true},
		{"owner", "repo", "", true},
		{"owner", "repo", "tok", false},
	}
	for _, tc := range cases {
		_, err := NewGitHubProvider(tc.owner, tc.repo, "", tc.token)
		if (err != nil) != tc.wantErr {
			t.Errorf("NewGitHubProvider(%q,%q,%q) err=%v wantErr=%v", tc.owner, tc.repo, tc.token, err, tc.wantErr)
		}
	}
}

func TestGitHubProvider_Name(t *testing.T) {
	p, _ := NewGitHubProvider("acme", "myrepo", "", "tok")
	if p.Name() != "github(acme/myrepo)" {
		t.Errorf("unexpected name: %s", p.Name())
	}
	p.env = "production"
	if p.Name() != "github(acme/myrepo@production)" {
		t.Errorf("unexpected name with env: %s", p.Name())
	}
}

func TestGitHubProvider_FetchEnv_AllKeys(t *testing.T) {
	secrets := []githubSecret{{Name: "DB_URL"}, {Name: "API_KEY"}, {Name: "SECRET"}}
	srv := makeGitHubServer(secrets, http.StatusOK)
	defer srv.Close()

	p, _ := NewGitHubProvider("o", "r", "", "tok")
	p.baseURL = srv.URL

	env, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 3 {
		t.Errorf("expected 3 secrets, got %d", len(env))
	}
}

func TestGitHubProvider_FetchEnv_FilterKeys(t *testing.T) {
	secrets := []githubSecret{{Name: "DB_URL"}, {Name: "API_KEY"}, {Name: "SECRET"}}
	srv := makeGitHubServer(secrets, http.StatusOK)
	defer srv.Close()

	p, _ := NewGitHubProvider("o", "r", "", "tok")
	p.baseURL = srv.URL

	env, err := p.FetchEnv(context.Background(), []string{"DB_URL", "SECRET"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 2 {
		t.Errorf("expected 2 secrets, got %d", len(env))
	}
	if _, ok := env["API_KEY"]; ok {
		t.Error("API_KEY should have been filtered out")
	}
}

func TestGitHubProvider_FetchEnv_HTTPError(t *testing.T) {
	srv := makeGitHubServer(nil, http.StatusForbidden)
	defer srv.Close()

	p, _ := NewGitHubProvider("o", "r", "", "tok")
	p.baseURL = srv.URL

	_, err := p.FetchEnv(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}
