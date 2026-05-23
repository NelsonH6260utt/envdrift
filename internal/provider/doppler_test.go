package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewDopplerProvider_Validation(t *testing.T) {
	cases := []struct {
		name    string
		token   string
		project string
		cfg     string
		wantErr bool
	}{
		{"missing token", "", "proj", "cfg", true},
		{"missing project", "tok", "", "cfg", true},
		{"missing config", "tok", "proj", "", true},
		{"valid", "tok", "proj", "cfg", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewDopplerProvider(tc.token, tc.project, tc.cfg)
			if (err != nil) != tc.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestDopplerProvider_Name(t *testing.T) {
	p, _ := NewDopplerProvider("tok", "myproject", "production")
	if got := p.Name(); got != "doppler(myproject/production)" {
		t.Errorf("unexpected name: %s", got)
	}
}

func makeDopplerServer(t *testing.T, payload map[string]string, status int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if status == http.StatusOK {
			json.NewEncoder(w).Encode(payload)
		}
	}))
}

func TestDopplerProvider_FetchEnv_AllKeys(t *testing.T) {
	payload := map[string]string{"DB_HOST": "localhost", "API_KEY": "secret"}
	svr := makeDopplerServer(t, payload, http.StatusOK)
	defer svr.Close()

	p, _ := NewDopplerProvider("tok", "proj", "cfg")
	p.client = svr.Client()
	// point to test server by overriding via round-tripper trick
	p.client = newBaseURLClient(svr.URL)

	env, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["DB_HOST"] != "localhost" || env["API_KEY"] != "secret" {
		t.Errorf("unexpected env: %v", env)
	}
}

func TestDopplerProvider_FetchEnv_FilterKeys(t *testing.T) {
	payload := map[string]string{"DB_HOST": "localhost", "API_KEY": "secret", "PORT": "8080"}
	svr := makeDopplerServer(t, payload, http.StatusOK)
	defer svr.Close()

	p, _ := NewDopplerProvider("tok", "proj", "cfg")
	p.client = newBaseURLClient(svr.URL)

	env, err := p.FetchEnv(context.Background(), []string{"DB_HOST", "PORT"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(env) != 2 || env["DB_HOST"] != "localhost" || env["PORT"] != "8080" {
		t.Errorf("unexpected env: %v", env)
	}
}

func TestDopplerProvider_FetchEnv_HTTPError(t *testing.T) {
	svr := makeDopplerServer(t, nil, http.StatusUnauthorized)
	defer svr.Close()

	p, _ := NewDopplerProvider("bad", "proj", "cfg")
	p.client = newBaseURLClient(svr.URL)

	_, err := p.FetchEnv(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}
