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

package show

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/git"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name  = "show"
	Usage = "Show current project info"
)

type command struct {
	env environment.Environment
}

type repository struct {
	path    string
	sha     string
	version semver.Version
	clean   bool
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

// Run show command
func (s *command) Run() error {
	if paths, err := workspace.FindPaths(s.env); err != nil {
		return fmt.Errorf("finding workspace paths: %w", err)
	} else if rootRepo, err := s.repositoryCheck(paths.Root); err != nil {
		return fmt.Errorf("repository check: %w", err)
	} else if submodules, err := s.submodulesCheck(paths.Submodules); err != nil {
		return fmt.Errorf("submodule check: %w", err)
	} else {
		repos := []repository{rootRepo}
		repos = append(repos, submodules...)
		for _, repo := range repos {
			fmt.Printf("Repository Report For %s\n", repo.path)
			fmt.Printf("Current HEAD: %s\n", repo.sha)
			fmt.Printf("Detected version: %s\n", repo.version)
			if !repo.clean {
				fmt.Println("CHANGES DETECTED")
				return fmt.Errorf("changes detected in git repository")
			} else {
				fmt.Printf("Repository Clean\n\n")
			}
		}

		return nil
	}
}

func (s *command) submodulesCheck(paths []string) ([]repository, error) {
	var submodules = make([]repository, 0, len(paths))

	for _, path := range paths {
		if repo, err := s.repositoryCheck(path); err != nil {
			return submodules, fmt.Errorf("checking repository for submodule %s: %w", path, err)
		} else {
			submodules = append(submodules, repo)
		}
	}

	return submodules, nil
}

func (s *command) repositoryCheck(cwd string) (repository, error) {
	var repo repository

	if sha, err := git.FetchCurrentCommitSHA(cwd, s.env); err != nil {
		return repo, fmt.Errorf("fetching current commit sha: %w", err)
	} else if version, err := git.ParseLatestVersionFromTags(cwd, s.env); err != nil {
		return repo, fmt.Errorf("parsing version: %w", err)
	} else if clean, err := git.CheckClean(cwd, s.env); err != nil {
		return repo, fmt.Errorf("checking repository clean: %w", err)
	} else {
		repo.path = cwd
		repo.sha = sha
		repo.version = version
		repo.clean = clean

		return repo, nil
	}
}
