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
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/codeclimate"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// Run golangci-lint for all module paths passed to it
//
// This is a single runner that accepts the paths for all passed modules, rather than separate runs for each path
func Run(cwd string, modPath string, env environment.Environment) (codeclimate.SeverityMap, error) {
	var (
		lintEntries []codeclimate.Entry
		output      *bytes.Buffer
		command     = "go"
		args        = []string{"tool", "golangci-lint", "run", "--fix", "--config", ".golangci.json", "--output.code-climate.path", "stdout", "--"}
	)

	args = append(args, filepath.Join(modPath, "..."))

	if result, err := cmdrunner.Run(command, args, cwd, env); err != nil {
		var errResult *cmdrunner.ExecutionError

		// exit code 1 is for major or higher analyzer output, higher exit codes indicate something wrong with golangci-lint
		// or its environment, so make sure to fail out
		if !errors.As(err, &errResult) || errResult.ReturnCode > 1 {
			return nil, fmt.Errorf("golangci-lint execution: %w", err)
		}

		output = errResult.StandardOutput
	} else {
		output = result.StandardOutput
	}

	if err := json.NewDecoder(output).Decode(&lintEntries); err != nil {
		return nil, fmt.Errorf("golangci-lint decoding output: %w", err)
	}

	return codeclimate.NewSeverityMap(lintEntries), nil
}
