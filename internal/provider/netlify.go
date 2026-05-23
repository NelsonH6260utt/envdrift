package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const netlifyBaseURL = "https://api.netlify.com/api/v1"

type netlifyProvider struct {
	token   string
	siteID  string
	client  *http.Client
	baseURL string
}

// NewNetlifyProvider creates a provider that fetches env vars from Netlify site environment.
func NewNetlifyProvider(token, siteID string) (*netlifyProvider, error) {
	if token == "" {
		return nil, fmt.Errorf("netlify: token is required")
	}
	if siteID == "" {
		return nil, fmt.Errorf("netlify: site_id is required")
	}
	return &netlifyProvider{
		token:   token,
		siteID:  siteID,
		client:  &http.Client{},
		baseURL: netlifyBaseURL,
	}, nil
}

func (p *netlifyProvider) Name() string { return "netlify" }

func (p *netlifyProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	url := fmt.Sprintf("%s/sites/%s/env", p.baseURL, p.siteID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("netlify: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("netlify: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("netlify: unexpected status %d", resp.StatusCode)
	}

	// Netlify returns an object: { "KEY": { "value": "...", ... }, ... }
	var raw map[string]struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("netlify: decode response: %w", err)
	}

	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}

	result := make(map[string]string)
	for k, v := range raw {
		if len(keys) == 0 || keySet[k] {
			result[k] = v.Value
		}
	}
	return result, nil
}
