# DORA Metrics - Internal Package

This package provides the core functionality for DORA metrics collection and calculation.

## Authentication

### GitHub Authentication

The DORA metrics tool supports two methods for GitHub authentication:

#### Option 1: GitHub CLI (Recommended)

Use the [GitHub CLI](https://cli.github.com/) for authentication:

```bash
# Install gh CLI (if not already installed)
# macOS: brew install gh
# Linux: See https://github.com/cli/cli/blob/trunk/docs/install_linux.md
# Windows: See https://github.com/cli/cli#installation

# Authenticate with GitHub
gh auth login

# Verify authentication
stbernard dora auth --status
```

The tool will automatically use the GitHub CLI token when available.

**Required Scopes:**
- `repo` - Access to repository data
- `workflow` - Access to GitHub Actions workflows

**Usage:**

```bash
# Check authentication status
stbernard dora auth --status

# Guide to authenticate (prompts to run gh auth login)
stbernard dora auth
```

#### Option 2: Environment Variable (For CI/CD)

Set the `GITHUB_TOKEN` environment variable with a personal access token:

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

**Required Scopes:**
- `repo` - Access to repository data
- `workflow` - Access to GitHub Actions workflows

Create a token at: https://github.com/settings/tokens/new

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
