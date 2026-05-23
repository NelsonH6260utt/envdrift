package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const vercelAPIBase = "https://api.vercel.com"

type vercelProvider struct {
	token     string
	projectID string
	teamID    string
	baseURL   string
	client    *http.Client
}

type vercelEnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type vercelEnvResponse struct {
	Envs []vercelEnvVar `json:"envs"`
}

// NewVercelProvider creates a provider that fetches env vars from Vercel.
func NewVercelProvider(token, projectID, teamID string) (*vercelProvider, error) {
	if token == "" {
		return nil, fmt.Errorf("vercel: token is required")
	}
	if projectID == "" {
		return nil, fmt.Errorf("vercel: project_id is required")
	}
	return &vercelProvider{
		token:     token,
		projectID: projectID,
		teamID:    teamID,
		baseURL:   vercelAPIBase,
		client:    &http.Client{},
	}, nil
}

func (p *vercelProvider) Name() string { return "vercel" }

func (p *vercelProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	url := fmt.Sprintf("%s/v9/projects/%s/env", p.baseURL, p.projectID)
	if p.teamID != "" {
		url += "?teamId=" + p.teamID
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("vercel: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("vercel: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vercel: unexpected status %d", resp.StatusCode)
	}

	var result vercelEnvResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("vercel: decode response: %w", err)
	}

	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}

	out := make(map[string]string)
	for _, env := range result.Envs {
		if len(keys) == 0 || keySet[env.Key] {
			out[env.Key] = env.Value
		}
	}
	return out, nil
}
