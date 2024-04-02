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
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// ParseModulesAbsPaths parses the modules listed in the go.work file from the given
// directory and returns a list of absolute paths to those modules
func ParseModulesAbsPaths(cwd string) ([]string, error) {
	var workfilePath = filepath.Join(cwd, "go.work")
	// go.work files aren't particularly heavy, so we'll just read into memory
	if data, err := os.ReadFile(workfilePath); err != nil {
		return nil, fmt.Errorf("reading go.work file: %w", err)
	} else if workfile, err := modfile.ParseWork(workfilePath, data, nil); err != nil {
		return nil, fmt.Errorf("parsing go.work file: %w", err)
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

func moduleListPackages(modPath string) ([]GoPackage, error) {
	var (
		packages = make([]GoPackage, 0)
	)

	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = modPath
	if out, err := cmd.StdoutPipe(); err != nil {
		return packages, fmt.Errorf("creating stdout pipe for module %s: %w", modPath, err)
	} else if err := cmd.Start(); err != nil {
		return packages, fmt.Errorf("listing packages for module %s: %w", modPath, err)
	} else {
		decoder := json.NewDecoder(out)
		for {
			var p GoPackage
			if err := decoder.Decode(&p); err == io.EOF {
				break
			} else if err != nil {
				return packages, fmt.Errorf("decoding package in module %s: %w", modPath, err)
			}
			packages = append(packages, p)
		}
		cmd.Wait()
		return packages, nil
	}
}
