package provider

import (
	"net/http"
	"net/url"
	"strings"
)

// baseURLTransport rewrites the host of every request to a fixed base URL.
// Used in tests to redirect API calls to a local httptest.Server.
type baseURLTransport struct {
	base      *url.URL
	wrapped   http.RoundTripper
}

func (t *baseURLTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = t.base.Scheme
	clone.URL.Host = t.base.Host
	// strip any path prefix so query params are preserved
	if strings.HasPrefix(clone.URL.Path, "/v3") {
		clone.URL.Path = "/"
	}
	return t.wrapped.RoundTrip(clone)
}

// newBaseURLClient returns an *http.Client that redirects all requests to baseURL.
func newBaseURLClient(baseURL string) *http.Client {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		panic("newBaseURLClient: invalid URL: " + err.Error())
	}
	return &http.Client{
		Transport: &baseURLTransport{
			base:    parsed,
			wrapped: http.DefaultTransport,
		},
	}
}
