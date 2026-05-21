package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

// AWSProvider fetches environment variables from AWS SSM Parameter Store.
type AWSProvider struct {
	client    *ssm.Client
	pathPrefix string
}

// NewAWSProvider creates a new AWSProvider using the default AWS config.
// pathPrefix is the SSM path prefix, e.g. "/myapp/prod/".
func NewAWSProvider(ctx context.Context, pathPrefix string) (*AWSProvider, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws: load config: %w", err)
	}
	return &AWSProvider{
		client:    ssm.NewFromConfig(cfg),
		pathPrefix: pathPrefix,
	}, nil
}

// Name returns the provider identifier.
func (a *AWSProvider) Name() string {
	return "aws-ssm"
}

// FetchEnv retrieves parameters from SSM under the configured path prefix.
// If keys is non-empty, only those keys are fetched (appended to the prefix).
// If keys is empty, all parameters under the prefix are returned.
func (a *AWSProvider) FetchEnv(ctx context.Context, keys []string) (map[string]string, error) {
	result := make(map[string]string)

	if len(keys) > 0 {
		names := make([]string, len(keys))
		for i, k := range keys {
			names[i] = a.pathPrefix + k
		}
		out, err := a.client.GetParameters(ctx, &ssm.GetParametersInput{
			Names:          names,
			WithDecryption: aws.Bool(true),
		})
		if err != nil {
			return nil, fmt.Errorf("aws: get parameters: %w", err)
		}
		for _, p := range out.Parameters {
			key := stripPrefix(*p.Name, a.pathPrefix)
			result[key] = aws.ToString(p.Value)
		}
		return result, nil
	}

	paginator := ssm.NewGetParametersByPathPaginator(a.client, &ssm.GetParametersByPathInput{
		Path:           aws.String(a.pathPrefix),
		Recursive:      aws.Bool(false),
		WithDecryption: aws.Bool(true),
	})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("aws: get parameters by path: %w", err)
		}
		for _, p := range page.Parameters {
			key := stripPrefix(*p.Name, a.pathPrefix)
			result[key] = aws.ToString(p.Value)
		}
	}
	return result, nil
}

func stripPrefix(s, prefix string) string {
	if len(s) > len(prefix) && s[:len(prefix)] == prefix {
		return s[len(prefix):]
	}
	return s
}
