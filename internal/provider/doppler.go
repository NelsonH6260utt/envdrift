package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const dopplerAPIBase = "https://api.doppler.com/v3/configs/config/secrets/download"

type dopplerClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// DopplerProvider fetches secrets from Doppler using the Secrets Download API.
type DopplerProvider struct {
	token   string
	project string
	config  string
	client  dopplerClient
}

// NewDopplerProvider creates a DopplerProvider.
// token, project, and config are required.
func NewDopplerProvider(token, project, config string) (*DopplerProvider, error) {
	if token == "" {
		return nil, fmt.Errorf("doppler: token is required")
	}
	if project == "" {
		return nil, fmt.Errorf("doppler: project is required")
	}
	if config == "" {
		return nil, fmt.Errorf("doppler: config is required")
	}
	return &DopplerProvider{
		token:   token,
		project: project,
		config:  config,
		client:  &http.Client{},
	}, nil
}

func (d *DopplerProvider) Name() string {
	return fmt.Sprintf("doppler(%s/%s)", d.project, d.config)
}

func (d *DopplerProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	url := fmt.Sprintf("%s?project=%s&config=%s&format=json", dopplerAPIBase, d.project, d.config)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("doppler: build request: %w", err)
	}
	req.SetBasicAuth(d.token, "")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("doppler: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("doppler: unexpected status %d", resp.StatusCode)
	}

	var raw map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("doppler: decode response: %w", err)
	}

	if len(keys) == 0 {
		return raw, nil
	}

	want := make(map[string]bool, len(keys))
	for _, k := range keys {
		want[strings.ToUpper(k)] = true
	}

	result := make(map[string]string)
	for k, v := range raw {
		if want[k] {
			result[k] = v
		}
	}
	return result, nil
}
