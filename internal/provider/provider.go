package provider

import "context"

// Provider is the interface that cloud/environment adapters must implement.
type Provider interface {
	// Name returns a human-readable identifier for this provider.
	Name() string

	// FetchEnv retrieves environment variables from the underlying source.
	// If keys is non-empty, only those keys are fetched.
	// If keys is empty, all available variables are returned.
	FetchEnv(ctx context.Context, keys []string) (map[string]string, error)
}
