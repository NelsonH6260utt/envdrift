package provider

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/iterator"
)

type mockGCPClient struct {
	secrets []*secretmanagerpb.Secret
	values  map[string]string
	listErr error
	accessErr error
}

type mockGCPIter struct {
	secrets []*secretmanagerpb.Secret
	pos     int
}

func (m *mockGCPIter) Next() (*secretmanagerpb.Secret, error) {
	if m.pos >= len(m.secrets) {
		return nil, iterator.Done
	}
	s := m.secrets[m.pos]
	m.pos++
	return s, nil
}

func (m *mockGCPClient) ListSecrets(_ context.Context, _ *secretmanagerpb.ListSecretsRequest) gcpSecretIterator {
	return &mockGCPIter{secrets: m.secrets}
}

func (m *mockGCPClient) AccessSecretVersion(_ context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if m.accessErr != nil {
		return nil, m.accessErr
	}
	val := m.values[req.Name]
	return &secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{Data: []byte(val)},
	}, nil
}

func (m *mockGCPClient) Close() error { return nil }

func makeSecret(name string) *secretmanagerpb.Secret {
	return &secretmanagerpb.Secret{Name: "projects/my-project/secrets/" + name}
}

func TestGCPProvider_Name(t *testing.T) {
	p := &GCPProvider{project: "proj", prefix: ""}
	if p.Name() != "gcp" {
		t.Fatalf("expected gcp, got %s", p.Name())
	}
}

func TestGCPProvider_FetchEnv_AllKeys(t *testing.T) {
	mock := &mockGCPClient{
		secrets: []*secretmanagerpb.Secret{makeSecret("DB_HOST"), makeSecret("API_KEY")},
		values: map[string]string{
			"projects/my-project/secrets/DB_HOST/versions/latest": "localhost",
			"projects/my-project/secrets/API_KEY/versions/latest": "secret123",
		},
	}
	p := &GCPProvider{project: "my-project", prefix: "", client: mock}
	env, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if env["DB_HOST"] != "localhost" || env["API_KEY"] != "secret123" {
		t.Fatalf("unexpected env: %v", env)
	}
}

func TestGCPProvider_FetchEnv_FilterKeys(t *testing.T) {
	mock := &mockGCPClient{
		secrets: []*secretmanagerpb.Secret{makeSecret("DB_HOST"), makeSecret("API_KEY")},
		values: map[string]string{
			"projects/my-project/secrets/DB_HOST/versions/latest": "localhost",
		},
	}
	p := &GCPProvider{project: "my-project", prefix: "", client: mock}
	env, err := p.FetchEnv(context.Background(), []string{"DB_HOST"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := env["API_KEY"]; ok {
		t.Fatal("API_KEY should not be present")
	}
	if env["DB_HOST"] != "localhost" {
		t.Fatalf("expected localhost, got %s", env["DB_HOST"])
	}
}

func TestGCPProvider_FetchEnv_AccessError(t *testing.T) {
	mock := &mockGCPClient{
		secrets:   []*secretmanagerpb.Secret{makeSecret("DB_HOST")},
		accessErr: errors.New("permission denied"),
	}
	p := &GCPProvider{project: "my-project", prefix: "", client: mock}
	_, err := p.FetchEnv(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGCPStripPrefix(t *testing.T) {
	if gcpStripPrefix("APP_DB_HOST", "APP_") != "DB_HOST" {
		t.Fatal("prefix not stripped")
	}
	if gcpStripPrefix("DB_HOST", "") != "DB_HOST" {
		t.Fatal("empty prefix should be no-op")
	}
}
