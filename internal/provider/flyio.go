package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const flyioBaseURL = "https://api.fly.io"

type flyioProvider struct {
	token   string
	appName string
	client  *http.Client
}

type flyioSecret struct {
	Name      string `json:"name"`
	Digest    string `json:"digest"`
	CreatedAt string `json:"created_at"`
}

// NewFlyioProvider creates a Fly.io provider that fetches app secrets.
// token is the Fly.io API token and appName is the target application name.
func NewFlyioProvider(token, appName string) (*flyioProvider, error) {
	if token == "" {
		return nil, fmt.Errorf("flyio: token is required")
	}
	if appName == "" {
		return nil, fmt.Errorf("flyio: app name is required")
	}
	return &flyioProvider{
		token:   token,
		appName: appName,
		client:  newBaseURLClient(flyioBaseURL),
	}, nil
}

func (p *flyioProvider) Name() string {
	return "flyio"
}

func (p *flyioProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	url := fmt.Sprintf("%s/v1/apps/%s/secrets", flyioBaseURL, p.appName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("flyio: failed to build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Accept", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("flyio: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flyio: unexpected status %d", resp.StatusCode)
	}

	var secrets []flyioSecret
	if err := json.NewDecoder(resp.Body).Decode(&secrets); err != nil {
		return nil, fmt.Errorf("flyio: failed to decode response: %w", err)
	}

	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[strings.ToUpper(k)] = true
	}

	result := make(map[string]string, len(secrets))
	for _, s := range secrets {
		name := strings.ToUpper(s.Name)
		if len(keys) == 0 || keySet[name] {
			// Fly.io secrets API only returns metadata, not values.
			// We mark presence with a sentinel so drift detection can
			// identify missing keys; value comparison requires deploy-time env.
			result[name] = "<fly-secret-present>"
		}
	}
	return result, nil
}
