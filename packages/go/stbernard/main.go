/// 2>/dev/null ; gorun "$0" "$@" ; exit $?

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

package main

import (
	"errors"

	"log/slog"
	"os"

	"github.com/specterops/bloodhound/packages/go/bhlog"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/level"
	"github.com/specterops/bloodhound/packages/go/stbernard/command"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

func main() {
	env, err := environment.NewEnvironment()
	if err != nil {
		slog.Error(
			"Could not init environment",
			attr.Error(err),
		)
		os.Exit(1)
	}

	var rawLvl = env[environment.LogLevelVarName]

	bhlog.ConfigureDefaultText(os.Stderr)

	if rawLvl == "" {
		rawLvl = "info"
	}

	if lvl, err := bhlog.ParseLevel(rawLvl); err != nil {
		slog.Error(
			"Could not parse log level from environment variable",
			slog.String("env_name", environment.LogLevelVarName),
			attr.Error(err),
		)
	} else {
		level.SetGlobalLevel(lvl)
	}

	if cmd, err := command.ParseCLI(env); errors.Is(err, command.ErrNoCmd) {
		slog.Error("No valid command specified")
		os.Exit(1)
	} else if errors.Is(err, command.ErrHelpRequested) {
		// No need to exit 1 if help was requested
		return
	} else if err != nil {
		slog.Error(
			"Error while parsing command",
			attr.Error(err),
		)
		os.Exit(1)
	} else if err := cmd.Run(); err != nil {
		slog.Error(
			"Failed to run command",
			slog.String("command", cmd.Name()),
			attr.Error(err),
		)
		os.Exit(1)
	} else {
		slog.Info(
			"Command completed successfully",
			slog.String("command", cmd.Name()),
		)
	}
}
