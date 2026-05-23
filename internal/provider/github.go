package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type githubProvider struct {
	owner  string
	repo   string
	env    string
	token  string
	client *http.Client
	baseURL string
}

type githubSecret struct {
	Name string `json:"name"`
}

type githubSecretsResponse struct {
	Secrets []githubSecret `json:"secrets"`
}

// NewGitHubProvider creates a provider that fetches secrets from GitHub Actions environment secrets.
func NewGitHubProvider(owner, repo, env, token string) (*githubProvider, error) {
	if owner == "" {
		return nil, fmt.Errorf("github: owner is required")
	}
	if repo == "" {
		return nil, fmt.Errorf("github: repo is required")
	}
	if token == "" {
		return nil, fmt.Errorf("github: token is required")
	}
	return &githubProvider{
		owner:   owner,
		repo:    repo,
		env:     env,
		token:   token,
		client:  http.DefaultClient,
		baseURL: "https://api.github.com",
	}, nil
}

func (g *githubProvider) Name() string {
	if g.env != "" {
		return fmt.Sprintf("github(%s/%s@%s)", g.owner, g.repo, g.env)
	}
	return fmt.Sprintf("github(%s/%s)", g.owner, g.repo)
}

func (g *githubProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	var url string
	if g.env != "" {
		url = fmt.Sprintf("%s/repos/%s/%s/environments/%s/secrets", g.baseURL, g.owner, g.repo, g.env)
	} else {
		url = fmt.Sprintf("%s/repos/%s/%s/actions/secrets", g.baseURL, g.owner, g.repo)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("github: create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("github: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result githubSecretsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("github: decode response: %w", err)
	}

	keySet := make(map[string]bool, len(keys))
	for _, k := range keys {
		keySet[k] = true
	}

	env := make(map[string]string)
	for _, s := range result.Secrets {
		if len(keys) == 0 || keySet[s.Name] {
			// GitHub API does not expose secret values; record presence with a sentinel.
			env[s.Name] = base64.StdEncoding.EncodeToString([]byte("<github-secret>"))
		}
	}
	return env, nil
}
