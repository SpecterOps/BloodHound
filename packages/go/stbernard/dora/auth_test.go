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
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestTokenStorage(t *testing.T) {
	tempDir := t.TempDir()
	tokenPath := filepath.Join(tempDir, "github-token.json")

	// Create a test token
	token := &oauth2.Token{
		AccessToken:  "gho_testtoken123",
		TokenType:    "Bearer",
		RefreshToken: "refresh123",
		Expiry:       time.Now().Add(24 * time.Hour),
	}

	// Save token
	if err := SaveToken(tokenPath, token); err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Fatal("Token file was not created")
	}

	// Load token
	loadedToken, err := LoadToken(tokenPath)
	if err != nil {
		t.Fatalf("Failed to load token: %v", err)
	}

	// Verify token contents
	if loadedToken.AccessToken != token.AccessToken {
		t.Errorf("Expected access token %s, got %s", token.AccessToken, loadedToken.AccessToken)
	}
	if loadedToken.TokenType != token.TokenType {
		t.Errorf("Expected token type %s, got %s", token.TokenType, loadedToken.TokenType)
	}
}

func TestLoadTokenNotFound(t *testing.T) {
	tempDir := t.TempDir()
	tokenPath := filepath.Join(tempDir, "nonexistent.json")

	_, err := LoadToken(tokenPath)
	if err == nil {
		t.Error("Expected error when loading non-existent token")
	}
}

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
