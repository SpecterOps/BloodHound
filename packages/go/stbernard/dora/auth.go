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
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

const (
	GitHubTokenFileName = "github-token.json"
	// GitHub OAuth App Client ID for DORA metrics
	// This is a public client ID for device flow authentication
	GitHubClientID = "Ov23liQwertyExample" // TODO: Replace with actual OAuth app client ID
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenExpired  = errors.New("token is expired")
	ErrTokenInvalid  = errors.New("token is invalid")
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

// GitHubAuthenticator handles GitHub OAuth authentication
type GitHubAuthenticator struct {
	workspaceRoot string
	config        *oauth2.Config
}

// NewGitHubAuthenticator creates a new GitHub authenticator
func NewGitHubAuthenticator(workspaceRoot string) *GitHubAuthenticator {
	return &GitHubAuthenticator{
		workspaceRoot: workspaceRoot,
		config: &oauth2.Config{
			ClientID: GitHubClientID,
			Endpoint: oauth2.Endpoint{
				AuthURL:       "https://github.com/login/oauth/authorize",
				TokenURL:      "https://github.com/login/oauth/access_token",
				DeviceAuthURL: "https://github.com/login/device/code",
			},
			Scopes: []string{"repo", "workflow"},
		},
	}
}

// GetToken retrieves a valid token from file or environment
func (s *GitHubAuthenticator) GetToken(ctx context.Context) (*oauth2.Token, error) {
	// Try environment variable first
	if token := GetTokenFromEnv(); token != nil {
		return token, nil
	}

	// Try loading from file
	tokenPath := GetGitHubTokenPath(s.workspaceRoot)
	token, err := LoadToken(tokenPath)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return nil, fmt.Errorf("no GitHub token found. Run 'dora auth github' to authenticate")
		}
		return nil, err
	}

	// Validate token
	if err := ValidateToken(token); err != nil {
		return nil, fmt.Errorf("stored token is invalid: %w", err)
	}

	return token, nil
}

// AuthenticateDeviceFlow performs OAuth device flow authentication
func (s *GitHubAuthenticator) AuthenticateDeviceFlow(ctx context.Context) (*oauth2.Token, error) {
	// Request device code
	deviceCode, err := s.config.DeviceAuth(ctx)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}

	// Display instructions to user
	fmt.Println()
	fmt.Println("GitHub Authentication Required")
	fmt.Println("==============================")
	fmt.Println()
	fmt.Printf("Please visit: %s\n", deviceCode.VerificationURI)
	fmt.Printf("And enter code: %s\n", deviceCode.UserCode)
	fmt.Println()
	fmt.Println("Waiting for authorization...")
	fmt.Println()

	// Poll for token
	token, err := s.config.DeviceAccessToken(ctx, deviceCode)
	if err != nil {
		return nil, fmt.Errorf("waiting for authorization: %w", err)
	}

	return token, nil
}

// SaveAuthToken saves the authentication token to file
func (s *GitHubAuthenticator) SaveAuthToken(token *oauth2.Token) error {
	tokenPath := GetGitHubTokenPath(s.workspaceRoot)
	return SaveToken(tokenPath, token)
}

// ClearToken removes the stored authentication token
func (s *GitHubAuthenticator) ClearToken() error {
	tokenPath := GetGitHubTokenPath(s.workspaceRoot)
	if err := os.Remove(tokenPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Already removed
		}
		return fmt.Errorf("removing token file: %w", err)
	}
	return nil
}
