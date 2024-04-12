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

package golang

import (
	"errors"
	"fmt"
	"sync"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// TidyModules runs go mod tidy for all module paths passed
// Do not use currently, since go mod tidy is not compatible with go workspaces out of the box
func TidyModules(modPaths []string, env environment.Environment) error {
	var (
		errs []error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(modPath string) {
			defer wg.Done()

			var (
				command = "go"
				args    = []string{"mod", "tidy"}
			)

			if err := cmdrunner.Run(command, args, modPath, env); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("go mod tidy in %s: %w", modPath, err))
				mu.Unlock()
			}
		}(modPath)
	}

	wg.Wait()

	return errors.Join(errs...)
}

// DownloadModules runs go mod download for all module paths passed
func DownloadModules(modPaths []string, env environment.Environment) error {
	var (
		errs []error
		wg   sync.WaitGroup
		mu   sync.Mutex
	)

	for _, modPath := range modPaths {
		wg.Add(1)
		go func(modPath string) {
			defer wg.Done()

			var (
				command = "go"
				args    = []string{"mod", "download"}
			)

			if err := cmdrunner.Run(command, args, modPath, env); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("go mod download in %s: %w", modPath, err))
				mu.Unlock()
			}
		}(modPath)
	}

	wg.Wait()

	return errors.Join(errs...)
}

// SyncWorkspace runs go work sync in the given directory with a given set of environment
// variables
func SyncWorkspace(cwd string, env environment.Environment) error {
	var (
		command = "go"
		args    = []string{"work", "sync"}
	)

	if err := cmdrunner.Run(command, args, cwd, env); err != nil {
		return fmt.Errorf("go work sync: %w", err)
	} else {
		return nil
	}
}
