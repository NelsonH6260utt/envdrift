package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func makeRailwayServer(vars map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var edges []map[string]interface{}
		for k, v := range vars {
			edges = append(edges, map[string]interface{}{"node": map[string]string{"name": k, "value": v}})
		}
		resp := map[string]interface{}{
			"data": map[string]interface{}{
				"variables": map[string]interface{}{
					"edges": edges,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestNewRailwayProvider_Validation(t *testing.T) {
	cases := []struct {
		token, project, env string
		wantErr             bool
	}{
		{"", "proj", "env", true},
		{"tok", "", "env", true},
		{"tok", "proj", "", true},
		{"tok", "proj", "env", false},
	}
	for _, tc := range cases {
		_, err := NewRailwayProvider(tc.token, tc.project, tc.env)
		if (err != nil) != tc.wantErr {
			t.Errorf("NewRailwayProvider(%q,%q,%q) error=%v wantErr=%v",
				tc.token, tc.project, tc.env, err, tc.wantErr)
		}
	}
}

func TestRailwayProvider_Name(t *testing.T) {
	p, _ := NewRailwayProvider("tok", "proj", "env")
	if p.Name() != "railway" {
		t.Fatalf("expected 'railway', got %q", p.Name())
	}
}

func TestRailwayProvider_FetchEnv_AllKeys(t *testing.T) {
	srv := makeRailwayServer(map[string]string{"DB_URL": "postgres://x", "PORT": "8080"})
	defer srv.Close()

	p, _ := NewRailwayProvider("tok", "proj", "env")
	p.client = srv.Client()
	railwayAPIURL = srv.URL // override for test

	got, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["DB_URL"] != "postgres://x" || got["PORT"] != "8080" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestRailwayProvider_FetchEnv_FilterKeys(t *testing.T) {
	srv := makeRailwayServer(map[string]string{"DB_URL": "postgres://x", "PORT": "8080", "SECRET": "s"})
	defer srv.Close()

	p, _ := NewRailwayProvider("tok", "proj", "env")
	p.client = srv.Client()
	railwayAPIURL = srv.URL

	got, err := p.FetchEnv(context.Background(), []string{"PORT"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["PORT"] != "8080" {
		t.Errorf("expected only PORT=8080, got %v", got)
	}
}
