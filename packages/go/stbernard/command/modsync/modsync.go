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

package modsync

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name  = "modsync"
	Usage = "Sync all modules in current workspace"
)

type flags struct {
	verbose bool
}

type Config struct {
	flags       flags
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
	} else if err := workspace.DownloadModules(modPaths, s.config.Environment); err != nil {
		return fmt.Errorf("could not download modules: %w", err)
	} else if err := workspace.SyncWorkspace(cwd, s.config.Environment); err != nil {
		return fmt.Errorf("could not sync workspace: %w", err)
	} else {
		return nil
	}
}

func Create(config Config) (command, error) {
	modsyncCmd := flag.NewFlagSet(Name, flag.ExitOnError)
	modsyncCmd.BoolVar(&config.flags.verbose, "v", false, "Print verbose logs")

	modsyncCmd.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s\n\nUsage: %s %s [OPTIONS]\n\nOptions:\n", Usage, filepath.Base(os.Args[0]), Name)
		modsyncCmd.PrintDefaults()
	}

	if err := modsyncCmd.Parse(os.Args[2:]); err != nil {
		modsyncCmd.Usage()
		return command{}, fmt.Errorf("failed to parse modsync command: %w", err)
	} else {
		return command{config: config}, nil
	}
}
