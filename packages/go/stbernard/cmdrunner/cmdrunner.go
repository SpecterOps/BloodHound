// Copyright 2024 Specter Ops, Inc.
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

package cmdrunner

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

var (
	ErrNonZeroExit = errors.New("non-zero exit status")
)

// Run a command with ars and environment variables set at a specified path
//
// The CmdModifiers parameter is an optional list of modifying functions that can alter the generated *exec.Cmd after default setup.
//
// If debug log level is set globally, command output will be combined and sent to os.Stderr.
func Run(command string, args []string, path string, env environment.Environment, cmdModifiers ...func(*exec.Cmd)) error {
	var (
		exitErr error

		cmdstr       = command + " " + args[0]
		cmd          = exec.Command(command, args...)
		debugEnabled = log.GlobalAccepts(log.LevelDebug)
	)

	cmd.Env = env.Slice()
	cmd.Dir = path

	// Default to mapping stdout directly to stdout
	cmd.Stdout = os.Stdout

	if debugEnabled {
		cmd.Stderr = os.Stderr
		cmdstr = command + " " + strings.Join(args, " ")
	}

	// If we got any cmdModifiers, apply them in order
	// This is often used for capturing Stdout and other modifications
	if len(cmdModifiers) > 0 {
		for _, modifier := range cmdModifiers {
			modifier(cmd)
		}
	}

	log.Infof("Running %s for %s", cmdstr, path)

	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		exitErr = ErrNonZeroExit
	} else if err != nil {
		return fmt.Errorf("%s: %w", cmdstr, err)
	}

	log.Infof("Finished %s for %s", cmdstr, path)

	return exitErr
}
