package provider

import (
	"context"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/iterator"
)

type GCPProvider struct {
	project string
	prefix  string
	client  gcpSecretClient
}

type gcpSecretClient interface {
	ListSecrets(ctx context.Context, req *secretmanagerpb.ListSecretsRequest) gcpSecretIterator
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

type gcpSecretIterator interface {
	Next() (*secretmanagerpb.Secret, error)
}

type realGCPClient struct {
	client *secretmanager.Client
}

func (r *realGCPClient) ListSecrets(ctx context.Context, req *secretmanagerpb.ListSecretsRequest) gcpSecretIterator {
	return r.client.ListSecrets(ctx, req)
}

func (r *realGCPClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return r.client.AccessSecretVersion(ctx, req)
}

func (r *realGCPClient) Close() error {
	return r.client.Close()
}

func NewGCPProvider(project, prefix string) (*GCPProvider, error) {
	ctx := context.Background()
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcp: failed to create client: %w", err)
	}
	return &GCPProvider{project: project, prefix: prefix, client: &realGCPClient{client: c}}, nil
}

func (g *GCPProvider) Name() string { return "gcp" }

func (g *GCPProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)

	it := g.client.ListSecrets(ctx, &secretmanagerpb.ListSecretsRequest{
		Parent: fmt.Sprintf("projects/%s", g.project),
	})

	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}

	for {
		secret, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gcp: listing secrets: %w", err)
		}

		parts := strings.Split(secret.Name, "/")
		rawName := parts[len(parts)-1]
		envKey := gcpStripPrefix(rawName, g.prefix)

		if len(keys) > 0 && !keySet[envKey] {
			continue
		}

		resp, err := g.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: fmt.Sprintf("%s/versions/latest", secret.Name),
		})
		if err != nil {
			return nil, fmt.Errorf("gcp: accessing secret %s: %w", rawName, err)
		}
		result[envKey] = string(resp.Payload.Data)
	}

	return result, nil
}

func gcpStripPrefix(name, prefix string) string {
	if prefix == "" {
		return name
	}
	return strings.TrimPrefix(name, prefix)
}
