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
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// ParseModulesAbsPaths walks the filesystem looking for additional go.mod files as children of the workspace root
func ParseModulesAbsPaths(cwd string) ([]string, error) {
	var modules = make([]string, 0, 4)

	slog.Info("found go module", slog.String("path", cwd))
	modules = append(modules, cwd)

	err := filepath.WalkDir(cwd, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking directories for go modules: %w", err)
		}

		// Skip hidden files and directories
		if entry.IsDir() && strings.HasPrefix(entry.Name(), ".") {
			return filepath.SkipDir
		}

		if path != filepath.Join(cwd, "go.mod") && entry.Name() == "go.mod" {
			slog.Info("found go module", slog.String("path", path))
			absPath, err := filepath.Abs(filepath.Dir(path))
			if err != nil {
				return fmt.Errorf("absolute path for discovered module: %w", err)
			}
			modules = append(modules, absPath)
		}

		return nil
	})

	if err != nil {
		return modules, fmt.Errorf("parsing modules absolute paths: %w", err)
	}

	return modules, nil
}

func moduleListPackages(modPath string) ([]GoPackage, error) {
	if result, err := cmdrunner.Run("go", []string{"list", "-json", "./..."}, modPath, environment.NewEnvironment()); err != nil {
		return nil, fmt.Errorf("running go mod list: %w", err)
	} else {
		var (
			decoder     = json.NewDecoder(result.StandardOutput)
			packages    []GoPackage
			nextPackage GoPackage
		)

		for {
			if err := decoder.Decode(&nextPackage); err == io.EOF {
				break
			} else if err != nil {
				return nil, fmt.Errorf("decoding package in module %s: %w", modPath, err)
			}

			packages = append(packages, nextPackage)
		}

		return packages, nil
	}
}
