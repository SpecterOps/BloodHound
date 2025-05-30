// Copyright Specter Ops, Inc.
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
// 
package license

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

func Run(env environment.Environment) error {
	var (
		ignoreDir            = []string{".git", ".vscode", ".devcontainer", "node_modules", "dist", ".beagle", ".yarn", "sha256"}
		ignorePaths          = []string{"tools/docker-compose/configs/pgadmin/pgpass", "justfile", "cmd/api/src/api/static/assets", "packages/python/beagle/beagle/semver", "cmd/api/src/cmd/testidp/samlidp"}
		disallowedExtensions = []string{".zip", ".example", ".git", ".gitignore", ".png", ".mdx", ".iml", ".g4", ".sum", ".bazel", ".bzl", ".typed", ".md", ".json", ".template", "sha256", ".pyc", ".gif", ".tiff", ".lock", ".txt", ".png", ".jpg", ".jpeg", ".ico", ".gz", ".tar", ".woff2", ".header", ".pro", ".cert", ".crt", ".key", ".example", ".svg", ".sha256"}
		now                  = time.Now()

		// Concurrency primitives
		errs       []error
		wg         = &sync.WaitGroup{}
		errsMu     = &sync.Mutex{}
		numWorkers = runtime.NumCPU()
		pathChan   = make(chan string, numWorkers)
	)

	wrkPaths, err := workspace.FindPaths(env)
	if err != nil {
		return fmt.Errorf("failed to find workspace paths: %w", err)
	}

	licensePath := filepath.Join(wrkPaths.Root, "LICENSE")

	// Make sure root LICENSE FILE exists
	if _, err := os.Stat(licensePath); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(licensePath, []byte(licenseContent), 0644); err != nil {
			return fmt.Errorf("failed to write root license file: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check for root license file: %w", err)
	}

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for path := range pathChan {
				result, err := isHeaderPresent(path)
				if err != nil {
					errsMu.Lock()
					errs = append(errs, err)
					errsMu.Unlock()
					return
				}

				if !result {
					if err := processFile(path); err != nil {
						errsMu.Lock()
						errs = append(errs, err)
						errsMu.Unlock()
					}
				}
			}
		}()
	}

	err = filepath.Walk(wrkPaths.Root, func(path string, info fs.FileInfo, err error) error {
		// ignore directories and paths that are in the ignore list
		scanPath := shouldProcessPath(ignorePaths, path)
		if info.IsDir() && (slices.Contains(ignoreDir, info.Name()) || !scanPath) {
			return filepath.SkipDir
		}

		// ignore files that are in the ignore list
		ext := filepath.Ext(path)
		if !info.IsDir() && !slices.Contains(disallowedExtensions, ext) {
			pathChan <- path
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path: %w", err)
	}

	// close path channel to signal we're done sending values
	close(pathChan)
	// block main until all the goroutines in done state
	wg.Wait()
	diff := time.Since(now)

	slog.Info("running scans on bhce", slog.Duration("execution_time", diff))
	return errors.Join(errs...)
}

func processFile(path string) error {
	switch filepath.Ext(path) {
	case ".go", ".work", ".mod", ".ts", ".tsx", ".js", ".cjs", ".jsx", ".cue", ".scss":
		h := generateLicenseHeader("//")
		return writeFile(path, h)
	case ".yaml", ".yml", ".py", ".ssh", ".Dockerfile", ".toml":
		h := generateLicenseHeader("#")
		return writeFile(path, h)
	case ".sql":
		h := generateLicenseHeader("--")
		return writeFile(path, h)
	case ".xml", ".html":
		h := generateXMLLicenseHeader()
		return writeXMLFile(path, h)
	default:
		fmt.Printf("unknown extension file: %v\n", path)
		return nil
	}
}
