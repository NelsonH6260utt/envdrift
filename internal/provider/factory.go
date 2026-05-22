package provider

import (
	"context"
	"fmt"
	"strings"
)

// Provider is the interface all cloud/env providers must satisfy.
type Provider interface {
	Name() string
	FetchEnv(ctx context.Context, keys []string) (map[string]string, error)
}

// Config holds configuration for constructing a provider.
type Config struct {
	Kind      string // "env", "aws", "gcp"
	AWSPrefix string
	GCPProject string
	GCPPrefix  string
}

// New constructs a Provider from the given Config.
func New(cfg Config) (Provider, error) {
	switch strings.ToLower(cfg.Kind) {
	case "env":
		return NewEnvProvider(), nil
	case "aws":
		if cfg.AWSPrefix == "" {
			return nil, fmt.Errorf("provider: aws requires a non-empty prefix")
		}
		return NewAWSProvider(cfg.AWSPrefix)
	case "gcp":
		if cfg.GCPProject == "" {
			return nil, fmt.Errorf("provider: gcp requires a non-empty project")
		}
		return NewGCPProvider(cfg.GCPProject, cfg.GCPPrefix)
	default:
		return nil, fmt.Errorf("provider: unknown kind %q", cfg.Kind)
	}
}
