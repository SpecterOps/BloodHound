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

package yarn

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/slicesext"
)

// InstallWorkspaceDeps runs yarn install for a given list of jsPaths
func InstallWorkspaceDeps(cwd string, jsPaths []string, env environment.Environment) error {
	var (
		command = "yarn"
		args    = []string{"install"}
	)

	for _, path := range jsPaths {
		if err := cmdrunner.Run(command, args, path, env); err != nil {
			return fmt.Errorf("yarn install at %v: %w", path, err)
		}
	}

	return nil
}

// BuildWorkspace runs yarn build for the current working directory
func BuildWorkspace(cwd string, env environment.Environment) error {
	var (
		command = "yarn"
		args    = []string{"build"}
	)

	if err := cmdrunner.Run(command, args, cwd, env); err != nil {
		return fmt.Errorf("yarn build at %v: %w", cwd, err)
	} else {
		return nil
	}
}

// TestWorkspace runs yarn tests for all yarn workspaces
func TestWorkspace(cwd string, env environment.Environment) error {
	var (
		command = "yarn"
		args    = []string{"test", "--coverage", "--run"}
	)

	if err := cmdrunner.Run(command, args, cwd, env); err != nil {
		return fmt.Errorf("yarn test at %v: %w", cwd, err)
	} else {
		return nil
	}
}

// ParseYarnAbsPaths parses list of yarn workspaces from `yarn-workspaces.json` in cwd
func ParseYarnAbsPaths(cwd string) ([]string, error) {
	var (
		paths  []string
		ywPath = filepath.Join(cwd, "yarn-workspaces.json")
	)

	if data, err := os.ReadFile(ywPath); err != nil {
		return paths, fmt.Errorf("reading yarn-workspaces.json file: %w", err)
	} else if err := json.Unmarshal(data, &paths); err != nil {
		return paths, fmt.Errorf("unmarshaling yarn-workspaces.json file: %w", err)
	} else {
		return slicesext.Map(paths, func(path string) string { return filepath.Join(filepath.Dir(ywPath), path) }), nil
	}
}
