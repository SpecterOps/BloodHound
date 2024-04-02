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

package golang

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path"

	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/codeclimate"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/slicesext"
)

// Run golangci-lint for all module paths passed to it
//
// This is a single runner that accepts the paths for all passed modules, rather than separate runs for each path
func Run(cwd string, modPaths []string, env environment.Environment) ([]codeclimate.Entry, error) {
	var (
		result []codeclimate.Entry
		outb   bytes.Buffer

		command        = "golangci-lint"
		args           = []string{"run", "--out-format", "code-climate", "--config", ".golangci.json", "--"}
		redirectStdout = func(c *exec.Cmd) { c.Stdout = &outb }
	)

	args = append(args, slicesext.Map(modPaths, func(modPath string) string {
		return path.Join(modPath, "...")
	})...)

	cmdErr := cmdrunner.Run(command, args, cwd, env, redirectStdout)
	// If the command has a non-zero exit, we're going to return it up the stack, but we want to attempt to process the output anyway
	if cmdErr != nil && !errors.Is(cmdErr, cmdrunner.ErrNonZeroExit) {
		return result, fmt.Errorf("unexpected run error: %w", cmdErr)
	}

	if err := json.NewDecoder(&outb).Decode(&result); err != nil {
		return result, fmt.Errorf("golangci-lint: decoding output: %w", err)
	}

	return result, cmdErr
}
