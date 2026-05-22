package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// GCPRuntimeProvider fetches environment variables from a running Cloud Run
// or App Engine instance by reading them from the process environment,
// filtered by a GCP-specific naming convention (e.g. a shared prefix).
type GCPRuntimeProvider struct {
	project string
	prefix  string
}

// NewGCPRuntimeProvider creates a provider that reads GCP runtime env vars.
// project is used for identification; prefix filters which env vars to include.
func NewGCPRuntimeProvider(project, prefix string) (*GCPRuntimeProvider, error) {
	if project == "" {
		return nil, fmt.Errorf("gcp-runtime: project must not be empty")
	}
	return &GCPRuntimeProvider{project: project, prefix: prefix}, nil
}

// Name returns a human-readable identifier for this provider.
func (g *GCPRuntimeProvider) Name() string {
	return fmt.Sprintf("gcp-runtime(%s)", g.project)
}

// FetchEnv returns environment variables from the current process environment.
// If keys is non-empty, only those keys are returned.
// If prefix is set, only variables whose names start with the prefix are
// considered when fetching all keys.
func (g *GCPRuntimeProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)

	if len(keys) > 0 {
		for _, k := range keys {
			if v, ok := os.LookupEnv(k); ok {
				result[k] = v
			}
		}
		return result, nil
	}

	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		k, v := parts[0], parts[1]
		if g.prefix == "" || strings.HasPrefix(k, g.prefix) {
			key := k
			if g.prefix != "" {
				key = strings.TrimPrefix(k, g.prefix)
			}
			result[key] = v
		}
	}
	return result, nil
}
