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

package analysis

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers"
	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/golang"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"github.com/specterops/bloodhound/packages/go/stbernard/yarn"
)

const (
	Name  = "analysis"
	Usage = "Run static analyzers"
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
	if cwd, err := workspace.FindRoot(); err != nil {
		return fmt.Errorf("finding workspace root: %w", err)
	} else if modPaths, err := workspace.ParseModulesAbsPaths(cwd); err != nil {
		return fmt.Errorf("parsing module absolute paths: %w", err)
	} else if jsPaths, err := workspace.ParseJSAbsPaths(cwd); err != nil {
		return fmt.Errorf("parsing JS absolute paths: %w", err)
	} else if err := golang.InstallGolangCiLint(cwd, s.env); err != nil {
		return fmt.Errorf("installing golangci-lint: %w", err)
	} else if err := yarn.InstallWorkspaceDeps(cwd, jsPaths, s.env); err != nil {
		return fmt.Errorf("yarn install: %w", err)
	} else if result, err := analyzers.Run(cwd, modPaths, jsPaths, s.env); errors.Is(err, analyzers.ErrSeverityExit) {
		fmt.Println(result)
		return err
	} else if err != nil {
		return fmt.Errorf("analyzers incomplete: %w", err)
	} else {
		fmt.Println(result)
		return nil
	}
}
