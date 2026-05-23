package provider

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the parameters needed to construct any supported provider.
type Config struct {
	// Provider selects the backend: "env", "aws", "gcp", "gcp-runtime",
	// "azure", "vault", "doppler", "github", "railway".
	Provider string

	// Shared / generic
	Prefix string
	Keys   []string

	// AWS
	AWSRegion string

	// GCP
	GCPProject string

	// Azure
	AzureVaultURL string

	// Vault
	VaultAddr  string
	VaultToken string
	VaultMount string
	VaultPath  string

	// Doppler
	DopplerToken   string
	DopplerProject string
	DopplerConfig  string

	// GitHub
	GitHubToken string
	GitHubOwner string
	GitHubRepo  string

	// Railway
	RailwayToken         string
	RailwayProjectID     string
	RailwayEnvironmentID string
}

// New constructs the Provider described by cfg.
func New(cfg Config) (Provider, error) {
	switch strings.ToLower(cfg.Provider) {
	case "env":
		return NewEnvProvider(cfg.Keys), nil

	case "aws":
		if cfg.Prefix == "" {
			return nil, fmt.Errorf("factory: aws provider requires a prefix")
		}
		region := cfg.AWSRegion
		if region == "" {
			region = os.Getenv("AWS_REGION")
		}
		return NewAWSProvider(region, cfg.Prefix), nil

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
			return nil, fmt.Errorf("factory: azure provider requires a vault URL")
		}
		return NewAzureProvider(cfg.AzureVaultURL, cfg.Prefix)

	case "vault":
		return NewVaultProvider(cfg.VaultAddr, cfg.VaultToken, cfg.VaultMount, cfg.VaultPath)

	case "doppler":
		return NewDopplerProvider(cfg.DopplerToken, cfg.DopplerProject, cfg.DopplerConfig)

	case "github":
		return NewGitHubProvider(cfg.GitHubToken, cfg.GitHubOwner, cfg.GitHubRepo)

	case "railway":
		return NewRailwayProvider(cfg.RailwayToken, cfg.RailwayProjectID, cfg.RailwayEnvironmentID)

	default:
		return nil, fmt.Errorf("factory: unknown provider %q", cfg.Provider)
	}
}
