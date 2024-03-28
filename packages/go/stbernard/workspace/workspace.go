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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/slicesext"
	"golang.org/x/mod/modfile"
)

const (
	CoverageManifest = "manifest.json"
	CoverageExt      = ".coverage"
	CombinedCoverage = "combined" + CoverageExt
)

var (
	DefaultCoveragePath          = filepath.Join("tmp", "coverage")
	DefaultIntegrationConfigPath = filepath.Join("local-harnesses", "integration.config.json")
)

// WorkspacePaths defines important paths for the current workspace
type WorkspacePaths struct {
	Root     string
	Coverage string
}

// GoPackage represents a parsed Go package
type GoPackage struct {
	Name   string `json:"name"`
	Dir    string `json:"dir"`
	Import string `json:"importpath"`
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
		path = filepath.Join(cwd, DefaultCoveragePath)
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

// ParseModulesAbsPaths parses the modules listed in the go.work file from the given
// directory and returns a list of absolute paths to those modules
func ParseModulesAbsPaths(cwd string) ([]string, error) {
	var workfilePath = filepath.Join(cwd, "go.work")
	// go.work files aren't particularly heavy, so we'll just read into memory
	if data, err := os.ReadFile(workfilePath); err != nil {
		return nil, fmt.Errorf("reading go.work file: %w", err)
	} else if workfile, err := modfile.ParseWork(workfilePath, data, nil); err != nil {
		return nil, fmt.Errorf("parsing go.work file: %w", err)
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

// ParseJSAbsPaths parses list of yarn workspaces from `yarn-workspaces.json` in cwd
func ParseJSAbsPaths(cwd string) ([]string, error) {
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

// moduleListPackages runs go list for the given module and returns the list of packages in that module
func moduleListPackages(modPath string) ([]GoPackage, error) {
	var (
		packages = make([]GoPackage, 0)
	)

	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = modPath
	if out, err := cmd.StdoutPipe(); err != nil {
		return packages, fmt.Errorf("creating stdout pipe for module %s: %w", modPath, err)
	} else if err := cmd.Start(); err != nil {
		return packages, fmt.Errorf("listing packages for module %s: %w", modPath, err)
	} else {
		decoder := json.NewDecoder(out)
		for {
			var p GoPackage
			if err := decoder.Decode(&p); err == io.EOF {
				break
			} else if err != nil {
				return packages, fmt.Errorf("decoding package in module %s: %w", modPath, err)
			}
			packages = append(packages, p)
		}
		cmd.Wait()
		return packages, nil
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

// moduleGenerate runs go generate in each package of the given module
func moduleGenerate(modPath string, env environment.Environment) error {
	var (
		errs []error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)

	if packages, err := moduleListPackages(modPath); err != nil {
		return fmt.Errorf("listing packages for module %s: %w", modPath, err)
	} else {
		for _, pkg := range packages {
			wg.Add(1)
			go func(pkg GoPackage) {
				defer wg.Done()

				var (
					command = "go"
					args    = []string{"generate", pkg.Dir}
				)

				if err := cmdrunner.Run(command, args, pkg.Dir, env); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("generate code for package %s: %w", pkg, err))
					mu.Unlock()
				}
			}(pkg)
		}

		wg.Wait()

		return errors.Join(errs...)
	}
}
