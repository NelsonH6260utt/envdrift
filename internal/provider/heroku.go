package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const herokuAPIBase = "https://api.heroku.com"

type herokuProvider struct {
	appID  string
	token  string
	client *http.Client
}

// NewHerokuProvider creates a provider that fetches config vars from a Heroku app.
// Requires appID (app name or UUID) and a Heroku API token.
func NewHerokuProvider(appID, token string) (*herokuProvider, error) {
	if appID == "" {
		return nil, fmt.Errorf("heroku: app ID is required")
	}
	if token == "" {
		return nil, fmt.Errorf("heroku: API token is required")
	}
	return &herokuProvider{
		appID:  appID,
		token:  token,
		client: &http.Client{},
	}, nil
}

func (h *herokuProvider) Name() string {
	return "heroku"
}

func (h *herokuProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	url := fmt.Sprintf("%s/apps/%s/config-vars", herokuAPIBase, h.appID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("heroku: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+h.token)
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("heroku: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("heroku: unexpected status %d", resp.StatusCode)
	}

	var configVars map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&configVars); err != nil {
		return nil, fmt.Errorf("heroku: decode response: %w", err)
	}

	if len(keys) == 0 {
		return configVars, nil
	}

	filtered := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := configVars[k]; ok {
			filtered[k] = v
		}
	}
	return filtered, nil
}
