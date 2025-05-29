//  Copyright 2025 Specter Ops, Inc.
//
//  Licensed under the Apache License, Version 2.0
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.
//
//  SPDX-License-Identifier: Apache-2.0
//

package license

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"
)

func Run() error {

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}
	if err := checkRootFiles(wd); err != nil {
		return fmt.Errorf("failed to check root files: %w", err)
	}

	ignoreDir := []string{".git", ".vscode", ".devcontainer", "node_modules", "dist", ".beagle", ".yarn", "sha256"}
	ignorPaths := []string{"tools/docker-compose/configs/pgadmin/pgpass", "justfile", "cmd/api/src/api/static/assets", "packages/python/beagle/beagle/semver", "cmd/api/src/cmd/testidp/samlidp"}
	disallowedExtensions := []string{".zip", ".example", ".git", ".gitignore", ".png", ".mdx", ".iml", ".g4", ".sum", ".bazel", ".bzl", ".typed", ".md", ".json", ".template", "sha256", ".pyc", ".gif", ".tiff", ".lock", ".txt", ".png", ".jpg", ".jpeg", ".ico", ".gz", ".tar", ".woff2", ".header", ".pro", ".cert", ".crt", ".key", ".example", ".svg", ".sha256"}

	paths := []string{}

	now := time.Now()
	err = filepath.Walk(wd, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			// Log the error and continue walking
			slog.Warn(fmt.Sprintf("error accessing path %q: %v", path, err))
			return nil // Continue walking despite the error
		}

		// ignore directories
		if info.IsDir() && slices.Contains(ignoreDir, info.Name()) {
			return filepath.SkipDir
		}

		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path: %w", err)
	}

	// append all the errors from the goroutines
	var (
		errs   []error
		errsMu sync.Mutex
	)
	var wg sync.WaitGroup

	// Use a worker pool pattern with limited concurrency
	const maxWorkers = 10
	sem := make(chan struct{}, maxWorkers)

	for _, path := range paths {
		// series of validations before processing all the paths against directories, disallowed extensions and paths
		ext := parseFileExtension(path)
		isDir := dirCheck(path)
		scanPath := shouldProcessPath(ignorPaths, wd, path)

		if !slices.Contains(disallowedExtensions, ext) && !isDir && len(ext) != 0 && scanPath {
			wg.Add(1)
			go func(filePath, fileExt string) {
				defer wg.Done()
				sem <- struct{}{}        // acquire semaphore
				defer func() { <-sem }() // release semaphore

				result, err := isHeaderPresent(filePath)
				if err != nil {
					errsMu.Lock()
					errs = append(errs, err)
					errsMu.Unlock()
					return
				}

				if !result {
					if err := processFile(filePath, fileExt); err != nil {
						errsMu.Lock()
						errs = append(errs, err)
						errsMu.Unlock()
					}
				}
			}(path, ext)
		}
	}

	// block main untill all the goroutines in done state
	wg.Wait()
	diff := time.Since(now)
	
	slog.Info(fmt.Sprintf("running scans on bhce took %v", diff))
	return errors.Join(errs...)
}

func processFile(path, ext string) error {
	switch ext {
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
