package provider

import (
	"fmt"
	"os"
	"strings"
)

// Config holds the parameters needed to construct any supported provider.
type Config struct {
	// Provider name: "env", "aws", "gcp", "gcp-runtime", "azure", "vault", "doppler", "github", "railway", "render"
	Name string

	// Shared / generic
	Keys   []string
	Prefix string

	// AWS
	AWSRegion string

	// GCP / GCP Runtime
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
	RailwayToken     string
	RailwayProjectID string
	RailwayEnvID     string

	// Render
	RenderServiceID string
	RenderAPIKey    string
}

// New constructs the appropriate Provider from a Config.
func New(cfg Config) (Provider, error) {
	switch strings.ToLower(cfg.Name) {
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
		return NewAWSProvider(region, cfg.Prefix)

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
		return NewRailwayProvider(cfg.RailwayToken, cfg.RailwayProjectID, cfg.RailwayEnvID)

	case "render":
		apiKey := cfg.RenderAPIKey
		if apiKey == "" {
			apiKey = os.Getenv("RENDER_API_KEY")
		}
		return NewRenderProvider(cfg.RenderServiceID, apiKey)

	default:
		return nil, fmt.Errorf("factory: unknown provider %q", cfg.Name)
	}
}
