# envdrift

Detects configuration drift between `.env` files and deployed environment variables in cloud providers.

## Installation

```bash
go install github.com/yourusername/envdrift@latest
```

## Usage

Compare your local `.env` file against a deployed environment (e.g., AWS, GCP, Heroku):

```bash
# Compare .env with an AWS Lambda function
envdrift check --provider aws --resource my-lambda-function --env .env

# Compare .env with a Heroku app
envdrift check --provider heroku --resource my-app --env .env.production
```

Example output:

```
Checking drift for: my-lambda-function
✔  DATABASE_URL        matched
✘  API_KEY             missing in deployed environment
✘  CACHE_TTL           local=300, deployed=60
⚠  OLD_SECRET          present in deployed but not in .env

Drift detected: 3 issue(s) found
```

### Supported Providers

| Provider | Status |
|----------|--------|
| AWS Lambda / ECS | ✅ |
| Google Cloud Run | ✅ |
| Heroku | ✅ |
| Azure App Service | 🚧 In progress |

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--provider` | Cloud provider to target | `aws` |
| `--resource` | Resource name or ID | required |
| `--env` | Path to local `.env` file | `.env` |
| `--strict` | Exit with non-zero code on any drift | `false` |

## Contributing

Pull requests are welcome. Please open an issue first to discuss any significant changes.

## License

[MIT](LICENSE)