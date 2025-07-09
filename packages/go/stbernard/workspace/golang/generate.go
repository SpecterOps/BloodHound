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
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/dawgs/util/channels"
)

// fileContentContainsGenerationDirective uses a bufio.Scanner to search a given reader line-by-line
// looking for golang code generation directives. Upon finding the first code generation directive
// this function returns true with no error. If no code generation directives exist this function will
// return false instead.
func fileContentContainsGenerationDirective(fin io.Reader) (bool, error) {
	scanner := bufio.NewScanner(fin)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "//go:generate") {
			// If we find a go generate directive return right away
			return true, nil
		}
	}

	return false, scanner.Err()
}

// packageHasGenerationDirectives scans a golang package at a given path for any files that contain a
// golang code generation directive. Upon finding any file in the package that contains a code
// generation directive, this function returns true with no error. If no code generation directives exist
// in the package, this function returns false instead.
func packageHasGenerationDirectives(packagePath string) (bool, error) {
	hasGolangCodeGenDirectives := false

	if err := filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
		// Don't bother reading anything that isn't a golang source file
		if info.IsDir() || filepath.Ext(info.Name()) != ".go" {
			return nil
		}

		// Open the file and search for code generation directives
		if fin, err := os.Open(path); err != nil {
			return err
		} else {
			defer fin.Close()

			if hasGolangCodeGenDirectives, err = fileContentContainsGenerationDirective(fin); err != nil {
				return err
			} else if hasGolangCodeGenDirectives {
				// Skip the rest of the FS walk for this package
				return filepath.SkipAll
			}
		}

		return nil
	}); err != nil {
		return false, err
	}

	return hasGolangCodeGenDirectives, nil
}

// parallelGenerateModulePackages spins up runtime.NumCPU() concurrent workers that will attempt golang code generation
// for each GoPackage transmitted over the jobC channel.
func parallelGenerateModulePackages(jobC <-chan GoPackage, waitGroup *sync.WaitGroup, env environment.Environment, addErr func(err error)) {
	for workerID := 1; workerID <= runtime.NumCPU(); workerID++ {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			for {
				if nextPackage, canContinue := channels.Receive(context.TODO(), jobC); !canContinue {
					break
				} else if hasGenerationDirectives, err := packageHasGenerationDirectives(nextPackage.Dir); err != nil {
					addErr(err)
				} else if hasGenerationDirectives {
					var (
						command = "go"
						args    = []string{"generate", nextPackage.Dir}
					)

					if _, err := cmdrunner.RunInteractive(command, args, nextPackage.Dir, env); err != nil {
						addErr(err)
					}
				}
			}
		}()
	}
}

// WorkspaceGenerate runs go generate ./... for all module paths passed
func WorkspaceGenerate(modPaths []string, env environment.Environment) error {
	defer measure.LogAndMeasure(slog.LevelDebug, "WorkspaceGenerate")()
	var (
		errs     []error
		errsLock = &sync.Mutex{}
		addErr   = func(err error) {
			errsLock.Lock()
			defer errsLock.Unlock()

			errs = append(errs, err)
		}

		jobC      = make(chan GoPackage)
		waitGroup = &sync.WaitGroup{}
	)

	// Start the parallel workers first
	go parallelGenerateModulePackages(jobC, waitGroup, env, addErr)

	// For each known module path attempt generation of each module package
	for _, modPath := range modPaths {
		if modulePackages, err := moduleListPackages(modPath); err != nil {
			return fmt.Errorf("getting module packages for %s: %w", modPath, err)
		} else {
			for _, modulePackage := range modulePackages {
				if !channels.Submit(context.Background(), jobC, modulePackage) {
					return fmt.Errorf("canceled")
				}
			}
		}
	}

	close(jobC)

	waitGroup.Wait()

	return errors.Join(errs...)
}
