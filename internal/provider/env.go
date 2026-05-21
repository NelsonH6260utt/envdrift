package provider

import (
	"os"
	"strings"
)

// EnvProvider reads variables directly from the current process environment.
// It is useful for local comparisons and testing.
type EnvProvider struct{}

// NewEnvProvider returns a new EnvProvider.
func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

// Name returns the provider identifier.
func (e *EnvProvider) Name() string {
	return "env"
}

// FetchEnv returns environment variables from the current process.
// If keys is non-empty, only those keys are returned; missing keys are omitted.
// If keys is empty, all process environment variables are returned.
func (e *EnvProvider) FetchEnv(keys []string) (map[string]string, error) {
	result := make(map[string]string)

	if len(keys) == 0 {
		for _, entry := range os.Environ() {
			parts := strings.SplitN(entry, "=", 2)
			if len(parts) == 2 {
				result[parts[0]] = parts[1]
			}
		}
		return result, nil
	}

	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok {
			result[key] = val
		}
	}
	return result, nil
}
