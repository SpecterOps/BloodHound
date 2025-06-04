// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
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
	"strings"
	"sync"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

func Run(env environment.Environment) error {
	var (
		ignoreDir   = []string{".git", ".vscode", ".devcontainer", "node_modules", "dist", ".beagle", ".yarn", "sha256"}
		ignorePaths = []string{
			filepath.Join("tools", "docker-compose", "configs", "pgadmin", "pgpass"),
			"justfile",
			filepath.Join("cmd", "api", "src", "api", "static", "assets"),
			filepath.Join("packages", "python", "beagle", "beagle", "semver"),
			filepath.Join("cmd", "api", "src", "cmd", "testidp", "samlidp"),
		}
		disallowedExtensions = []string{".zip", ".example", ".git", ".gitignore", ".gitattributes", ".png", ".mdx", ".iml", ".g4", ".sum", ".bazel", ".bzl", ".typed", ".md", ".json", ".template", "sha256", ".pyc", ".gif", ".tiff", ".lock", ".txt", ".png", ".jpg", ".jpeg", ".ico", ".gz", ".tar", ".woff", ".woff2", ".header", ".pro", ".cert", ".crt", ".key", ".example", ".sha256", ".actrc", ".all-contributorsrc", ".editorconfig", ".conf", ".dockerignore", ".prettierrc", ".lintstagedrc", ".webp", ".bak", ".java", ".interp", ".tokens", "justfile", "pgpass", "LICENSE"}
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
	} else if err := writeLicenseHeaderFile(licensePath); err != nil {
		return fmt.Errorf("failed to write LICENSE.header: %w", err)
	}

	// worker pool pattern for handling files
	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Ranging over a channel will end when the channel is closed, so this is a nice simplification
			for path := range pathChan {
				result, err := isHeaderPresent(path)
				if err != nil {
					errsMu.Lock()
					errs = append(errs, err)
					errsMu.Unlock()
					// explicitly continue to process files since we want to process as much as possible even if some errors occur
					continue
				}

				if !result {
					if err := processFile(path); err != nil {
						errsMu.Lock()
						errs = append(errs, err)
						errsMu.Unlock()
						// explicitly continue to process files since we want to process as much as possible even if some errors occur
						continue
					}
				}
			}
		}()
	}

	err = filepath.Walk(wrkPaths.Root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path for consistent matching
		relPath, err := filepath.Rel(wrkPaths.Root, path)
		if err != nil {
			return err
		}

		// Check if the current path contains one of our ignored paths
		ignorePath := slices.ContainsFunc(ignorePaths, func(igPath string) bool {
			// Use HasPrefix for more precise matching
			return strings.HasPrefix(relPath, igPath) || relPath == igPath
		})

		// ignore directories and paths that are in the ignore list
		if info.IsDir() && (slices.Contains(ignoreDir, info.Name()) || ignorePath) {
			return filepath.SkipDir
		}

		// ignore files that are in the ignore list
		ext := filepath.Ext(path)
		// if there is no extension, use the filename as the extension
		if ext == "" {
			ext = filepath.Base(path)
		}
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
		return writeFile(path, generateLicenseHeader("//"))
	case ".yaml", ".yml", ".py", ".ssh", ".Dockerfile", ".toml":
		return writeFile(path, generateLicenseHeader("#"))
	case ".sql":
		return writeFile(path, generateLicenseHeader("--"))
	case ".xml", ".html", ".svg":
		return writeFile(path, generateLicenseHeader("<!--"))
	case ".css":
		return writeFile(path, generateLicenseHeader("/*"))
	default:
		slog.Warn("unknown extension", slog.String("path", path))
		return nil
	}
}

func writeLicenseHeaderFile(licensePath string) error {
	formattedHeader := generateLicenseHeader("")
	// Cut final \n for better formatting with MockGen
	formattedHeader, _ = strings.CutSuffix(formattedHeader, "\n")
	return os.WriteFile(licensePath+".header", []byte(formattedHeader), 0644)
}
