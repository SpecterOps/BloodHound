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

package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/specterops/bloodhound/slicesext"
	"golang.org/x/mod/modfile"
)

// FindRoot will attempt to crawl up the path until it finds a go.work file
func FindRoot() (string, error) {
	if cwd, err := os.Getwd(); err != nil {
		return "", fmt.Errorf("could not get current working directory: %w", err)
	} else {
		var found bool

		for !found {
			found, err = WorkFileExists(cwd)
			if err != nil {
				return cwd, fmt.Errorf("error while trying to find go.work file: %w", err)
			}

			if found {
				break
			}

			prevCwd := cwd

			// Go up a directory before retrying
			cwd = filepath.Dir(cwd)

			if cwd == prevCwd {
				return cwd, errors.New("found root path without finding go.work file")
			}
		}

		return cwd, nil
	}
}

// WorkFileExists checks if a go.work file exists in the given directory
func WorkFileExists(cwd string) (bool, error) {
	if _, err := os.Stat(filepath.Join(cwd, "go.work")); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("could not stat go.work file: %w", err)
	} else {
		return true, nil
	}
}

// ParseModulesAbsPaths parses the modules listed in the go.work file from the given
// directory and returns a list of absolute paths to those modules
func ParseModulesAbsPaths(cwd string) ([]string, error) {
	var workfilePath = filepath.Join(cwd, "go.work")
	// go.work files aren't particularly heavy, so we'll just read into memory
	if data, err := os.ReadFile(workfilePath); err != nil {
		return nil, fmt.Errorf("could not read go.work file: %w", err)
	} else if workfile, err := modfile.ParseWork(workfilePath, data, nil); err != nil {
		return nil, fmt.Errorf("could not parse go.work file: %w", err)
	} else {
		var (
			modulePaths = make([]string, 0, len(workfile.Use))
			workDir     = filepath.Dir(workfilePath)
		)

		for _, use := range workfile.Use {
			modulePaths = append(modulePaths, filepath.Join(workDir, use.Path))
		}

		return modulePaths, nil
	}
}

func ParseJsAbsPaths(cwd string) ([]string, error) {
	var (
		paths  []string
		ywPath = filepath.Join(cwd, "yarn-workspaces.json")
	)

	if data, err := os.ReadFile(ywPath); err != nil {
		return paths, fmt.Errorf("could not read yarn-workspaces.json file: %w", err)
	} else if err := json.Unmarshal(data, &paths); err != nil {
		return paths, fmt.Errorf("could not unmarshal yarn-workspaces.json file: %w", err)
	} else {
		var workDir = filepath.Dir(ywPath)

		return slicesext.Map(paths, func(path string) string { return filepath.Join(workDir, path) }), nil
	}
}

// DownloadModules runs go mod download for all module paths passed with a given
// set of environment variables
func DownloadModules(modPaths []string, env []string) error {
	var errs = make([]error, 0)

	for _, modPath := range modPaths {
		cmd := exec.Command("go", "mod", "download")
		cmd.Env = env
		cmd.Dir = modPath
		if err := cmd.Run(); err != nil {
			errs = append(errs, fmt.Errorf("failure when running command: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to download all modules: %w", errors.Join(errs...))
	} else {
		return nil
	}
}

// SyncWorkspace runs go work sync in the given directory with a given set of environment
// variables
func SyncWorkspace(cwd string, env []string) error {
	cmd := exec.Command("go", "work", "sync")
	cmd.Env = env
	cmd.Dir = cwd
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed running go work sync: %w", err)
	} else {
		return nil
	}
}
