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
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"golang.org/x/oauth2"
)

func TestTokenValidation(t *testing.T) {
	tests := []struct {
		name    string
		token   *oauth2.Token
		wantErr bool
	}{
		{
			name: "valid token",
			token: &oauth2.Token{
				AccessToken: "gho_validtoken",
				Expiry:      time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "expired token",
			token: &oauth2.Token{
				AccessToken: "gho_expiredtoken",
				Expiry:      time.Now().Add(-24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "empty token",
			token: &oauth2.Token{
				AccessToken: "",
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateToken(tc.token)
			if tc.wantErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestGetTokenPath(t *testing.T) {
	workspaceRoot := "/tmp/test-workspace"
	expectedPath := filepath.Join(workspaceRoot, DoraDataDir, TokensDir, "github-token.json")

	path := GetGitHubTokenPath(workspaceRoot)
	if path != expectedPath {
		t.Errorf("Expected token path %s, got %s", expectedPath, path)
	}
}

func TestTokenFromEnv(t *testing.T) {
	// Set environment variable
	testToken := "gho_envtoken123"
	os.Setenv("GITHUB_TOKEN", testToken)
	defer os.Unsetenv("GITHUB_TOKEN")

	token := GetTokenFromEnv()
	if token == nil {
		t.Fatal("Expected token from environment, got nil")
	}
	if token.AccessToken != testToken {
		t.Errorf("Expected access token %s, got %s", testToken, token.AccessToken)
	}
}

func TestTokenFromEnvNotSet(t *testing.T) {
	os.Unsetenv("GITHUB_TOKEN")

	token := GetTokenFromEnv()
	if token != nil {
		t.Errorf("Expected nil token when env not set, got %v", token)
	}
}

func TestGHCLIIntegration(t *testing.T) {
	env, err := environment.NewEnvironment()
	if err != nil {
		t.Fatalf("Failed to create environment: %v", err)
	}

	// Check if gh CLI is available
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh CLI not installed, skipping integration test")
	}

	// Check if authenticated
	if err := CheckGHCLIAuth(env); err != nil {
		t.Skip("gh CLI not authenticated, skipping integration test")
	}

	// Try to get token
	token, err := GetTokenFromGHCLI(env)
	if err != nil {
		t.Fatalf("Failed to get token from gh CLI: %v", err)
	}

	if token == nil {
		t.Fatal("Expected token, got nil")
	}

	if token.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if token.TokenType != "Bearer" {
		t.Errorf("Expected token type Bearer, got %s", token.TokenType)
	}
}
