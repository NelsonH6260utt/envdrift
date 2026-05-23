package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const railwayAPIURL = "https://backboard.railway.app/graphql/v2"

type railwayProvider struct {
	token     string
	projectID string
	environmentID string
	client    *http.Client
}

type railwayVariable struct {
	Name  string
	Value string
}

// NewRailwayProvider creates a provider that fetches variables from Railway.
func NewRailwayProvider(token, projectID, environmentID string) (*railwayProvider, error) {
	if token == "" {
		return nil, fmt.Errorf("railway: RAILWAY_TOKEN is required")
	}
	if projectID == "" {
		return nil, fmt.Errorf("railway: RAILWAY_PROJECT_ID is required")
	}
	if environmentID == "" {
		return nil, fmt.Errorf("railway: RAILWAY_ENVIRONMENT_ID is required")
	}
	return &railwayProvider{
		token:         token,
		projectID:     projectID,
		environmentID: environmentID,
		client:        &http.Client{},
	}, nil
}

func (p *railwayProvider) Name() string { return "railway" }

func (p *railwayProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	query := fmt.Sprintf(`{"query":"{ variables(projectId: \"%s\", environmentId: \"%s\") { edges { node { name value } } } }"}`,
		p.projectID, p.environmentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, railwayAPIURL,
		strings.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("railway: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("railway: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("railway: unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Variables struct {
				Edges []struct {
					Node railwayVariable `json:"node"`
				} `json:"edges"`
			} `json:"variables"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("railway: decode response: %w", err)
	}

	wantSet := make(map[string]bool, len(keys))
	for _, k := range keys {
		wantSet[k] = true
	}

	out := make(map[string]string)
	for _, edge := range result.Data.Variables.Edges {
		n := edge.Node
		if len(keys) == 0 || wantSet[n.Name] {
			out[n.Name] = n.Value
		}
	}
	return out, nil
}
