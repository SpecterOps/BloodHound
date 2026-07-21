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
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/dora"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name  = "dora"
	Usage = "DORA metrics collection and reporting"
)

var (
	ErrNoSubcommand  = errors.New("no subcommand specified")
	ErrInvalidSubcmd = errors.New("invalid subcommand")
)

type command struct {
	env        environment.Environment
	subcommand string
	subcmdIdx  int
	flagSet    *flag.FlagSet
}

// Create new instance of DORA command
func Create(env environment.Environment) *command {
	return &command{
		env: env,
	}
}

// Usage returns the usage string
func (s *command) Usage() string {
	return Usage
}

// Name returns the command name
func (s *command) Name() string {
	return Name
}

// Parse parses command flags and determines subcommand
func (s *command) Parse(cmdIndex int) error {
	s.flagSet = flag.NewFlagSet(Name, flag.ExitOnError)

	s.flagSet.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [SUBCOMMAND] [OPTIONS]\n\n", Usage, filepath.Base(os.Args[0]), Name)
		fmt.Fprintf(w, "Subcommands:\n")
		fmt.Fprintf(w, "  init      Initialize DORA metrics configuration\n")
		fmt.Fprintf(w, "  auth      Authenticate with GitHub\n")
		fmt.Fprintf(w, "  collect   Collect data from GitHub\n")
		fmt.Fprintf(w, "  status    Show configuration and authentication status\n")
		fmt.Fprintf(w, "\nOptions:\n")
		s.flagSet.PrintDefaults()
	}

	// Find the subcommand
	for idx := cmdIndex + 1; idx < len(os.Args); idx++ {
		arg := os.Args[idx]

		// Skip flags
		if len(arg) > 0 && arg[0] == '-' {
			continue
		}

		// Found subcommand
		s.subcommand = arg
		s.subcmdIdx = idx
		break
	}

	if err := s.flagSet.Parse(os.Args[cmdIndex+1:]); err != nil {
		s.flagSet.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	return nil
}

// Run executes the DORA command
func (s *command) Run() error {
	switch s.subcommand {
	case "init":
		return s.runInit()
	case "auth":
		return s.runAuth()
	case "collect":
		return s.runCollect()
	case "status":
		return s.runStatus()
	case "":
		s.flagSet.Usage()
		return ErrNoSubcommand
	default:
		s.flagSet.Usage()
		return fmt.Errorf("%w: %s", ErrInvalidSubcmd, s.subcommand)
	}
}

// runInit initializes the DORA configuration
func (s *command) runInit() error {
	var (
		cmd       = flag.NewFlagSet("dora init", flag.ExitOnError)
		forceFlag bool
		localFlag bool
	)

	cmd.BoolVar(&forceFlag, "force", false, "Overwrite existing configuration")
	cmd.BoolVar(&localFlag, "local", false, "Create local override configuration (.dora.local.yaml)")

	if s.subcmdIdx > 0 && s.subcmdIdx+1 < len(os.Args) {
		if err := cmd.Parse(os.Args[s.subcmdIdx+1:]); err != nil {
			return fmt.Errorf("parsing init flags: %w", err)
		}
	}

	return s.initCommand(forceFlag, localFlag)
}

// runStatus shows the current configuration and authentication status
func (s *command) runStatus() error {
	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("finding workspace: %w", err)
	}

	fmt.Println("DORA Metrics Status")
	fmt.Println("===================")
	fmt.Println()

	// Check for configuration file
	configPath := filepath.Join(paths.Root, dora.ConfigFileName)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("❌ Configuration file not found: %s\n", configPath)
		fmt.Println("\nRun 'dora init' to create a configuration file.")
		return nil
	}

	// Load configuration
	config, err := dora.LoadConfig(paths.Root)
	if err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}

	fmt.Printf("✅ Configuration file: %s\n", configPath)
	fmt.Println()

	// Show configuration
	fmt.Println("GitHub Configuration:")
	fmt.Printf("  Owner: %s\n", config.GitHub.Owner)
	fmt.Printf("  Repo: %s\n", config.GitHub.Repo)
	fmt.Printf("  Production Workflow: %s\n", config.GitHub.Production.Workflow)
	fmt.Printf("  Production Environment: %s\n", config.GitHub.Production.Environment)
	fmt.Println()

	if config.JIRA.Domain != "" {
		fmt.Println("JIRA Configuration:")
		fmt.Printf("  Domain: %s\n", config.JIRA.Domain)
		fmt.Printf("  Projects: %v\n", config.JIRA.ProjectKeys)
		fmt.Println()
	}

	fmt.Println("Storage Configuration:")
	fmt.Printf("  Type: %s\n", config.Storage.Type)
	fmt.Printf("  Path: %s\n", config.GetStoragePath(paths.Root))
	fmt.Println()

	// Check if data directory exists
	doraDir := filepath.Join(paths.Root, dora.DoraDataDir)
	if stat, err := os.Stat(doraDir); err == nil && stat.IsDir() {
		fmt.Printf("✅ Data directory exists: %s\n", doraDir)
	} else {
		fmt.Printf("ℹ️  Data directory will be created on first use: %s\n", doraDir)
	}

	return nil
}

// initCommand creates the DORA configuration files and directories
func (s *command) initCommand(force, local bool) error {
	paths, err := workspace.FindPaths(s.env)
	if err != nil {
		return fmt.Errorf("finding workspace: %w", err)
	}

	var (
		configPath = filepath.Join(paths.Root, dora.ConfigFileName)
		doraDir    = filepath.Join(paths.Root, dora.DoraDataDir)
	)

	if local {
		configPath = filepath.Join(paths.Root, dora.LocalConfigFileName)
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil && !force {
		return fmt.Errorf("configuration file already exists: %s (use --force to overwrite)", configPath)
	}

	// Create configuration
	config := dora.DefaultConfig()

	if local {
		// For local config, create a minimal override template
		config = dora.Config{
			JIRA: dora.JIRAConfig{
				Domain: "",
			},
		}
		slog.Info("Creating local configuration override", slog.String("path", configPath))
	} else {
		slog.Info("Creating DORA configuration", slog.String("path", configPath))
	}

	if err := config.SaveToFile(configPath); err != nil {
		return fmt.Errorf("saving configuration: %w", err)
	}

	fmt.Printf("✅ Created configuration file: %s\n", configPath)

	// Create data directory
	if err := os.MkdirAll(doraDir, 0755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	tokensDir := config.GetTokensDir(paths.Root)
	if err := os.MkdirAll(tokensDir, 0700); err != nil {
		return fmt.Errorf("creating tokens directory: %w", err)
	}

	fmt.Printf("✅ Created data directory: %s\n", doraDir)
	fmt.Println()

	if local {
		fmt.Println("Local configuration created. Add your overrides to this file.")
		fmt.Println("This file is gitignored and can contain sensitive information.")
	} else {
		fmt.Println("DORA metrics initialized successfully!")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  1. Review the configuration in .dora.yaml")
		fmt.Println("  2. Authenticate with GitHub:")
		fmt.Println("     - Option A: Set GITHUB_TOKEN environment variable")
		fmt.Println("     - Option B: Run 'dora auth' (requires OAuth app setup)")
		fmt.Println("  3. Start collecting data with 'dora collect'")
	}

	return nil
}
