package provider

import (
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/envdrift/internal/provider"
)

// Config holds the parameters needed to construct any supported provider.
type Config struct {
	Kind string // env, aws, gcp, gcp-runtime, azure, vault, doppler

	// AWS
	AWSPrefix string

	// GCP / GCP-Runtime
	GCPProject string
	GCPPrefix  string

	// Azure
	AzureVaultURL string

	// Vault
	VaultAddr  string
	VaultMount string
	VaultPath  string

	// Doppler
	DopplerToken   string
	DopplerProject string
	DopplerConfig  string
}

// New constructs a Provider from cfg.
func New(cfg Config) (provider.Provider, error) {
	switch strings.ToLower(cfg.Kind) {
	case "env":
		return NewEnvProvider(), nil

	case "aws":
		if cfg.AWSPrefix == "" {
			return nil, fmt.Errorf("factory: aws provider requires a prefix")
		}
		return NewAWSProvider(cfg.AWSPrefix), nil

	case "gcp":
		if cfg.GCPProject == "" {
			return nil, fmt.Errorf("factory: gcp provider requires a project")
		}
		return NewGCPProvider(cfg.GCPProject, cfg.GCPPrefix)

	case "gcp-runtime":
		if cfg.GCPProject == "" {
			return nil, fmt.Errorf("factory: gcp-runtime provider requires a project")
		}
		return NewGCPRuntimeProvider(cfg.GCPProject, cfg.GCPPrefix)

	case "azure":
		if cfg.AzureVaultURL == "" {
			return nil, fmt.Errorf("factory: azure provider requires a vault URL")
		}
		return NewAzureProvider(cfg.AzureVaultURL)

	case "vault":
		addr := cfg.VaultAddr
		if addr == "" {
			addr = os.Getenv("VAULT_ADDR")
		}
		if addr == "" {
			return nil, fmt.Errorf("factory: vault provider requires an address")
		}
		if cfg.VaultMount == "" {
			return nil, fmt.Errorf("factory: vault provider requires a mount")
		}
		if cfg.VaultPath == "" {
			return nil, fmt.Errorf("factory: vault provider requires a path")
		}
		return NewVaultProvider(addr, cfg.VaultMount, cfg.VaultPath)

	case "doppler":
		return NewDopplerProvider(cfg.DopplerToken, cfg.DopplerProject, cfg.DopplerConfig)

	default:
		return nil, fmt.Errorf("factory: unknown provider kind %q", cfg.Kind)
	}
}
