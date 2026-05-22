package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

// --- fakes ---

type fakeAzurePager struct {
	pages [][]azsecrets.SecretProperties
	idx   int
}

func (f *fakeAzurePager) More() bool { return f.idx < len(f.pages) }
func (f *fakeAzurePager) NextPage(_ context.Context) (azsecrets.ListSecretPropertiesResponse, error) {
	page := f.pages[f.idx]
	f.idx++
	items := make([]*azsecrets.SecretProperties, len(page))
	for i := range page {
		items[i] = &page[i]
	}
	return azsecrets.ListSecretPropertiesResponse{
		SecretPropertiesListResult: azsecrets.SecretPropertiesListResult{Value: items},
	}, nil
}

type fakeAzureClient struct {
	secrets map[string]string
	pager   *fakeAzurePager
}

func (f *fakeAzureClient) NewListSecretPropertiesPager(_ *azsecrets.ListSecretPropertiesOptions) azureSecretsPager {
	return f.pager
}

func (f *fakeAzureClient) GetSecret(_ context.Context, name, _ string, _ *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	v, ok := f.secrets[name]
	if !ok {
		return azsecrets.GetSecretResponse{}, fmt.Errorf("secret not found: %s", name)
	}
	return azsecrets.GetSecretResponse{Secret: azsecrets.Secret{Value: &v}}, nil
}

func makeAzureSecret(name string) azsecrets.SecretProperties {
	id := azsecrets.ID("https://vault.azure.net/secrets/" + name)
	return azsecrets.SecretProperties{ID: &id}
}

func newFakeAzureProvider(secrets map[string]string, prefix string) *azureProvider {
	var props []azsecrets.SecretProperties
	for name := range secrets {
		props = append(props, makeAzureSecret(name))
	}
	return &azureProvider{
		vaultURL: "https://fake.vault.azure.net/",
		prefix:   prefix,
		client: &fakeAzureClient{
			secrets: secrets,
			pager:   &fakeAzurePager{pages: [][]azsecrets.SecretProperties{props}},
		},
	}
}

// --- tests ---

func TestAzureProvider_Name(t *testing.T) {
	p := &azureProvider{}
	if p.Name() != "azure" {
		t.Fatalf("expected azure, got %s", p.Name())
	}
}

func TestAzureNameToEnvKey_NoPrefix(t *testing.T) {
	got := azureNameToEnvKey("my-db-password", "")
	if got != "MY_DB_PASSWORD" {
		t.Fatalf("expected MY_DB_PASSWORD, got %s", got)
	}
}

func TestAzureNameToEnvKey_WithPrefix(t *testing.T) {
	got := azureNameToEnvKey("app-db-password", "APP_")
	if got != "DB_PASSWORD" {
		t.Fatalf("expected DB_PASSWORD, got %s", got)
	}
}

func TestAzureNameToEnvKey_PrefixMismatch(t *testing.T) {
	got := azureNameToEnvKey("other-secret", "APP_")
	if got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

func TestAzureProvider_FetchEnv_AllKeys(t *testing.T) {
	p := newFakeAzureProvider(map[string]string{
		"db-host": "localhost",
		"db-port": "5432",
	}, "")
	env, err := p.FetchEnv(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if env["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %q", env["DB_HOST"])
	}
	if env["DB_PORT"] != "5432" {
		t.Errorf("expected DB_PORT=5432, got %q", env["DB_PORT"])
	}
}

func TestAzureProvider_FetchEnv_SpecificKeys(t *testing.T) {
	p := newFakeAzureProvider(map[string]string{
		"db-host": "localhost",
		"db-port": "5432",
	}, "")
	env, err := p.FetchEnv(context.Background(), []string{"DB_HOST"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := env["DB_PORT"]; ok {
		t.Error("DB_PORT should not be present")
	}
	if env["DB_HOST"] != "localhost" {
		t.Errorf("expected DB_HOST=localhost, got %q", env["DB_HOST"])
	}
}
