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

package analyzers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/specterops/bloodhound/slices"
)

var (
	ErrSeverityExit = errors.New("high severity linter result")
	ErrNonZeroExit  = errors.New("non-zero exit status")
)

type esLintEntry struct {
	FilePath        string          `json:"filePath"`
	ErrorCount      int             `json:"errorCount"`
	WarningCount    int             `json:"warningCount"`
	FatalErrorCount int             `json:"fatalErrorCount"`
	Messages        []esLintMessage `json:"messages"`
}

type esLintMessage struct {
	RuleID   string `json:"ruleId"`
	Severity int    `json:"severity"`
	Message  string `json:"message"`
	Line     uint64 `json:"line"`
}

type codeClimateEntry struct {
	Description string              `json:"description"`
	Severity    string              `json:"severity"`
	Location    codeClimateLocation `json:"location"`
}

type codeClimateLocation struct {
	Path  string           `json:"path"`
	Lines codeClimateLines `json:"lines"`
}

type codeClimateLines struct {
	Begin uint64 `json:"begin"`
}

func Run(cwd string, modPaths []string, jsPaths []string, env []string) (string, error) {
	var (
		severityError bool
	)

	golint, err := runGolangcilint(cwd, modPaths, env)
	if errors.Is(err, ErrNonZeroExit) {
		log.Println("Ignoring golangci-lint exit code")
	} else if err != nil {
		return "", fmt.Errorf("golangci-lint: %w", err)
	}

	eslint, err := runAllEslint(jsPaths, env)
	if errors.Is(err, ErrNonZeroExit) {
		log.Println("Ignoring eslint exit code")
	} else if err != nil {
		return "", fmt.Errorf("eslint: %w", err)
	}

	codeClimateReport := append(golint, eslint...)

	for idx, entry := range codeClimateReport {
		// We're using err == nil here because we want to do nothing if an error occurs
		if path, err := filepath.Rel(cwd, entry.Location.Path); err == nil {
			codeClimateReport[idx].Location.Path = path
		}

		if entry.Severity == "error" {
			severityError = true
		}
	}

	if jsonBytes, err := json.MarshalIndent(codeClimateReport, "", "    "); err != nil {
		return "", fmt.Errorf("could not marshal code climate report: %w", err)
	} else if severityError {
		return string(jsonBytes), ErrSeverityExit
	} else {
		return string(jsonBytes), nil
	}
}

func runGolangcilint(cwd string, modPaths []string, env []string) ([]codeClimateEntry, error) {
	var (
		result []codeClimateEntry
		args   = []string{"run", "--out-format", "code-climate", "--config", ".golangci.json", "--"}
		outb   bytes.Buffer
	)

	args = append(args, slices.Map(modPaths, func(modPath string) string {
		return path.Join(modPath, "...")
	})...)

	cmd := exec.Command("golangci-lint")
	cmd.Env = env
	cmd.Dir = cwd
	cmd.Stdout = &outb
	cmd.Args = append(cmd.Args, args...)

	cmdErr := cmd.Run()
	if cmdErr != nil {
		cmdErr = ErrNonZeroExit
	}

	if err := json.NewDecoder(&outb).Decode(&result); err != nil {
		return result, fmt.Errorf("failed to decode output: %w", err)
	} else {
		return result, cmdErr
	}
}

func runAllEslint(jsPaths []string, env []string) ([]codeClimateEntry, error) {
	var (
		exitError error = nil
		result          = make([]codeClimateEntry, 0, len(jsPaths))
	)

	for _, path := range jsPaths {
		entries, err := runEslint(path, env)
		if errors.Is(err, ErrNonZeroExit) {
			exitError = ErrNonZeroExit
		} else if err != nil {
			return result, fmt.Errorf("failed to run eslint at %v: %w", path, err)
		}
		result = append(result, entries...)
	}

	return result, exitError
}

func runEslint(cwd string, env []string) ([]codeClimateEntry, error) {
	var (
		result    []codeClimateEntry
		rawResult []esLintEntry
		outb      bytes.Buffer
	)

	cmd := exec.Command("yarn", "run", "lint", "--format", "json", "--quiet")
	cmd.Env = env
	cmd.Dir = cwd
	cmd.Stdout = &outb

	cmdErr := cmd.Run()
	if cmdErr != nil {
		cmdErr = ErrNonZeroExit
	}

	err := json.NewDecoder(&outb).Decode(&rawResult)
	if err != nil {
		return result, fmt.Errorf("failed to decode output: %w", err)
	}

	for _, entry := range rawResult {
		for _, msg := range entry.Messages {
			var severity string

			switch msg.Severity {
			case 0:
				severity = "info"
			case 1:
				severity = "warning"
			case 2:
				severity = "error"
			}

			ccEntry := codeClimateEntry{
				Description: msg.RuleID + ": " + msg.Message,
				Severity:    severity,
				Location: codeClimateLocation{
					Path: entry.FilePath,
					Lines: codeClimateLines{
						Begin: msg.Line,
					},
				},
			}

			result = append(result, ccEntry)
		}
	}

	return result, cmdErr
}
