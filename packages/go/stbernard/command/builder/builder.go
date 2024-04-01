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

package builder

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace/golang"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace/yarn"
)

const (
	Name  = "build"
	Usage = "Build commands in current workspace"
)

type command struct {
	env environment.Environment
}

func Create(env environment.Environment) *command {
	return &command{
		env: env,
	}
}

func (s *command) Usage() string {
	return Usage
}

func (s *command) Name() string {
	return Name
}

func (s *command) Parse(cmdIndex int) error {
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	return nil
}

func (s *command) Run() error {
	if paths, err := workspace.FindPaths(s.env); err != nil {
		return fmt.Errorf("finding workspace root: %w", err)
	} else if cfg, err := workspace.ParseConfig(paths.Root); err != nil {
		return fmt.Errorf("getting build configuration file: %w", err)
	} else if err := filepath.WalkDir(filepath.Join(paths.Root, cfg.AssetsDir), clearFiles); err != nil {
		return fmt.Errorf("clearing asset directory: %w", err)
	} else if err := s.runJSBuild(paths.Root, filepath.Join(paths.Root, cfg.AssetsDir)); err != nil {
		return fmt.Errorf("building JS artifacts: %w", err)
	} else if err := s.runGoBuild(paths.Root); err != nil {
		return fmt.Errorf("building Go artifacts: %w", err)
	} else {
		return nil
	}
}

func (s *command) runJSBuild(cwd string, buildPath string) error {
	s.env.SetIfEmpty("BUILD_PATH", buildPath)

	if jsPaths, err := yarn.ParseYarnAbsPaths(cwd); err != nil {
		return fmt.Errorf("retrieving JS paths: %w", err)
	} else if err := yarn.InstallWorkspaceDeps(cwd, jsPaths, s.env); err != nil {
		return fmt.Errorf("installing JS deps: %w", err)
	} else if err := yarn.BuildWorkspace(cwd, s.env); err != nil {
		return fmt.Errorf("building JS workspace: %w", err)
	} else {
		return nil
	}
}

func (s command) runGoBuild(cwd string) error {
	s.env.SetIfEmpty("CGO_ENABLED", "0")

	if modPaths, err := golang.ParseModulesAbsPaths(cwd); err != nil {
		return fmt.Errorf("parsing module absolute paths: %w", err)
	} else if err := golang.BuildMainPackages(cwd, modPaths, s.env); err != nil {
		return fmt.Errorf("building main packages: %w", err)
	} else {
		return nil
	}
}

func clearFiles(path string, entry os.DirEntry, err error) error {
	if err != nil {
		return err
	}

	if slices.Contains([]string{".keep", "keep", ".gitkeep", "gitkeep"}, entry.Name()) {
		// Early return
		return nil
	}

	log.Debugf("Removing %s", filepath.Join(path, entry.Name()))

	if entry.IsDir() {
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	} else {
		if err := os.Remove(path); err != nil {
			return err
		}
	}

	return nil
}
