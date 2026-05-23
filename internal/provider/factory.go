package provider

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the parameters used to construct a Provider via New.
type Config struct {
	Kind string

	// AWS
	AWSPrefix string

	// GCP Secret Manager
	GCPProject string
	GCPPrefix  string

	// GCP Cloud Run / runtime config
	GCPRuntimeProject string

	// Azure Key Vault
	AzureVaultURL string

	// HashiCorp Vault
	VaultAddr  string
	VaultMount string
	VaultPath  string
	VaultToken string

	// Doppler
	DopplerToken   string
	DopplerProject string
	DopplerConfig  string

	// GitHub Actions
	GitHubOwner string
	GitHubRepo  string
	GitHubEnv   string
	GitHubToken string
}

// New constructs a Provider from the supplied Config.
func New(cfg Config) (Provider, error) {
	switch strings.ToLower(cfg.Kind) {
	case "env":
		return NewEnvProvider(os.Environ()), nil

	case "aws":
		if cfg.AWSPrefix == "" {
			return nil, fmt.Errorf("factory: aws provider requires AWSPrefix")
		}
		return NewAWSProvider(cfg.AWSPrefix), nil

	case "gcp":
		if cfg.GCPProject == "" {
			return nil, fmt.Errorf("factory: gcp provider requires GCPProject")
		}
		return NewGCPProvider(cfg.GCPProject, cfg.GCPPrefix)

	case "gcp-runtime":
		if cfg.GCPRuntimeProject == "" {
			return nil, fmt.Errorf("factory: gcp-runtime provider requires GCPRuntimeProject")
		}
		return NewGCPRuntimeProvider(cfg.GCPRuntimeProject)

	case "azure":
		if cfg.AzureVaultURL == "" {
			return nil, fmt.Errorf("factory: azure provider requires AzureVaultURL")
		}
		return NewAzureProvider(cfg.AzureVaultURL)

	case "vault":
		return NewVaultProvider(cfg.VaultAddr, cfg.VaultMount, cfg.VaultPath, cfg.VaultToken)

	case "doppler":
		return NewDopplerProvider(cfg.DopplerToken, cfg.DopplerProject, cfg.DopplerConfig)

	case "github":
		if cfg.GitHubOwner == "" {
			return nil, fmt.Errorf("factory: github provider requires GitHubOwner")
		}
		if cfg.GitHubRepo == "" {
			return nil, fmt.Errorf("factory: github provider requires GitHubRepo")
		}
		if cfg.GitHubToken == "" {
			return nil, fmt.Errorf("factory: github provider requires GitHubToken")
		}
		return NewGitHubProvider(cfg.GitHubOwner, cfg.GitHubRepo, cfg.GitHubEnv, cfg.GitHubToken)

	default:
		return nil, fmt.Errorf("factory: unknown provider kind %q", cfg.Kind)
	}
}
