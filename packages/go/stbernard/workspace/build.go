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
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// BuildGoMainPackages builds all main packages for a list of module paths
func BuildGoMainPackages(workRoot string, modPaths []string) error {
	var (
		errs     []error
		wg       sync.WaitGroup
		mu       sync.Mutex
		buildDir = filepath.Join(workRoot, "dist") + string(filepath.Separator)
	)

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(buildDir, modPath string) {
			defer wg.Done()
			if err := buildGoModuleMainPackages(buildDir, modPath); err != nil {
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
func buildGoModuleMainPackages(buildDir string, modPath string) error {
	var (
		wg   sync.WaitGroup
		errs []error
		mu   sync.Mutex
	)

	if packages, err := moduleListPackages(modPath); err != nil {
		return fmt.Errorf("failed to list module packages: %w", err)
	} else {
		for _, p := range packages {
			if p.Name == "main" && !strings.Contains(p.Dir, "plugin") {
				wg.Add(1)
				go func(p GoPackage) {
					defer wg.Done()
					cmd := exec.Command("go", "build", "-o", buildDir)
					cmd.Dir = p.Dir
					if err := cmd.Run(); err != nil {
						mu.Lock()
						errs = append(errs, fmt.Errorf("failed running go build for package %s: %w", p.Import, err))
						mu.Unlock()
					} else {
						slog.Info("Built package", "package", p.Import, "dir", p.Dir)
					}
				}(p)
			}
		}

		wg.Wait()

		return errors.Join(errs...)
	}
}
