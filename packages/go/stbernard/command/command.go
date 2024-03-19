// Copyright 2023 Specter Ops, Inc.
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

package command

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/specterops/bloodhound/packages/go/stbernard/command/analysis"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/builder"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/generate"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/modsync"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/tester"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// Commander is an interface for commands, allowing commands to implement the minimum
// set of requirements to observe and run the command from above. It is used as a return
// type to allow passing a usable command to the caller after parsing and creating
// the command implementation
type Commander interface {
	Name() string
	Usage() string
	Run() error
}

var (
	ErrNoCmd           = errors.New("no command specified")
	ErrInvalidCmd      = errors.New("invalid command specified")
	ErrFailedCreateCmd = errors.New("command creation failed")
	ErrMultipleCmd     = errors.New("multiple commands specified")
)

// ParseCLI parses for a subcommand as the first argument to the calling binary,
// and initializes the command (if it exists). It also provides the default usage
// statement.
//
// It does not support flags of its own, each subcommand is responsible for parsing
// their flags.
func ParseCLI() (Commander, error) {
	var (
		cmdName string
		err     error

		env = environment.NewEnvironment()
	)

	// Generate a nice usage message
	flag.Usage = usage

	// Default usage if no arguments provided
	if len(os.Args) < 2 {
		flag.Usage()
		return nil, ErrNoCmd
	}

	for _, arg := range os.Args[1:] {
		if cmdName != "" {
			err = ErrMultipleCmd
			break
		}

		for _, command := range Commands() {
			if arg == command.String() {
				cmdName = command.String()
			}
		}
	}

	if err != nil {
		flag.Parse()
		flag.Usage()
		return nil, err
	}

	switch cmdName {
	case ModSync.String():
		config := modsync.Config{Environment: env}
		if cmd, err := modsync.Create(config); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedCreateCmd, err)
		} else {
			return cmd, nil
		}

	case Generate.String():
		config := generate.Config{Environment: env}
		if cmd, err := generate.Create(config); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedCreateCmd, err)
		} else {
			return cmd, nil
		}

	case Test.String():
		config := tester.Config{Environment: env}
		if cmd, err := tester.Create(config); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedCreateCmd, err)
		} else {
			return cmd, nil
		}

	case Analysis.String():
		config := analysis.Config{Environment: env}
		if cmd, err := analysis.Create(config); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedCreateCmd, err)
		} else {
			return cmd, nil
		}

	case Build.String():
		config := builder.Config{Environment: env}
		if cmd, err := builder.Create(config); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedCreateCmd, err)
		} else {
			return cmd, nil
		}

	case EnvDump.String():
		config := envdump.Config{Environment: env}
		if cmd, err := envdump.Create(config); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedCreateCmd, err)
		} else {
			return cmd, nil
		}

	default:
		flag.Parse()
		flag.Usage()
		return nil, ErrInvalidCmd
	}
}

// usage creates a pretty usage message for our main command
func usage() {
	var longestCmdLen int

	w := flag.CommandLine.Output()
	fmt.Fprint(w, "A BloodHound Swiss Army Knife\n\nUsage:  stbernard COMMAND\n\nCommands:\n")

	for _, cmd := range Commands() {
		if len(cmd.String()) > longestCmdLen {
			longestCmdLen = len(cmd.String())
		}
	}

	for cmd, usage := range CommandsUsage() {
		cmdStr := Command(cmd).String()
		padding := strings.Repeat(" ", longestCmdLen-len(cmdStr))
		fmt.Fprintf(w, "  %s%s    %s\n", cmdStr, padding, usage)
	}
}
