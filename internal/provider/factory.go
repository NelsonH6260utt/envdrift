package provider

import (
	"fmt"
	"os"
)

// Config holds the parameters needed to construct any supported provider.
type Config struct {
	Name string

	// AWS
	AWSPrefix string

	// GCP Secret Manager
	GCPProject string
	GCPPrefix  string

	// GCP Cloud Run runtime config
	GCPRuntimeProject string

	// Azure Key Vault
	AzureVaultURL string

	// HashiCorp Vault
	VaultAddr  string
	VaultToken string
	VaultMount string
	VaultPath  string

	// Doppler
	DopplerToken   string
	DopplerProject string
	DopplerConfig  string

	// GitHub Actions
	GitHubToken string
	GitHubOwner string
	GitHubRepo  string

	// Railway
	RailwayToken     string
	RailwayProjectID string
	RailwayServiceID string

	// Render
	RenderToken     string
	RenderServiceID string

	// Fly.io
	FlyioToken  string
	FlyioAppID  string

	// Vercel
	VercelToken     string
	VercelProjectID string
	VercelTeamID    string

	// Netlify
	NetlifyToken  string
	NetlifySiteID string
}

// New constructs the Provider identified by cfg.Name.
func New(cfg Config) (Provider, error) {
	switch cfg.Name {
	case "env":
		return NewEnvProvider(), nil

	case "aws":
		if cfg.AWSPrefix == "" {
			return nil, fmt.Errorf("provider aws: aws_prefix is required")
		}
		return NewAWSProvider(cfg.AWSPrefix), nil

	case "gcp":
		if cfg.GCPProject == "" {
			return nil, fmt.Errorf("provider gcp: gcp_project is required")
		}
		return NewGCPProvider(cfg.GCPProject, cfg.GCPPrefix)

	case "gcp-runtime":
		if cfg.GCPRuntimeProject == "" {
			return nil, fmt.Errorf("provider gcp-runtime: gcp_project is required")
		}
		return NewGCPRuntimeProvider(cfg.GCPRuntimeProject)

	case "azure":
		if cfg.AzureVaultURL == "" {
			return nil, fmt.Errorf("provider azure: azure_vault_url is required")
		}
		return NewAzureProvider(cfg.AzureVaultURL)

	case "vault":
		return NewVaultProvider(cfg.VaultAddr, cfg.VaultToken, cfg.VaultMount, cfg.VaultPath)

	case "doppler":
		return NewDopplerProvider(cfg.DopplerToken, cfg.DopplerProject, cfg.DopplerConfig)

	case "github":
		return NewGitHubProvider(cfg.GitHubToken, cfg.GitHubOwner, cfg.GitHubRepo)

	case "railway":
		return NewRailwayProvider(cfg.RailwayToken, cfg.RailwayProjectID, cfg.RailwayServiceID)

	case "render":
		return NewRenderProvider(cfg.RenderToken, cfg.RenderServiceID)

	case "flyio":
		return NewFlyioProvider(cfg.FlyioToken, cfg.FlyioAppID)

	case "vercel":
		return NewVercelProvider(cfg.VercelToken, cfg.VercelProjectID, cfg.VercelTeamID)

	case "netlify":
		token := cfg.NetlifyToken
		if token == "" {
			token = os.Getenv("NETLIFY_TOKEN")
		}
		siteID := cfg.NetlifySiteID
		if siteID == "" {
			siteID = os.Getenv("NETLIFY_SITE_ID")
		}
		return NewNetlifyProvider(token, siteID)

	default:
		return nil, fmt.Errorf("provider %q is not supported", cfg.Name)
	}
}
