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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/git"
)

// BuildGoMainPackages builds all main packages for a list of module paths
func BuildGoMainPackages(workRoot string, modPaths []string, env []string) error {
	var (
		errs     []error
		wg       sync.WaitGroup
		mu       sync.Mutex
		buildDir = filepath.Join(workRoot, "dist") + string(filepath.Separator)
	)

	version, err := git.ParseLatestVersionFromTags(workRoot, env)
	if err != nil {
		return fmt.Errorf("failed to parse latest version from git tags: %w", err)
	}

	log.Infof("Building for version %s", version.Original())

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(buildDir, modPath string) {
			defer wg.Done()
			if err := buildGoModuleMainPackages(buildDir, modPath, version, env); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("failed to build main package: %w", err))
				mu.Unlock()
			}
		}(buildDir, modPath)
	}

	wg.Wait()

	return errors.Join(errs...)
}

// buildGoModuleMainPackages runs go build for all main packages in a given module
func buildGoModuleMainPackages(buildDir string, modPath string, version semver.Version, env []string) error {
	var (
		wg   sync.WaitGroup
		errs []error
		mu   sync.Mutex

		majorString         = fmt.Sprintf("-X 'github.com/specterops/bloodhound/src/version.majorVersion=%d'", version.Major())
		minorString         = fmt.Sprintf("-X 'github.com/specterops/bloodhound/src/version.minorVersion=%d'", version.Minor())
		patchString         = fmt.Sprintf("-X 'github.com/specterops/bloodhound/src/version.patchVersion=%d'", version.Patch())
		prereleaseString    = fmt.Sprintf("-X 'github.com/specterops/bloodhound/src/version.prereleaseVersion=%s'", version.Prerelease())
		ldflagArgComponents = []string{majorString, minorString, patchString}
	)

	if version.Prerelease() != "" {
		ldflagArgComponents = append(ldflagArgComponents, prereleaseString)
	}

	args := []string{"-ldflags", strings.Join(ldflagArgComponents, " "), "-o", buildDir}

	if packages, err := moduleListPackages(modPath); err != nil {
		return fmt.Errorf("failed to list module packages: %w", err)
	} else {
		for _, p := range packages {
			if p.Name == "main" && !strings.Contains(p.Dir, "plugin") {
				wg.Add(1)
				go func(p GoPackage) {
					defer wg.Done()
					cmd := exec.Command("go", "build")
					cmd.Args = append(cmd.Args, args...)
					cmd.Env = env
					cmd.Dir = p.Dir
					if log.GlobalAccepts(log.LevelDebug) {
						cmd.Stdout = os.Stderr
						cmd.Stderr = os.Stderr
					}

					if err := cmd.Run(); err != nil {
						mu.Lock()
						errs = append(errs, fmt.Errorf("failed running go build for package %s: %w", p.Import, err))
						mu.Unlock()
					} else {
						log.Infof("Built package %s", p.Import)
					}
				}(p)
			}
		}

		wg.Wait()

		return errors.Join(errs...)
	}
}
