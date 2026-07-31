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
	"testing"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

func TestTokenFromEnv(t *testing.T) {
	// Set environment variable
	testToken := "gho_envtoken123"
	os.Setenv("GITHUB_TOKEN", testToken)
	defer os.Unsetenv("GITHUB_TOKEN")

	token := GetTokenFromEnv()
	if token == nil {
		t.Fatal("Expected token from environment, got nil")
	}
	// Check nil again before dereferencing to satisfy linter
	if token != nil && token.AccessToken != testToken {
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

	// Check nil again before dereferencing to satisfy linter
	if token != nil && token.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}

	if token != nil && token.TokenType != "Bearer" {
		t.Errorf("Expected token type Bearer, got %s", token.TokenType)
	}
}
