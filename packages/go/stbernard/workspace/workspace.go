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

package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/git"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace/golang"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace/yarn"
)

var (
	// No workspace was found in current path
	ErrNoWorkspaceFound = errors.New("found root path without finding project root")
)

// WorkspacePaths defines important paths for the current workspace
type WorkspacePaths struct {
	Root           string
	Coverage       string
	Assets         string
	Submodules     []string
	YarnWorkspaces []string
	GoModules      []string
}

// FindPaths will attempt to crawl up the path until it finds a go.work file, then calculate all WorkspacePaths
func FindPaths(env environment.Environment) (WorkspacePaths, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return WorkspacePaths{}, fmt.Errorf("getting current working directory: %w", err)
	}

	var found bool

	for !found {
		found, err = projectDirExists(cwd)
		if err != nil {
			return WorkspacePaths{}, fmt.Errorf("finding project root: %w", err)
		}

		if found {
			break
		}

		prevCwd := cwd

		// Go up a directory before retrying
		cwd = filepath.Dir(cwd)

		if cwd == prevCwd {
			return WorkspacePaths{}, ErrNoWorkspaceFound
		}
	}

	// Build coverage path
	path, ok := env["SB_COVERAGE_PATH"]
	if !ok || path == "" {
		path = filepath.Join(cwd, golang.DefaultCoveragePath)
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}

	// Build submodule paths
	subPaths, err := git.ListSubmodulePaths(cwd, env)
	if err != nil {
		return WorkspacePaths{}, fmt.Errorf("listing submodule paths: %w", err)
	}

	// Build Yarn paths
	yarnWorkspaces, err := yarn.ParseWorkspace(cwd)
	if err != nil {
		return WorkspacePaths{}, fmt.Errorf("parsing yarn workspace: %w", err)
	}

	// Build Go modules paths
	goModules, err := golang.ParseModulesAbsPaths(cwd)
	if err != nil {
		return WorkspacePaths{}, fmt.Errorf("parsing go module paths: %w", err)
	}

	return WorkspacePaths{
		Root:           cwd,
		Coverage:       path,
		Submodules:     subPaths,
		Assets:         yarnWorkspaces.AssetsDir,
		YarnWorkspaces: yarnWorkspaces.Workspaces,
		GoModules:      goModules,
	}, nil
}

// GenerateSchema runs schemagen for the current workspace
func GenerateSchema(cwd string, env environment.Environment) error {
	var (
		command = "go"
		args    = []string{"run"}
	)

	if _, err := os.Stat(filepath.Join(cwd, "cmd", "schemagen")); !errors.Is(err, os.ErrNotExist) && err != nil {
		return fmt.Errorf("attempted to find cmd/schemagen: %w", err)
	} else if errors.Is(err, os.ErrNotExist) {
		args = append(args, "github.com/specterops/bloodhound/schemagen")
	} else {
		args = append(args, "git.bloodhound-ad.net/schemagen")
	}

	if err := cmdrunner.Run(command, args, cwd, env); err != nil {
		return fmt.Errorf("running schemagen: %w", err)
	} else {
		return nil
	}
}

func projectDirExists(cwd string) (bool, error) {
	if _, err := os.Stat(filepath.Join(cwd, "go.work")); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("stat go.work file: %w", err)
	} else {
		return true, nil
	}
}
