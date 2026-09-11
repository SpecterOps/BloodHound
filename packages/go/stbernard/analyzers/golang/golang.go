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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/codeclimate"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

const (
	customGolangCILintName = "bloodhound-golangci-lint"
	customGolangCILintDir  = "tmp/golangci-lint"
)

// Run golangci-lint for all module paths passed to it
//
// This is a single runner that accepts the paths for all passed modules, rather than separate runs for each path
func Run(cwd string, modPath string, env environment.Environment, fix bool) (codeclimate.SeverityMap, error) {
	var lintEntries []codeclimate.Entry

	if err := buildCustomGolangCILint(cwd, env); err != nil {
		return nil, err
	}

	executionPlan := cmdrunner.ExecutionPlan{
		Command:        customGolangCILintPath(cwd),
		Args:           golangCILintArgs(modPath, fix),
		Path:           cwd,
		Env:            env.Slice(),
		SuppressErrors: true,
	}

	result, err := cmdrunner.Run(context.TODO(), executionPlan)
	if err != nil {
		if !errors.Is(err, cmdrunner.ErrCmdExecutionFailed) {
			return nil, fmt.Errorf("golangci-lint execution: %w", err)
		}

		// exit code 1 is for major or higher analyzer output, higher exit codes indicate something wrong with golangci-lint
		// or its environment, so make sure to fail out
		if result.ReturnCode > 1 {
			return nil, fmt.Errorf("golangci-lint execution: exit code %d", result.ReturnCode)
		}
	}

	if err := json.NewDecoder(result.StandardOutput).Decode(&lintEntries); err != nil {
		return nil, fmt.Errorf("golangci-lint decoding output: %w", err)
	}

	return codeclimate.NewSeverityMap(lintEntries), nil
}

func golangCILintArgs(modPath string, fix bool) []string {
	args := []string{"run", "--config", ".golangci.json", "--output.code-climate.path", "stdout"}
	if fix {
		args = append(args, "--fix")
	}

	return append(args, "--", filepath.Join(modPath, "..."))
}

func buildCustomGolangCILint(cwd string, env environment.Environment) error {
	executionPlan := cmdrunner.ExecutionPlan{
		Command: "go",
		Args:    []string{"tool", "golangci-lint", "custom"},
		Path:    cwd,
		Env:     env.Slice(),
	}

	if _, err := cmdrunner.Run(context.TODO(), executionPlan); err != nil {
		return fmt.Errorf("building custom golangci-lint: %w", err)
	}

	return nil
}

func customGolangCILintPath(cwd string) string {
	return filepath.Join(cwd, customGolangCILintDir, customGolangCILintName)
}
