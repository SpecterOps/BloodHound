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

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/analysis"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/builder"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/cover"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/deps"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/envdump"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/generate"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/modsync"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/show"
	"github.com/specterops/bloodhound/packages/go/stbernard/command/tester"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// CommandRunner is an interface for commands, allowing commands to implement the minimum
// set of requirements to observe and run the command from above. It is used as a return
// type to allow passing a usable command to the caller after parsing and creating
// the command implementation
type CommandRunner interface {
	// Name gets the name of the Command
	Name() string
	// Usage gets the usage string for the Command
	Usage() string
	// Run will run the command and return any errors
	Run() error
}

type command interface {
	CommandRunner

	// Parse parses flags for the command using the command index as the starting point
	Parse(cmdIdx int) error
}

var (
	ErrNoCmd           = errors.New("no command specified")
	ErrInvalidCmd      = errors.New("invalid command specified")
	ErrFailedCreateCmd = errors.New("command creation failed")
	ErrMultipleCmd     = errors.New("multiple commands specified")
	ErrHelpRequested   = errors.New("help requested")
)

type usageFunc func()

// ParseCLI parses for a subcommand as the first argument to the calling binary,
// and initializes the command (if it exists). It also provides the default usage
// statement.
//
// It does not support flags of its own, each subcommand is responsible for parsing
// their flags.
func ParseCLI() (CommandRunner, error) {
	var (
		verboseEnabled *bool
		debugEnabled   *bool
		cmdStartIdx    int
		currentCmd     command
		helpRequested  bool

		env      = environment.NewEnvironment()
		commands = []command{
			envdump.Create(env),
			deps.Create(env),
			modsync.Create(env),
			generate.Create(env),
			show.Create(env),
			analysis.Create(env),
			tester.Create(env),
			builder.Create(env),
			cover.Create(env),
		}
	)

	mainCmd := flag.NewFlagSet("main", flag.ExitOnError)

	verboseEnabled = mainCmd.Bool("v", false, "Verbose output")
	debugEnabled = mainCmd.Bool("vv", false, "Debug output")

	for idx, arg := range os.Args {
		if idx == 0 {
			// Skip main command name
			continue
		}

		if cmdStartIdx > 0 {
			break
		}

		if strings.HasPrefix(arg, "-") {
			continue
		}

		if arg == "help" {
			helpRequested = true
			break
		}

		for _, cmd := range commands {
			if arg == cmd.Name() {
				cmdStartIdx = idx
				currentCmd = cmd
				break
			}
		}
	}

	mainCmd.Usage = usageGenerator(mainCmd, commands)

	if err := mainCmd.Parse(os.Args[1:]); errors.Is(err, flag.ErrHelp) || helpRequested {
		mainCmd.Usage()
		return nil, ErrHelpRequested
	}

	if cmdStartIdx == 0 {
		mainCmd.Usage()
		return nil, ErrNoCmd
	}

	if *verboseEnabled {
		log.SetGlobalLevel(log.LevelInfo)
	}

	if *debugEnabled {
		log.SetGlobalLevel(log.LevelDebug)
	}

	return currentCmd, currentCmd.Parse(cmdStartIdx)
}

func usageGenerator(flagset *flag.FlagSet, commands []command) usageFunc {
	return func() {
		var longestCmdLen int

		for _, cmd := range commands {
			if len(cmd.Name()) > longestCmdLen {
				longestCmdLen = len(cmd.Name())
			}
		}

		w := flag.CommandLine.Output()
		fmt.Fprint(w, "A BloodHound Swiss Army Knife\n\nUsage:  stbernard [OPTIONS] COMMAND\n\nOptions:\n")

		flagset.VisitAll(func(f *flag.Flag) {
			padding := strings.Repeat(" ", longestCmdLen-len(f.Name)-1)
			fmt.Fprintf(w, "  -%s%s    %v\n", f.Name, padding, f.Usage)
		})

		fmt.Fprintf(w, "\nCommands:\n")

		for _, cmd := range commands {
			padding := strings.Repeat(" ", longestCmdLen-len(cmd.Name()))
			fmt.Fprintf(w, "  %s%s    %s\n", cmd.Name(), padding, cmd.Usage())
		}
	}
}
