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

package golang

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/git"
)

// BuildMainPackages builds all main packages for a list of module paths
func BuildMainPackages(workRoot string, modPaths []string, env environment.Environment) error {
	var (
		err      error
		errs     []error
		wg       sync.WaitGroup
		mu       sync.Mutex
		version  semver.Version
		buildDir = filepath.Join(workRoot, "dist") + string(filepath.Separator)
	)

	if version, err = git.ParseLatestVersionFromTags(workRoot, env); err != nil {
		log.Warnf("Failed to parse version from git tags, falling back to environment variable: %v", err)
		parsedVersion, err := semver.NewVersion(env[environment.VersionVarName])
		if err != nil {
			return fmt.Errorf("error parsing version from environment variable: %w", err)
		}
		version = *parsedVersion
	}

	log.Infof("Building for version %s", version.Original())

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(buildDir, modPath string) {
			defer wg.Done()
			if err := buildModuleMainPackages(buildDir, modPath, version, env); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("build main package: %w", err))
				mu.Unlock()
			}
		}(buildDir, modPath)
	}

	wg.Wait()

	return errors.Join(errs...)
}

func buildModuleMainPackages(buildDir string, modPath string, version semver.Version, env environment.Environment) error {
	var (
		wg   sync.WaitGroup
		errs []error
		mu   sync.Mutex

		command             = "go"
		majorString         = fmt.Sprintf("-X 'github.com/specterops/bloodhound/src/version.majorVersion=%d'", version.Major())
		minorString         = fmt.Sprintf("-X 'github.com/specterops/bloodhound/src/version.minorVersion=%d'", version.Minor())
		patchString         = fmt.Sprintf("-X 'github.com/specterops/bloodhound/src/version.patchVersion=%d'", version.Patch())
		prereleaseString    = fmt.Sprintf("-X 'github.com/specterops/bloodhound/src/version.prereleaseVersion=%s'", version.Prerelease())
		ldflagArgComponents = []string{majorString, minorString, patchString}
	)

	if version.Prerelease() != "" {
		ldflagArgComponents = append(ldflagArgComponents, prereleaseString)
	}

	args := []string{"build", "-ldflags", strings.Join(ldflagArgComponents, " "), "-o", buildDir}

	if packages, err := moduleListPackages(modPath); err != nil {
		return fmt.Errorf("list module packages: %w", err)
	} else {
		for _, pkg := range packages {
			if pkg.Name == "main" && !strings.Contains(pkg.Dir, "plugin") {
				wg.Add(1)
				go func(p GoPackage) {
					defer wg.Done()

					if err := cmdrunner.Run(command, args, p.Dir, env); err != nil {
						mu.Lock()
						errs = append(errs, fmt.Errorf("go build for package %s: %w", p.Import, err))
						mu.Unlock()
					}

					log.Infof("Built package %s", p.Import)
				}(pkg)
			}
		}

		wg.Wait()

		return errors.Join(errs...)
	}
}
