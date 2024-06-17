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

// Create new instance of command to capture given environment
func Create(env environment.Environment) *command {
	return &command{
		env: env,
	}
}

// Usage of command
func (s *command) Usage() string {
	return Usage
}

// Name of command
func (s *command) Name() string {
	return Name
}

// Parse command flags
func (s *command) Parse(cmdIndex int) error {
	cmd := flag.NewFlagSet(Name, flag.ExitOnError)
	targetOs := cmd.String("os", "", "Specify the OS to build for. This will override anything in GOOS.")
	targetArch := cmd.String("arch", "", "Specify the architecture to build for. This will override anything in GOARCH.")

	cmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		cmd.PrintDefaults()
	}

	if err := cmd.Parse(os.Args[cmdIndex+1:]); err != nil {
		cmd.Usage()
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}

	if *targetOs != "" {
		s.env.Override("GOOS", *targetOs)
	}
	if *targetArch != "" {
		s.env.Override("GOARCH", *targetArch)
	}

	return nil
}

// Run build command
func (s *command) Run() error {
	if paths, err := workspace.FindPaths(s.env); err != nil {
		return fmt.Errorf("finding workspace root: %w", err)
	} else if err := filepath.WalkDir(paths.Assets, clearFiles); err != nil {
		return fmt.Errorf("clearing asset directory: %w", err)
	} else if err := s.runJSBuild(paths.Root, paths.Assets); err != nil {
		return fmt.Errorf("building JS artifacts: %w", err)
	} else if err := s.runGoBuild(paths.Root, paths.GoModules); err != nil {
		return fmt.Errorf("building Go artifacts: %w", err)
	} else {
		return nil
	}
}

func (s *command) runJSBuild(cwd string, buildPath string) error {
	s.env.SetIfEmpty("BUILD_PATH", buildPath)

	if err := yarn.BuildWorkspace(cwd, s.env); err != nil {
		return fmt.Errorf("building JS workspace: %w", err)
	} else {
		return nil
	}
}

func (s command) runGoBuild(cwd string, modPaths []string) error {
	s.env.SetIfEmpty("CGO_ENABLED", "0")

	if err := golang.BuildMainPackages(cwd, modPaths, s.env); err != nil {
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
