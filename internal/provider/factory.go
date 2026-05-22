package provider

import (
	"fmt"
	"strings"
)

// Config holds the parameters needed to construct a Provider.
type Config struct {
	// Type is one of: "env", "aws", "gcp", "gcp-runtime", "azure"
	Type string

	// Keys restricts which env vars are fetched. Empty means all.
	Keys []string

	// AWS / GCP / Azure shared
	Prefix string

	// AWS-specific
	AWSRegion string

	// GCP-specific
	GCPProject string

	// Azure-specific
	AzureVaultURL string
}

// New constructs a Provider from a Config.
func New(cfg Config) (Provider, error) {
	switch strings.ToLower(cfg.Type) {
	case "env":
		return NewEnvProvider(cfg.Keys), nil

	case "aws":
		if cfg.Prefix == "" {
			return nil, fmt.Errorf("factory: aws provider requires a prefix")
		}
		return NewAWSProvider(cfg.Prefix, cfg.AWSRegion), nil

	case "gcp":
		if cfg.GCPProject == "" {
			return nil, fmt.Errorf("factory: gcp provider requires a project")
		}
		return NewGCPProvider(cfg.GCPProject, cfg.Prefix)

	case "gcp-runtime":
		if cfg.GCPProject == "" {
			return nil, fmt.Errorf("factory: gcp-runtime provider requires a project")
		}
		return NewGCPRuntimeProvider(cfg.GCPProject, cfg.Prefix)

	case "azure":
		if cfg.AzureVaultURL == "" {
			return nil, fmt.Errorf("factory: azure provider requires a vaultURL")
		}
		return NewAzureProvider(cfg.AzureVaultURL, cfg.Prefix)

	default:
		return nil, fmt.Errorf("factory: unknown provider type %q", cfg.Type)
	}
}
