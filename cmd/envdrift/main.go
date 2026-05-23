// Package main is the entry point for the envdrift CLI tool.
// It detects configuration drift between .env files and deployed
// environment variables across various cloud providers.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/yourorg/envdrift/internal/drift"
	"github.com/yourorg/envdrift/internal/envparser"
	"github.com/yourorg/envdrift/internal/provider"
	"github.com/yourorg/envdrift/internal/report"
)

const usage = `envdrift - detect configuration drift between .env files and cloud providers

Usage:
  envdrift [flags]

Flags:
`

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		envFile    = flag.String("env", ".env", "path to the .env file")
		provName   = flag.String("provider", "env", "cloud provider (env, aws, gcp, gcp-runtime, azure, vault, doppler, github, railway, render, flyio, vercel, netlify, heroku)")
		outputFmt  = flag.String("output", "text", "output format: text or json")
		keysFlag   = flag.String("keys", "", "comma-separated list of keys to compare (default: all keys from .env file)")
		providerOpts = flag.String("provider-opts", "", "comma-separated provider options as key=value pairs")
	)

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Parse the .env file.
	local, err := envparser.ParseFile(*envFile)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", *envFile, err)
	}

	// Determine keys to compare.
	var keys []string
	if *keysFlag != "" {
		for _, k := range strings.Split(*keysFlag, ",") {
			if k = strings.TrimSpace(k); k != "" {
				keys = append(keys, k)
			}
		}
	} else {
		for k := range local {
			keys = append(keys, k)
		}
	}

	// Build provider options map.
	opts := parseProviderOpts(*providerOpts)
	opts["provider"] = *provName

	// Construct the provider.
	p, err := provider.New(opts)
	if err != nil {
		return fmt.Errorf("initialising provider %q: %w", *provName, err)
	}

	// Fetch remote environment variables.
	ctx := context.Background()
	remote, err := p.FetchEnv(ctx, keys)
	if err != nil {
		return fmt.Errorf("fetching remote env from %s: %w", p.Name(), err)
	}

	// Run drift detection.
	results := drift.Detect(local, remote, nil)
	summary := report.ComputeSummary(results)

	// Write report.
	switch *outputFmt {
	case "json":
		if err := report.WriteJSON(os.Stdout, results, summary); err != nil {
			return fmt.Errorf("writing JSON report: %w", err)
		}
	case "text":
		report.WriteText(os.Stdout, results, summary)
	default:
		return fmt.Errorf("unknown output format %q (want: text, json)", *outputFmt)
	}

	// Exit with a non-zero code when drift is detected so the tool
	// can be used effectively in CI pipelines.
	if summary.DriftCount > 0 {
		os.Exit(2)
	}
	return nil
}

// parseProviderOpts converts a comma-separated "key=value" string into a map
// that can be passed directly to provider.New.
func parseProviderOpts(raw string) map[string]string {
	opts := make(map[string]string)
	if raw == "" {
		return opts
	}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		opts[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}
	return opts
}
