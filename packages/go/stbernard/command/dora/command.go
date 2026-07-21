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
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
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
	var (
		cmd = flag.NewFlagSet(Name, flag.ExitOnError)
	)

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [SUBCOMMAND] [OPTIONS]\n\n", Usage, filepath.Base(os.Args[0]), Name)
		fmt.Fprintf(w, "Subcommands:\n")
		fmt.Fprintf(w, "  init      Initialize DORA metrics configuration\n")
		fmt.Fprintf(w, "  status    Show configuration and authentication status\n")
		fmt.Fprintf(w, "\nOptions:\n")
		cmd.PrintDefaults()
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

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	return nil
}

// Run executes the DORA command
func (s *command) Run() error {
	switch s.subcommand {
	case "init":
		return s.runInit()
	case "status":
		return s.runStatus()
	case "":
		return ErrNoSubcommand
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSubcmd, s.subcommand)
	}
}

// runInit initializes the DORA configuration
func (s *command) runInit() error {
	fmt.Println("Initializing DORA metrics configuration...")
	fmt.Println("This feature is under development.")
	return nil
}

// runStatus shows the current configuration and authentication status
func (s *command) runStatus() error {
	fmt.Println("DORA Metrics Status")
	fmt.Println("===================")
	fmt.Println("This feature is under development.")
	return nil
}
