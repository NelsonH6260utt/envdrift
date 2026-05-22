package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type azureProvider struct {
	vaultURL string
	prefix   string
	client   azureSecretsClient
}

type azureSecretsClient interface {
	NewListSecretPropertiesPager(opts *azsecrets.ListSecretPropertiesOptions) azureSecretsPager
	GetSecret(ctx context.Context, name, version string, opts *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}

type azureSecretsPager interface {
	More() bool
	NextPage(ctx context.Context) (azsecrets.ListSecretPropertiesResponse, error)
}

type realAzureClient struct {
	client *azsecrets.Client
}

func (r *realAzureClient) NewListSecretPropertiesPager(opts *azsecrets.ListSecretPropertiesOptions) azureSecretsPager {
	return r.client.NewListSecretPropertiesPager(opts)
}

func (r *realAzureClient) GetSecret(ctx context.Context, name, version string, opts *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error) {
	return r.client.GetSecret(ctx, name, version, opts)
}

// NewAzureProvider creates a provider backed by Azure Key Vault.
// vaultURL must be the full vault URI, e.g. https://my-vault.vault.azure.net/
func NewAzureProvider(vaultURL, prefix string) (Provider, error) {
	if vaultURL == "" {
		return nil, fmt.Errorf("azure: vaultURL is required")
	}
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("azure: credential error: %w", err)
	}
	c, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("azure: client error: %w", err)
	}
	return &azureProvider{vaultURL: vaultURL, prefix: prefix, client: &realAzureClient{client: c}}, nil
}

func (a *azureProvider) Name() string { return "azure" }

func (a *azureProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)
	pager := a.client.NewListSecretPropertiesPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("azure: list secrets: %w", err)
		}
		for _, item := range page.Value {
			if item.ID == nil {
				continue
			}
			name := item.ID.Name()
			envKey := azureNameToEnvKey(name, a.prefix)
			if envKey == "" {
				continue
			}
			if len(keys) > 0 && !containsKey(keys, envKey) {
				continue
			}
			resp, err := a.client.GetSecret(ctx, name, "", nil)
			if err != nil {
				return nil, fmt.Errorf("azure: get secret %q: %w", name, err)
			}
			if resp.Value != nil {
				result[envKey] = *resp.Value
			}
		}
	}
	return result, nil
}

// azureNameToEnvKey converts a Key Vault secret name to an env key.
// Azure secret names use hyphens; we convert them to underscores and uppercase.
func azureNameToEnvKey(name, prefix string) string {
	key := strings.ToUpper(strings.ReplaceAll(name, "-", "_"))
	if prefix != "" {
		if !strings.HasPrefix(key, prefix) {
			return ""
		}
		key = strings.TrimPrefix(key, prefix)
	}
	return key
}

func containsKey(keys []string, target string) bool {
	for _, k := range keys {
		if k == target {
			return true
		}
	}
	return false
}
