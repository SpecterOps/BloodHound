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
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/dora"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

// runAuth handles the auth subcommand
func (s *command) runAuth() error {
	var (
		cmd        = flag.NewFlagSet("dora auth", flag.ExitOnError)
		statusFlag bool
	)

	cmd.BoolVar(&statusFlag, "status", false, "Show authentication status")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Authenticate with GitHub or check authentication status\n\n")
		fmt.Fprintf(w, "Usage: %s dora auth [OPTIONS]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "Options:\n")
		cmd.PrintDefaults()
		fmt.Fprintf(w, "\nAuthentication Methods:\n")
		fmt.Fprintf(w, "  1. GitHub CLI (gh): Run 'gh auth login' before using dora\n")
		fmt.Fprintf(w, "  2. Environment: Set GITHUB_TOKEN environment variable\n")
		fmt.Fprintf(w, "\nExamples:\n")
		fmt.Fprintf(w, "  # Check authentication status\n")
		fmt.Fprintf(w, "  %s dora auth -status\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Authenticate with GitHub CLI\n")
		fmt.Fprintf(w, "  %s dora auth\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Using environment variable\n")
		fmt.Fprintf(w, "  export GITHUB_TOKEN=ghp_xxxxxxxxxxxx\n")
		fmt.Fprintf(w, "  %s dora auth -status\n\n", filepath.Base(os.Args[0]))
	}

	if s.subcmdIdx > 0 && s.subcmdIdx+1 < len(os.Args) {
		if err := cmd.Parse(os.Args[s.subcmdIdx+1:]); err != nil {
			return fmt.Errorf("parsing auth flags: %w", err)
		}
	}

	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("finding workspace: %w", err)
	}

	authenticator := dora.NewGitHubAuthenticator(paths.Root, s.env)

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

	// Check gh CLI
	if err := dora.CheckGHCLIAuth(s.env); err == nil {
		fmt.Println("✅ Authenticated via GitHub CLI (gh)")
		token, _ := dora.GetTokenFromGHCLI(s.env)
		if token != nil && token.AccessToken != "" {
			fmt.Printf("Token: %s...%s\n", token.AccessToken[:7], token.AccessToken[len(token.AccessToken)-4:])
		}
		return nil
	} else if errors.Is(err, dora.ErrGHCLINotFound) {
		fmt.Println("❌ GitHub CLI (gh) not installed")
		fmt.Println()
		fmt.Println("Install from: https://cli.github.com/")
		fmt.Println("Or set GITHUB_TOKEN environment variable")
		return nil
	}

	// Not authenticated
	token, err := authenticator.GetToken(ctx)
	if err != nil {
		fmt.Println("❌ Not authenticated")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  1. Run: gh auth login")
		fmt.Println("  2. Set GITHUB_TOKEN environment variable")
		return nil
	}

	fmt.Println("✅ Authenticated")
	if !token.Expiry.IsZero() {
		fmt.Printf("Token expires: %s\n", token.Expiry.Format("2006-01-02 15:04:05"))
	}

	return nil
}

// authenticateGitHub guides user to authenticate with gh CLI
func (s *command) authenticateGitHub(authenticator *dora.GitHubAuthenticator) error {
	ctx := context.Background()

	// Check if already authenticated
	if _, err := authenticator.GetToken(ctx); err == nil {
		fmt.Println("✅ Already authenticated with GitHub")
		fmt.Println()
		if dora.GetTokenFromEnv() != nil {
			fmt.Println("Using GITHUB_TOKEN environment variable")
		} else {
			fmt.Println("Using GitHub CLI (gh)")
		}
		return nil
	}

	// Guide user to authenticate
	return authenticator.AuthenticateWithGHCLI(ctx)
}
