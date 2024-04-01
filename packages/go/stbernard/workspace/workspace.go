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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace/golang"
)

// WorkspacePaths defines important paths for the current workspace
type WorkspacePaths struct {
	Root     string
	Coverage string
}

// Config represents a St Bernard configuration
type Config struct {
	DistDir   string `json:"dist_dir"`
	AssetsDir string `json:"assets_dir"`
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
			return WorkspacePaths{}, errors.New("found root path without finding project root")
		}
	}

	path, ok := env["SB_COVERAGE_PATH"]
	if !ok || path == "" {
		path = filepath.Join(cwd, golang.DefaultCoveragePath)
	} else if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}

	return WorkspacePaths{Root: cwd, Coverage: path}, nil
}

// ParseConfig parses a configuration file in the .stbernard directory for the given workspace path
func ParseConfig(cwd string) (Config, error) {
	var cfg Config

	if bytes, err := os.ReadFile(filepath.Join(cwd, ".stbernard", "config.json")); err != nil {
		return cfg, fmt.Errorf("reading config file: %w", err)
	} else if err := json.Unmarshal(bytes, &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshaling config file contents: %w", err)
	} else {
		return cfg, nil
	}
}

// projectDirExists checks if a .stbernard directory exists in the given working directory
func projectDirExists(cwd string) (bool, error) {
	if _, err := os.Stat(filepath.Join(cwd, ".stbernard")); errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("stat .stbernard file: %w", err)
	} else {
		return true, nil
	}
}
