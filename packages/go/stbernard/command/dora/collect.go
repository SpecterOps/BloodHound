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
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/dora"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

// runCollect handles the collect subcommand
func (s *command) runCollect() error {
	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("finding workspace: %w", err)
	}

	// Load configuration to get default period
	config, err := dora.LoadConfig(paths.Root)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Parse default period from config
	defaultDays := parseDefaultPeriod(config.Metrics.DefaultPeriod)

	var (
		cmd        = flag.NewFlagSet("dora collect", flag.ExitOnError)
		daysFlag   int
		deployFlag bool
		commitFlag bool
		prFlag     bool
	)

	cmd.IntVar(&daysFlag, "days", defaultDays, fmt.Sprintf("Number of days to collect data for (default: %s from config)", config.Metrics.DefaultPeriod))
	cmd.BoolVar(&deployFlag, "deployments", false, "Collect deployment data only")
	cmd.BoolVar(&commitFlag, "commits", false, "Collect commit data only")
	cmd.BoolVar(&prFlag, "prs", false, "Collect pull request data only")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Collect DORA metrics data from GitHub\n\n")
		fmt.Fprintf(w, "Usage: %s dora collect [OPTIONS]\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "Options:\n")
		cmd.PrintDefaults()
		fmt.Fprintf(w, "\nData Types:\n")
		fmt.Fprintf(w, "  - Deployments: Git tags (semver format: v1.2.3, v1.2.3-rc1)\n")
		fmt.Fprintf(w, "  - Commits: All commits in the main branch\n")
		fmt.Fprintf(w, "  - Pull Requests: Merged PRs with timestamps\n")
		fmt.Fprintf(w, "\nExamples:\n")
		fmt.Fprintf(w, "  # Collect all data for last 90 days (default)\n")
		fmt.Fprintf(w, "  %s dora collect\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Collect last 365 days\n")
		fmt.Fprintf(w, "  %s dora collect -days 365\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Collect only deployments\n")
		fmt.Fprintf(w, "  %s dora collect -deployments\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "  # Collect commits and PRs only\n")
		fmt.Fprintf(w, "  %s dora collect -commits -prs\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "\nNote: Data is stored in SQLite database at .dora/dora.db\n")
	}

	if s.subcmdIdx > 0 && s.subcmdIdx+1 < len(os.Args) {
		if err := cmd.Parse(os.Args[s.subcmdIdx+1:]); err != nil {
			return fmt.Errorf("parsing collect flags: %w", err)
		}
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Create storage
	storagePath := config.Storage.Path
	if !filepath.IsAbs(storagePath) {
		storagePath = filepath.Join(paths.Root, storagePath)
	}

	storage, err := dora.NewStorage(storagePath)
	if err != nil {
		return fmt.Errorf("creating storage: %w", err)
	}
	defer storage.Close()

	// Create collector
	collector := dora.NewGitHubCollector(&config, s.env)

	// Calculate time range
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -daysFlag)

	ctx := context.Background()

	// If no specific flags, collect everything
	collectAll := !deployFlag && !commitFlag && !prFlag

	// Collect deployments
	if collectAll || deployFlag {
		fmt.Printf("Collecting deployments from %s to %s...\n",
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"))

		deployments, err := collector.CollectDeployments(ctx, startTime, endTime)
		if err != nil {
			return fmt.Errorf("collecting deployments: %w", err)
		}
		fmt.Printf("✅ Collected %d deployments\n", len(deployments))

		if len(deployments) > 0 {
			if err := storage.SaveDeployments(ctx, deployments); err != nil {
				return fmt.Errorf("saving deployments: %w", err)
			}
			fmt.Printf("💾 Saved %d deployments to database\n", len(deployments))
		}
	}

	// Collect commits
	if collectAll || commitFlag {
		fmt.Printf("Collecting commits from %s to %s...\n",
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"))

		commits, err := collector.CollectCommits(ctx, startTime, endTime)
		if err != nil {
			return fmt.Errorf("collecting commits: %w", err)
		}
		fmt.Printf("✅ Collected %d commits\n", len(commits))

		if len(commits) > 0 {
			if err := storage.SaveCommits(ctx, commits); err != nil {
				return fmt.Errorf("saving commits: %w", err)
			}
			fmt.Printf("💾 Saved %d commits to database\n", len(commits))
		}
	}

	// Collect pull requests
	if collectAll || prFlag {
		fmt.Printf("Collecting pull requests from %s to %s...\n",
			startTime.Format("2006-01-02"),
			endTime.Format("2006-01-02"))

		prs, err := collector.CollectPullRequests(ctx, startTime, endTime)
		if err != nil {
			return fmt.Errorf("collecting pull requests: %w", err)
		}
		fmt.Printf("✅ Collected %d pull requests\n", len(prs))

		if len(prs) > 0 {
			if err := storage.SavePullRequests(ctx, prs); err != nil {
				return fmt.Errorf("saving pull requests: %w", err)
			}
			fmt.Printf("💾 Saved %d pull requests to database\n", len(prs))
		}
	}

	fmt.Println()
	fmt.Printf("✅ Data collection complete! Database: %s\n", storagePath)

	return nil
}
