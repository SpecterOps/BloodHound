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
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/dora"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

// runCollect handles the collect subcommand
func (s *command) runCollect() error {
	var (
		cmd        = flag.NewFlagSet("dora collect", flag.ExitOnError)
		daysFlag   int
		deployFlag bool
		commitFlag bool
		prFlag     bool
	)

	cmd.IntVar(&daysFlag, "days", 30, "Number of days to collect data for")
	cmd.BoolVar(&deployFlag, "deployments", false, "Collect deployment data only")
	cmd.BoolVar(&commitFlag, "commits", false, "Collect commit data only")
	cmd.BoolVar(&prFlag, "prs", false, "Collect pull request data only")

	if s.subcmdIdx > 0 && s.subcmdIdx+1 < len(os.Args) {
		if err := cmd.Parse(os.Args[s.subcmdIdx+1:]); err != nil {
			return fmt.Errorf("parsing collect flags: %w", err)
		}
	}

	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("finding workspace: %w", err)
	}

	// Load configuration
	config, err := dora.LoadConfig(paths.Root)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

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
	}

	fmt.Println()
	fmt.Println("✅ Data collection complete!")
	fmt.Println()
	fmt.Println("Note: Data is not yet stored (storage layer coming in next phase)")

	return nil
}
