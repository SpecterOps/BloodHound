// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package dora

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"golang.org/x/oauth2"
)

const (
	GitHubTokenFileName = "github-token.json"
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token is expired")
	ErrTokenInvalid  = errors.New("token is invalid")
	ErrGHCLINotFound = errors.New("GitHub CLI (gh) not found - install from https://cli.github.com/")
)

// SaveToken saves an OAuth token to a file
func SaveToken(path string, token *oauth2.Token) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling token: %w", err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating token directory: %w", err)
	}

	// Write token file with restricted permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing token file: %w", err)
	}

	return nil
}

// LoadToken loads an OAuth token from a file
func LoadToken(path string) (*oauth2.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrTokenNotFound, path)
		}
		return nil, fmt.Errorf("reading token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("parsing token file: %w", err)
	}

	return &token, nil
}

// ValidateToken checks if a token is valid and not expired
func ValidateToken(token *oauth2.Token) error {
	if token == nil {
		return ErrTokenInvalid
	}
	if token.AccessToken == "" {
		return ErrTokenInvalid
	}
	if !token.Expiry.IsZero() && token.Expiry.Before(time.Now()) {
		return ErrTokenExpired
	}
	return nil
}

// GetGitHubTokenPath returns the path to the GitHub token file
func GetGitHubTokenPath(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, DoraDataDir, TokensDir, GitHubTokenFileName)
}

// GetTokenFromEnv retrieves a GitHub token from the GITHUB_TOKEN environment variable
func GetTokenFromEnv() *oauth2.Token {
	tokenStr := os.Getenv("GITHUB_TOKEN")
	if tokenStr == "" {
		return nil
	}

	return &oauth2.Token{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
	}
}

// GetTokenFromGHCLI retrieves the GitHub token from the gh CLI using cmdrunner
func GetTokenFromGHCLI(env environment.Environment) (*oauth2.Token, error) {
	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, ErrGHCLINotFound
	}

	// Run gh auth token
	executionPlan := cmdrunner.ExecutionPlan{
		Command:        "gh",
		Args:           []string{"auth", "token"},
		Env:            env.Slice(),
		SuppressErrors: true,
	}

	result, err := cmdrunner.Run(context.Background(), executionPlan)
	if err != nil {
		return nil, fmt.Errorf("getting token from gh CLI: %w", err)
	}

	tokenStr := strings.TrimSpace(result.StandardOutput.String())
	if tokenStr == "" {
		return nil, errors.New("gh CLI returned empty token - run 'gh auth login' first")
	}

	return &oauth2.Token{
		AccessToken: tokenStr,
		TokenType:   "Bearer",
	}, nil
}

// CheckGHCLIAuth checks if the user is authenticated with gh CLI
func CheckGHCLIAuth(env environment.Environment) error {
	if _, err := exec.LookPath("gh"); err != nil {
		return ErrGHCLINotFound
	}

	executionPlan := cmdrunner.ExecutionPlan{
		Command:        "gh",
		Args:           []string{"auth", "status"},
		Env:            env.Slice(),
		SuppressErrors: true,
	}

	_, err := cmdrunner.Run(context.Background(), executionPlan)
	if err != nil {
		return errors.New("not authenticated with gh CLI - run 'gh auth login' first")
	}

	return nil
}

// GitHubAuthenticator handles GitHub authentication
type GitHubAuthenticator struct {
	workspaceRoot string
	env           environment.Environment
}

// NewGitHubAuthenticator creates a new GitHub authenticator
func NewGitHubAuthenticator(workspaceRoot string, env environment.Environment) *GitHubAuthenticator {
	return &GitHubAuthenticator{
		workspaceRoot: workspaceRoot,
		env:           env,
	}
}

// GetToken retrieves a valid token from environment or gh CLI
func (s *GitHubAuthenticator) GetToken(ctx context.Context) (*oauth2.Token, error) {
	// Try environment variable first
	if token := GetTokenFromEnv(); token != nil {
		return token, nil
	}

	// Try gh CLI
	token, err := GetTokenFromGHCLI(s.env)
	if err != nil {
		if errors.Is(err, ErrGHCLINotFound) {
			return nil, fmt.Errorf("no GitHub token found. Install gh CLI from https://cli.github.com/ or set GITHUB_TOKEN environment variable")
		}
		return nil, err
	}

	return token, nil
}

// AuthenticateWithGHCLI guides the user to authenticate with gh CLI
func (s *GitHubAuthenticator) AuthenticateWithGHCLI(ctx context.Context) error {
	// Check if gh CLI is installed
	if _, err := exec.LookPath("gh"); err != nil {
		return ErrGHCLINotFound
	}

	// Check if already authenticated
	if err := CheckGHCLIAuth(s.env); err == nil {
		fmt.Println("✅ Already authenticated with GitHub CLI")
		return nil
	}

	// Guide user to authenticate
	fmt.Println()
	fmt.Println("GitHub Authentication Required")
	fmt.Println("==============================")
	fmt.Println()
	fmt.Println("Please run the following command to authenticate:")
	fmt.Println()
	fmt.Println("  gh auth login")
	fmt.Println()
	fmt.Println("Follow the prompts to complete authentication.")
	fmt.Println()

	return errors.New("authentication required - please run 'gh auth login'")
}
