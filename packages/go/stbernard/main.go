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
	"fmt"
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/bhlog"
	"github.com/specterops/bloodhound/bhlog/level"
	"github.com/specterops/bloodhound/packages/go/stbernard/command"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

func main() {
	env := environment.NewEnvironment()
	var rawLvl = env[environment.LogLevelVarName]

	bhlog.ConfigureDefaultText(os.Stderr)

	if rawLvl == "" {
		rawLvl = "warn"
	}

	if lvl, err := bhlog.ParseLevel(rawLvl); err != nil {
		slog.Error(fmt.Sprintf("Could not parse log level from %s: %v", environment.LogLevelVarName, err))
	} else {
		level.GlobalAccepts(lvl)
	}

	if cmd, err := command.ParseCLI(env); errors.Is(err, command.ErrNoCmd) {
		slog.Error("No valid command specified")
		os.Exit(1)
	} else if errors.Is(err, command.ErrHelpRequested) {
		// No need to exit 1 if help was requested
		return
	} else if err != nil {
		slog.Error(fmt.Sprintf("Error while parsing command: %v", err))
		os.Exit(1)
	} else if err := cmd.Run(); err != nil {
		slog.Error(fmt.Sprintf("Failed to run command `%s`: %v", cmd.Name(), err))
		os.Exit(1)
	} else {
		slog.Info(fmt.Sprintf("Command `%s` completed successfully", cmd.Name()))
	}
}
