package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const renderAPIBase = "https://api.render.com/v1"

type renderProvider struct {
	serviceID string
	client    *http.Client
}

type renderEnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type renderEnvVarWrapper struct {
	EnvVar renderEnvVar `json:"envVar"`
}

// NewRenderProvider creates a provider that fetches env vars from a Render service.
// Requires RENDER_API_KEY set in the environment and a valid serviceID.
func NewRenderProvider(serviceID, apiKey string) (Provider, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("render: serviceID is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("render: apiKey is required")
	}
	client := newBaseURLClient(renderAPIBase, apiKey)
	return &renderProvider{serviceID: serviceID, client: client}, nil
}

func (r *renderProvider) Name() string {
	return "render"
}

func (r *renderProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	url := fmt.Sprintf("%s/services/%s/env-vars", renderAPIBase, r.serviceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("render: build request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("render: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("render: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var wrappers []renderEnvVarWrapper
	if err := json.NewDecoder(resp.Body).Decode(&wrappers); err != nil {
		return nil, fmt.Errorf("render: decode response: %w", err)
	}

	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}

	result := make(map[string]string)
	for _, w := range wrappers {
		if len(keys) == 0 {
			result[w.EnvVar.Key] = w.EnvVar.Value
		} else if _, ok := keySet[w.EnvVar.Key]; ok {
			result[w.EnvVar.Key] = w.EnvVar.Value
		}
	}
	return result, nil
}
