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

package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	license "github.com/specterops/bloodhound/packages/go/stbernard/command/license/internal"
)

func main() {

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("checked failed license root files error: %v", err)
	}
	if err := license.CheckRootFiles(wd); err != nil {
		fmt.Printf("checked failed license root files error: %v", err)
	}

	ignoreDir := []string{".git", ".vscode", ".devcontainer", "node_modules", "dist", ".beagle", ".yarn", "sha256"}
	ignorPaths := []string{"tools/docker-compose/configs/pgadmin/pgpass", "justfile", "cmd/api/src/api/static/assets", "packages/python/beagle/beagle/semver", "cmd/api/src/cmd/testidp/samlidp"}
	disallowedExtensions := []string{".zip", ".example", ".git", ".gitignore", ".png", ".mdx", ".iml", ".g4", ".sum", ".bazel", ".bzl", ".typed", ".md", ".json", ".template", "sha256", ".pyc", ".gif", ".tiff", ".lock", ".txt", ".png", ".jpg", ".jpeg", ".ico", ".gz", ".tar", ".woff2", ".header", ".pro", ".cert", ".crt", ".key", ".example", ".svg", ".sha256"}

	paths := []string{}

	now := time.Now()
	err = filepath.Walk(wd, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		// ignore directories
		if info.IsDir() && slices.Contains(ignoreDir, info.Name()) {
			return filepath.SkipDir
		}

		paths = append(paths, path)
		return nil
	})

	// append all the errors from the goroutines
	var errs []error
	var wg sync.WaitGroup
	for _, path := range paths {
		// buffered channel to process tasks in FIFO
		task := make(chan string, len(paths))

		// series of validations before processing all the paths against directories, disallowed extensions and paths
		ext := license.ParseFileExtension(path)
		isDir := license.DirCheck(path)
		scanPath := license.IgnorePathValidation(ignorPaths, wd, path)

		if !slices.Contains(disallowedExtensions, ext) && !isDir && len(ext) != 0 && scanPath {
			// worker updates the task channel with one of two values. "skip" when the header is present
			// or a file path to be consumed by another worker to format the file
			wg.Add(1)
			go func() {
				defer wg.Done()

				result, err := license.IsHeaderPresent(path)
				if err != nil {
					err := fmt.Errorf("failed checking license header")
					errs = append(errs, err)
				}

				if !result {
					task <- path
				} else {
					task <- "skip"
				}
			}()

			// worker readings from the task channel and determines whether to format the file or skip the operation
			wg.Add(1)
			go func() {

				defer wg.Done()
				result := <-task

				if result != "skip" {
					switch ext {
					case ".go", ".work", ".mod", ".ts", ".tsx", ".js", ".cjs", ".cue", ".scss":
						h := license.GenerateLicenseHeader("//")
						if err := license.WriteFile(path, h); err != nil {
							err := fmt.Errorf("failed to append license header: %s", path)
							errs = append(errs, err)
						}
					case ".jsx", ".yaml", ".yml", ".py", ".ssh", ".Dockerfile", ".toml":
						h := license.GenerateLicenseHeader("#")
						if err := license.WriteFile(path, h); err != nil {
							err := fmt.Errorf("failed to append license header: %s", path)
							errs = append(errs, err)
						}
					case ".sql":
						h := license.GenerateLicenseHeader("--")
						if err := license.WriteFile(path, h); err != nil {
							err := fmt.Errorf("failed to append license header: %s", path)
							errs = append(errs, err)
						}
					case ".xml", ".html":
						h := license.GenerateXMLLicenseHeader()
						if err := license.WriteXMLFile(path, h); err != nil {
							err := fmt.Errorf("failed to append license header: %s", path)
							errs = append(errs, err)
						}
					default:
						fmt.Printf("unknown extension file: %v\n", path)
					}
				}
			}()
		}
	}

	if err != nil {
		fmt.Printf("error walking the path: %v\n", err)
	}

	tErr := errors.Join(errs...)
	if tErr != nil {
		fmt.Printf("processing errors %s\n", fmt.Sprint(tErr))
	}

	// block main untill all the goroutines in done state
	wg.Wait()
	diff := time.Since(now)
	fmt.Printf("running scans on bhce took %v\n", diff)
}
