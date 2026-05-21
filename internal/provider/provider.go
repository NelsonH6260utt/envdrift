package provider

// Provider defines the interface for fetching environment variables
// from a cloud provider or runtime environment.
type Provider interface {
	// Name returns a human-readable identifier for the provider.
	Name() string
	// FetchEnv retrieves environment variables, optionally filtered by keys.
	// If keys is empty, all available variables are returned.
	FetchEnv(keys []string) (map[string]string, error)
}

// ErrNotFound is returned when a requested key does not exist in the provider.
type ErrNotFound struct {
	Key string
}

func (e *ErrNotFound) Error() string {
	return "provider: key not found: " + e.Key
}

// ErrUnauthorized is returned when the provider rejects credentials.
type ErrUnauthorized struct {
	Provider string
}

func (e *ErrUnauthorized) Error() string {
	return "provider: unauthorized access to " + e.Provider
}
