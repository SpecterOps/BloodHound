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
	"sync"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// WorkspaceGenerate runs go generate ./... for all module paths passed
func WorkspaceGenerate(modPaths []string, env environment.Environment) error {
	var (
		errs []error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(modPath string) {
			defer wg.Done()
			if err := moduleGenerate(modPath, env); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("code generation for module %s: %w", modPath, err))
				mu.Unlock()
			}
		}(modPath)
	}

	wg.Wait()

	return errors.Join(errs...)
}

func moduleGenerate(modPath string, env environment.Environment) error {
	var (
		errs []error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)

	if packages, err := moduleListPackages(modPath); err != nil {
		return fmt.Errorf("listing packages for module %s: %w", modPath, err)
	} else {
		for _, pkg := range packages {
			wg.Add(1)
			go func(pkg GoPackage) {
				defer wg.Done()

				var (
					command = "go"
					args    = []string{"generate", pkg.Dir}
				)

				if err := cmdrunner.Run(command, args, pkg.Dir, env); err != nil {
					mu.Lock()
					errs = append(errs, fmt.Errorf("generate code for package %s: %w", pkg, err))
					mu.Unlock()
				}
			}(pkg)
		}

		wg.Wait()

		return errors.Join(errs...)
	}
}
