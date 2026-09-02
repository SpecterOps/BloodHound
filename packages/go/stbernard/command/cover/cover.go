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

package cover

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/coverfiles"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace/golang"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace/yarn"
	"golang.org/x/tools/cover"
)

const (
	Name  = "cover"
	Usage = "Collect coverage reports"
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

// Run cover command
func (s *command) Run() error {
	if paths, err := workspace.FindPaths(s.env); err != nil {
		return fmt.Errorf("finding workspace root: %w", err)
	} else if yarnWork, err := yarn.ParseWorkspace(paths.Root); err != nil {
		return fmt.Errorf("parsing yarn workspace paths: %w", err)
	} else if profiles, err := getProfilesFromManifest(paths.Coverage); err != nil {
		return fmt.Errorf("getting coverage manifest: %w", err)
	} else if err := writeCombinedProfiles(filepath.Join(paths.Coverage, golang.CombinedCoverage), profiles); err != nil {
		return fmt.Errorf("writing combined profiles: %w", err)
	} else if goCovPercent, err := golang.GetCombinedCoverage(filepath.Join(paths.Coverage, golang.CombinedCoverage), s.env); err != nil {
		return fmt.Errorf("getting total go coverage: %w", err)
	} else if yarnCovPercent, err := yarn.GetCombinedCoverage(yarnWork.Workspaces, s.env); err != nil {
		return fmt.Errorf("getting total yarn coverage: %w", err)
	} else {
		fmt.Printf("Total Go Test Coverage in %s: %s\n", paths.Root, goCovPercent)
		fmt.Printf("Total Yarn Test Coverage in %s: %s\n", paths.Root, yarnCovPercent)
		return nil
	}
}

func getProfilesFromManifest(coverPath string) ([]*cover.Profile, error) {
	var (
		profiles []*cover.Profile

		manifestFile = filepath.Join(coverPath, golang.CoverageManifest)
		manifest     = make(map[string]string)
	)

	if manifestBytes, err := os.ReadFile(manifestFile); err != nil {
		return profiles, fmt.Errorf("opening %s: %w", manifestFile, err)
	} else if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return profiles, fmt.Errorf("unmarshal manifest: %w", err)
	} else {
		profiles = make([]*cover.Profile, 0, len(manifest)*32)
		for _, coverFile := range manifest {
			if p, err := cover.ParseProfiles(coverFile); err != nil {
				return profiles, fmt.Errorf("parsing profiles for %s: %w", coverFile, err)
			} else {
				profiles = append(profiles, p...)
			}
		}

		return profiles, nil
	}
}

func writeCombinedProfiles(fileName string, profiles []*cover.Profile) error {
	if combinedFile, err := os.Create(fileName); err != nil {
		return fmt.Errorf("creating combined profile file %s: %w", fileName, err)
	} else if err := coverfiles.WriteProfile(combinedFile, profiles); err != nil {
		combinedFile.Close()
		return fmt.Errorf("writing combined profile %s: %w", fileName, err)
	} else if err := combinedFile.Close(); err != nil {
		return fmt.Errorf("closing combined profile %s: %w", fileName, err)
	} else {
		return nil
	}
}
