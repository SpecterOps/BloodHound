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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/slicesext"
)

const yarnWorkFile = "yarn-workspaces.json"

var (
	CoverageFile         = filepath.Join("coverage", "coverage-summary.json")
	ErrNoStatementsFound = errors.New("no statements found in coverage")
)

type coverage struct {
	Total covTotal `json:"total"`
}

type covTotal struct {
	Statements statements `json:"statements"`
}

type statements struct {
	Total   int `json:"total"`
	Covered int `json:"covered"`
}

// Workspace contains Yarn-related directories
type Workspace struct {
	AssetsDir  string   `json:"assets_dir"`
	Workspaces []string `json:"workspaces"`
}

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

// Format runs yarn format on current workspace
func Format(cwd string, env environment.Environment) error {
	var (
		command = "yarn"
		args    = []string{"format"}
	)

	if err := cmdrunner.Run(command, args, cwd, env); err != nil {
		return fmt.Errorf("running yarn format: %w", err)
	} else {
		return nil
	}
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

// ParseWorkspace parses `yarn-workspaces.json` for paths
func ParseWorkspace(cwd string) (Workspace, error) {
	var (
		relWorkspace Workspace
		ywPath       = filepath.Join(cwd, yarnWorkFile)
	)

	if b, err := os.ReadFile(ywPath); err != nil {
		return relWorkspace, fmt.Errorf("reading yarn-workspaces.json file: %w", err)
	} else if err := json.Unmarshal(b, &relWorkspace); err != nil {
		return relWorkspace, fmt.Errorf("unmarshaling yarn-workspaces.json file: %w", err)
	} else {
		return relWorkspaceToAbsWorkspace(cwd, relWorkspace), nil
	}
}

// GetCombinedCoverage combines statement coverage for given yarn workspaces and returns a single percentage value as a string
func GetCombinedCoverage(yarnAbsPaths []string, env environment.Environment) (string, error) {
	var (
		totalAccumulator   int
		coveredAccumulator int
	)

	for _, path := range yarnAbsPaths {
		if cov, err := getCoverage(filepath.Join(path, CoverageFile)); err != nil {
			return "", fmt.Errorf("getting coverage: %w", err)
		} else {
			totalAccumulator += cov.Total.Statements.Total
			coveredAccumulator += cov.Total.Statements.Covered
		}
	}

	if totalAccumulator == 0 {
		return "", ErrNoStatementsFound
	}

	return fmt.Sprintf("%.1f%%", float64(coveredAccumulator)/float64(totalAccumulator)*100), nil
}

func relWorkspaceToAbsWorkspace(cwd string, relWorkspace Workspace) Workspace {
	var absWorkspace Workspace

	absWorkspace.AssetsDir = filepath.Join(cwd, relWorkspace.AssetsDir)
	absWorkspace.Workspaces = slicesext.Map(relWorkspace.Workspaces, func(path string) string { return filepath.Join(cwd, path) })

	return absWorkspace
}

func getCoverage(coverFile string) (coverage, error) {
	var cov coverage
	if b, err := os.ReadFile(coverFile); err != nil {
		log.Warnf("Could not find coverage for %s, skipping", coverFile)
		return cov, nil
	} else if err := json.Unmarshal(b, &cov); err != nil {
		return cov, fmt.Errorf("unmarshal coverage file %s: %w", coverFile, err)
	} else {
		return cov, nil
	}
}
