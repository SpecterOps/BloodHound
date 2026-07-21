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
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/specterops/bloodhound/packages/go/stbernard/dora"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

// runAuth handles the auth subcommand
func (s *command) runAuth() error {
	var (
		cmd        = flag.NewFlagSet("dora auth", flag.ExitOnError)
		clearFlag  bool
		statusFlag bool
	)

	cmd.BoolVar(&clearFlag, "clear", false, "Clear stored authentication token")
	cmd.BoolVar(&statusFlag, "status", false, "Show authentication status")

	if s.subcmdIdx > 0 && s.subcmdIdx+1 < len(os.Args) {
		if err := cmd.Parse(os.Args[s.subcmdIdx+1:]); err != nil {
			return fmt.Errorf("parsing auth flags: %w", err)
		}
	}

	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("finding workspace: %w", err)
	}

	authenticator := dora.NewGitHubAuthenticator(paths.Root)

	// Handle clear flag
	if clearFlag {
		if err := authenticator.ClearToken(); err != nil {
			return fmt.Errorf("clearing token: %w", err)
		}
		fmt.Println("✅ Authentication token cleared")
		return nil
	}

	// Handle status flag
	if statusFlag {
		return s.showAuthStatus(authenticator)
	}

	// Default: authenticate
	return s.authenticateGitHub(authenticator)
}

// showAuthStatus displays the current authentication status
func (s *command) showAuthStatus(authenticator *dora.GitHubAuthenticator) error {
	ctx := context.Background()

	fmt.Println("GitHub Authentication Status")
	fmt.Println("============================")
	fmt.Println()

	// Check environment variable
	if envToken := dora.GetTokenFromEnv(); envToken != nil {
		fmt.Println("✅ Using token from GITHUB_TOKEN environment variable")
		return nil
	}

	// Check stored token
	token, err := authenticator.GetToken(ctx)
	if err != nil {
		if errors.Is(err, dora.ErrTokenNotFound) {
			fmt.Println("❌ Not authenticated")
			fmt.Println()
			fmt.Println("Run 'dora auth github' to authenticate")
			return nil
		}
		return fmt.Errorf("checking authentication: %w", err)
	}

	if err := dora.ValidateToken(token); err != nil {
		fmt.Printf("⚠️  Token is invalid: %v\n", err)
		fmt.Println()
		fmt.Println("Run 'dora auth github' to re-authenticate")
		return nil
	}

	fmt.Println("✅ Authenticated")
	if !token.Expiry.IsZero() {
		fmt.Printf("Token expires: %s\n", token.Expiry.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// authenticateGitHub performs GitHub OAuth device flow authentication
func (s *command) authenticateGitHub(authenticator *dora.GitHubAuthenticator) error {
	ctx := context.Background()

	// Check if already authenticated
	if token, err := authenticator.GetToken(ctx); err == nil && dora.ValidateToken(token) == nil {
		fmt.Println("Already authenticated with GitHub")
		fmt.Println("Use --clear to remove existing authentication")
		return nil
	}

	// Perform device flow authentication
	fmt.Println("Starting GitHub authentication...")
	fmt.Println()

	token, err := authenticator.AuthenticateDeviceFlow(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save token
	if err := authenticator.SaveAuthToken(token); err != nil {
		return fmt.Errorf("saving token: %w", err)
	}

	fmt.Println("✅ Successfully authenticated with GitHub!")
	fmt.Println()
	fmt.Println("Token has been saved securely.")

	return nil
}
