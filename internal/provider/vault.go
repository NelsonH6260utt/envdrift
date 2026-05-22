package provider

import (
	"context"
	"fmt"
	"strings"

	vaultapi "github.com/hashicorp/vault/api"
)

// VaultProvider fetches environment variables from HashiCorp Vault KV secrets.
type VaultProvider struct {
	client *vaultapi.Client
	mount  string
	path   string
	prefix string
}

// NewVaultProvider creates a new VaultProvider.
// mount is the KV mount point (e.g. "secret"), path is the secret path,
// and prefix is an optional key prefix to filter (e.g. "APP_").
func NewVaultProvider(addr, token, mount, path, prefix string) (*VaultProvider, error) {
	if addr == "" {
		return nil, fmt.Errorf("vault: address is required")
	}
	if mount == "" {
		return nil, fmt.Errorf("vault: mount is required")
	}
	if path == "" {
		return nil, fmt.Errorf("vault: secret path is required")
	}

	cfg := vaultapi.DefaultConfig()
	cfg.Address = addr

	client, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("vault: failed to create client: %w", err)
	}
	if token != "" {
		client.SetToken(token)
	}

	return &VaultProvider{
		client: client,
		mount:  mount,
		path:   path,
		prefix: prefix,
	}, nil
}

// Name returns the provider identifier.
func (v *VaultProvider) Name() string {
	return "vault"
}

// FetchEnv retrieves key/value pairs from a Vault KV v2 secret.
// If keys is non-empty, only those keys are returned.
// If prefix is set, only keys matching the prefix are included (prefix is stripped).
func (v *VaultProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	secret, err := v.client.KVv2(v.mount).Get(ctx, v.path)
	if err != nil {
		return nil, fmt.Errorf("vault: failed to read secret %s/%s: %w", v.mount, v.path, err)
	}
	if secret == nil || secret.Data == nil {
		return map[string]string{}, nil
	}

	wantKeys := make(map[string]bool, len(keys))
	for _, k := range keys {
		wantKeys[k] = true
	}

	result := make(map[string]string)
	for k, val := range secret.Data {
		if v.prefix != "" {
			if !strings.HasPrefix(k, v.prefix) {
				continue
			}
			k = strings.TrimPrefix(k, v.prefix)
		}
		if len(wantKeys) > 0 && !wantKeys[k] {
			continue
		}
		str, ok := val.(string)
		if !ok {
			continue
		}
		result[k] = str
	}
	return result, nil
}
