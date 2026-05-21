package provider

import (
	"context"
	"fmt"
)

// Config holds configuration for building a Provider via the factory.
type Config struct {
	// Type selects the provider: "env", "aws-ssm".
	Type string

	// Keys restricts which keys are fetched. Empty means all keys.
	Keys []string

	// EnvKeys is used by the "env" provider to specify which OS env vars to read.
	EnvKeys []string

	// AWSPathPrefix is the SSM parameter path prefix for the "aws-ssm" provider.
	AWSPathPrefix string
}

// New builds a Provider from the given Config.
func New(ctx context.Context, cfg Config) (Provider, error) {
	switch cfg.Type {
	case "env":
		return NewEnvProvider(cfg.EnvKeys), nil
	case "aws-ssm":
		if cfg.AWSPathPrefix == "" {
			return nil, fmt.Errorf("factory: aws-ssm requires AWSPathPrefix")
		}
		return NewAWSProvider(ctx, cfg.AWSPathPrefix)
	default:
		return nil, fmt.Errorf("factory: unknown provider type %q", cfg.Type)
	}
}
