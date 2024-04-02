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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/slicesext"
	"golang.org/x/mod/modfile"
)

const (
	// CoverageManifest is the manifest filename
	CoverageManifest = "manifest.json"
	// CoverageExt is the file extension used by coverage files
	CoverageExt = ".coverage"
	// CombinedCoverage is the combined coverage report filename
	CombinedCoverage = "combined" + CoverageExt
)

var (
	ErrTotalCoverageNotFound = errors.New("total coverage not found")
)

var (
	// Default path to store Go coverage
	DefaultCoveragePath          = filepath.Join("tmp", "coverage")
	defaultIntegrationConfigPath = filepath.Join("local-harnesses", "integration.config.json")
)

// TestWorkspace runs all Go tests for a given workspace. Setting integration to true will run integration tests, otherwise we only run unit tests
func TestWorkspace(cwd string, modPaths []string, profileDir string, env environment.Environment, integration bool) error {
	var (
		manifest = make(map[string]string, len(modPaths))
		command  = "go"
		args     = []string{"test"}
	)

	if integration {
		if integrationConfigPath, ok := env["INTEGRATION_CONFIG_PATH"]; !ok || integrationConfigPath == "" {
			env["INTEGRATION_CONFIG_PATH"] = filepath.Join(cwd, defaultIntegrationConfigPath)
		} else if !filepath.IsAbs(integrationConfigPath) {
			env["INTEGRATION_CONFIG_PATH"] = filepath.Join(cwd, integrationConfigPath)
		}

		args = append(args, []string{"-p", "1", "-tags", "integration serial_integration"}...)
	}

	for _, modPath := range modPaths {
		modName, err := getModuleName(modPath)
		if err != nil {
			return err
		}

		fileUUID, err := uuid.NewV4()
		if err != nil {
			return fmt.Errorf("create uuid for coverfile: %w", err)
		}

		coverFile := filepath.Join(profileDir, fileUUID.String()+".coverage")
		manifest[modName] = coverFile
		testArgs := slicesext.Concat(args, []string{"-coverprofile", coverFile, "./..."})

		if err := cmdrunner.Run(command, testArgs, modPath, env); err != nil {
			return fmt.Errorf("go test at %v: %w", modPath, err)
		}
	}

	if manifestFile, err := os.Create(filepath.Join(profileDir, CoverageManifest)); err != nil {
		return fmt.Errorf("create manifest file: %w", err)
	} else if marshaledManifest, err := json.Marshal(manifest); err != nil {
		manifestFile.Close()
		return fmt.Errorf("marshal manifest: %w", err)
	} else if _, err := manifestFile.Write(marshaledManifest); err != nil {
		manifestFile.Close()
		return fmt.Errorf("writing manifest to file: %w", err)
	} else if err := manifestFile.Close(); err != nil {
		return fmt.Errorf("closing manifest file: %w", err)
	} else {
		return nil
	}
}

// GetCombinedCoverage takes a coverage file and returns a string representation of percentage of statements covered
func GetCombinedCoverage(coverFile string, env environment.Environment) (string, error) {
	var (
		output bytes.Buffer

		args       = []string{"tool", "cover", "-func", filepath.Base(coverFile)}
		channelOut = func(c *exec.Cmd) {
			c.Stdout = &output
		}
	)

	if err := cmdrunner.Run("go", args, filepath.Dir(coverFile), env, channelOut); err != nil {
		return "", fmt.Errorf("combined coverage: %w", err)
	} else if re, err := regexp.Compile(`total:\s+\(.*?\)\s+(\d+(?:.\d+)?%)`); err != nil {
		return "", fmt.Errorf("regex failed to compile: %w", err)
	} else {
		matches := re.FindStringSubmatch(output.String())

		// This regex has only one capture group, so we expect the percentage to be in the capture group portion of the matches
		// There should be two matches since the first match result is the full string that was matched, and the second is the result of our capture group
		if len(matches) == 2 {
			return matches[1], nil
		}

		return "", ErrTotalCoverageNotFound
	}
}

func getModuleName(modPath string) (string, error) {
	if modFile, err := os.ReadFile(filepath.Join(modPath, "go.mod")); err != nil {
		return "", fmt.Errorf("reading go.mod file for %s: %w", modPath, err)
	} else {
		return modfile.ModulePath(modFile), nil
	}
}
