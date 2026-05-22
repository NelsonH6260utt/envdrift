package provider

import (
	"context"
	"testing"
)

func TestNewVaultProvider_MissingAddr(t *testing.T) {
	_, err := NewVaultProvider("", "token", "secret", "myapp", "")
	if err == nil {
		t.Fatal("expected error for missing addr")
	}
}

func TestNewVaultProvider_MissingMount(t *testing.T) {
	_, err := NewVaultProvider("http://127.0.0.1:8200", "token", "", "myapp", "")
	if err == nil {
		t.Fatal("expected error for missing mount")
	}
}

func TestNewVaultProvider_MissingPath(t *testing.T) {
	_, err := NewVaultProvider("http://127.0.0.1:8200", "token", "secret", "", "")
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestVaultProvider_Name(t *testing.T) {
	p, err := NewVaultProvider("http://127.0.0.1:8200", "token", "secret", "myapp", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := p.Name(); got != "vault" {
		t.Errorf("Name() = %q, want %q", got, "vault")
	}
}

// fakeKV simulates the data returned by a Vault KV secret for unit testing.
type fakeVaultProvider struct {
	data   map[string]interface{}
	prefix string
}

func (f *fakeVaultProvider) fetchFiltered(keys []string) (map[string]string, error) {
	wantKeys := make(map[string]bool, len(keys))
	for _, k := range keys {
		wantKeys[k] = true
	}

	result := make(map[string]string)
	for k, val := range f.data {
		effective := k
		if f.prefix != "" {
			if len(k) <= len(f.prefix) || k[:len(f.prefix)] != f.prefix {
				continue
			}
			effective = k[len(f.prefix):]
		}
		if len(wantKeys) > 0 && !wantKeys[effective] {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		result[effective] = str
	}
	return result, nil
}

func TestVaultFetchFiltered_SpecificKeys(t *testing.T) {
	f := &fakeVaultProvider{
		data: map[string]interface{}{
			"DB_HOST": "localhost",
			"DB_PORT": "5432",
			"SECRET":  "s3cr3t",
		},
	}
	got, err := f.fetchFiltered([]string{"DB_HOST"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got["DB_HOST"] != "localhost" {
		t.Errorf("unexpected result: %v", got)
	}
}

func TestVaultFetchFiltered_PrefixStripped(t *testing.T) {
	f := &fakeVaultProvider{
		prefix: "APP_",
		data: map[string]interface{}{
			"APP_HOST": "prod.example.com",
			"OTHER":    "ignored",
		},
	}
	got, err := f.fetchFiltered(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["HOST"] != "prod.example.com" {
		t.Errorf("expected HOST=prod.example.com, got %v", got)
	}
	if _, ok := got["OTHER"]; ok {
		t.Error("OTHER should have been filtered out by prefix")
	}
}

func TestVaultProvider_FetchEnv_ReturnsErrorOnBadAddr(t *testing.T) {
	p, err := NewVaultProvider("http://127.0.0.1:19999", "", "secret", "noexist", "")
	if err != nil {
		t.Fatalf("unexpected construction error: %v", err)
	}
	_, fetchErr := p.FetchEnv(context.Background(), nil)
	if fetchErr == nil {
		t.Error("expected error fetching from unreachable Vault, got nil")
	}
}
