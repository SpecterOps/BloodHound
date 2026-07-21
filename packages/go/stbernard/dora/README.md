# DORA Metrics - Internal Package

This package provides the core functionality for DORA metrics collection and calculation.

## Authentication

### GitHub Authentication

The DORA metrics tool supports two methods for GitHub authentication:

#### Option 1: Environment Variable (Recommended for CI/CD)

Set the `GITHUB_TOKEN` environment variable with a personal access token:

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

**Required Scopes:**
- `repo` - Access to repository data
- `workflow` - Access to GitHub Actions workflows

Create a token at: https://github.com/settings/tokens/new

#### Option 2: OAuth Device Flow (Interactive)

For interactive authentication, an OAuth App must be configured in GitHub.

**Setup OAuth App:**

1. Go to GitHub Settings > Developer settings > OAuth Apps
2. Click "New OAuth App"
3. Fill in the details:
   - Application name: "DORA Metrics Tool"
   - Homepage URL: `https://github.com/SpecterOps/bloodhound-enterprise`
   - Authorization callback URL: (leave empty for device flow)
4. Click "Register application"
5. Copy the Client ID
6. Update `GitHubClientID` in `auth.go` with your Client ID

**Note:** OAuth device flow does NOT require a client secret for public clients.

**Usage:**

```bash
# Authenticate interactively
stbernard dora auth

# Check authentication status
stbernard dora auth --status

# Clear stored token
stbernard dora auth --clear
```

### Token Storage

Tokens are stored securely in:
- File: `<workspace>/.dora/tokens/github-token.json`
- Permissions: `0600` (owner read/write only)
- Format: JSON (OAuth2 token structure)

The tokens directory is gitignored to prevent accidental commits.

## Testing

```bash
# Run all tests
go test ./... -v

# Run specific tests
go test -v -run TestToken

# Test with coverage
go test -cover ./...
```

## Configuration

See the main documentation in `docs/dora-metrics/` for full configuration details.
