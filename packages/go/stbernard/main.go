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

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func main() {
	if workfile, err := getGoWorkFileAbsPath(); err != nil {
		log.Fatalf("Could not get Go Workspace file path: %v", err)
	} else if modules, err := parseModuleAbsPaths(workfile); err != nil {
		log.Fatalf("Could not parse defined module paths in Go Workspace file: %v", err)
	} else {
		for _, module := range modules {
			if err := runGoModDownload(module); err != nil {
				log.Fatalf("Could not download module at %s: %v", module, err)
			}
		}

		log.Printf("Successfully downloaded all modules from %s", workfile)
	}
}

func getGoWorkFileAbsPath() (string, error) {
	if cwd, err := os.Getwd(); err != nil {
		return "", fmt.Errorf("could not get current working directory: %w", err)
	} else if _, err := os.Stat("go.work"); err != nil {
		return "", fmt.Errorf("go workspace file does not exist in current directory: %w", err)
	} else {
		return filepath.Join(cwd, "go.work"), nil
	}
}

func parseModuleAbsPaths(workfilePath string) ([]string, error) {
	// go.work files aren't particularly heavy, so we'll just read into memory
	if data, err := os.ReadFile(workfilePath); err != nil {
		return nil, fmt.Errorf("could not read go.work file: %w", err)
	} else if workfile, err := modfile.ParseWork(workfilePath, data, nil); err != nil {
		return nil, fmt.Errorf("could not parse go.work file: %w", err)
	} else {
		var (
			modulePaths = make([]string, 0, len(workfile.Use))
			workDir     = filepath.Dir(workfilePath)
		)

		for _, use := range workfile.Use {
			modulePaths = append(modulePaths, filepath.Join(workDir, use.Path))
		}

		return modulePaths, nil
	}
}

func runGoModDownload(modulePath string) error {
	log.Printf("Running go mod download for path: %s", modulePath)
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = modulePath
	return cmd.Run()
}
