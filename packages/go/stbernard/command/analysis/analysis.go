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
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"github.com/specterops/bloodhound/packages/go/stbernard/yarn"
)

const (
	Name  = "analysis"
	Usage = "Run static analyzers"
)

type Config struct {
	Environment []string
}

type command struct {
	config Config
}

func (s command) Usage() string {
	return Usage
}

func (s command) Name() string {
	return Name
}

func (s command) Run() error {
	if cwd, err := workspace.FindRoot(); err != nil {
		return fmt.Errorf("could not find workspace root: %w", err)
	} else if modPaths, err := workspace.ParseModulesAbsPaths(cwd); err != nil {
		return fmt.Errorf("could not parse module absolute paths: %w", err)
	} else if jsPaths, err := workspace.ParseJsAbsPaths(cwd); err != nil {
		return fmt.Errorf("could not parse JS absolute paths: %w", err)
	} else if err := preAnalysisSetup(jsPaths, s.config.Environment); err != nil {
		return fmt.Errorf("could not complete environmental setup: %w", err)
	} else if result, err := analyzers.Run(cwd, modPaths, jsPaths, s.config.Environment); errors.Is(err, analyzers.ErrSeverityExit) {
		fmt.Println(result)
		return err
	} else if err != nil {
		return fmt.Errorf("analyzers could not run completely: %w", err)
	} else {
		fmt.Println(result)
		return nil
	}
}

func Create(config Config) (command, error) {
	analysisCmd := flag.NewFlagSet(Name, flag.ExitOnError)

	analysisCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		analysisCmd.PrintDefaults()
	}

	if err := analysisCmd.Parse(os.Args[2:]); err != nil {
		analysisCmd.Usage()
		return command{}, fmt.Errorf("failed to parse analysis command: %w", err)
	} else {
		return command{config: config}, nil
	}
}

func preAnalysisSetup(jsPaths []string, env []string) error {
	if err := golang.InstallGolangCiLint(env); err != nil {
		return fmt.Errorf("golangci-lint failed to install: %w", err)
	} else if err := yarn.InstallWorkspaceDeps(jsPaths, env); err != nil {
		return fmt.Errorf("yarn install failed: %w", err)
	} else {
		return nil
	}
}
