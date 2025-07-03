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

package js

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/codeclimate"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
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

// Run eslint for all passed jsPaths and returns a slice of CodeClimate-like entries
func Run(jsPaths []string, env environment.Environment) ([]codeclimate.Entry, error) {
	result := make([]codeclimate.Entry, 0, len(jsPaths))

	slog.Info("Running eslint")

	for _, path := range jsPaths {
		if entries, err := runEslint(path, env); err != nil {
			return result, fmt.Errorf("running eslint at %v: %w", path, err)
		} else {
			result = append(result, entries...)
		}
	}

	slog.Info("Completed eslint")

	return result, nil
}

// runEslint runs the actual yarn command and processes the raw output to a slice of CodeClimate-like entries
func runEslint(path string, env environment.Environment) ([]codeclimate.Entry, error) {
	var (
		lintEntries   []codeclimate.Entry
		esLintEntries []esLintEntry

		command     = "yarn"
		args        = []string{"run", "lint", "--format", "json"}
		result, err = cmdrunner.Run(command, args, path, env)
	)

	if err != nil {
		var errResult *cmdrunner.ExecutionResult

		if !errors.As(err, &errResult) {
			return lintEntries, fmt.Errorf("yarn run lint: %w", err)
		}

		result = errResult
	}

	if err := json.NewDecoder(result.StandardOutput).Decode(&esLintEntries); err != nil {
		return lintEntries, fmt.Errorf("yarn run lint: decoding output: %w", err)
	}

	for _, entry := range esLintEntries {
		for _, msg := range entry.Messages {
			var severity string

			switch msg.Severity {
			case 0:
				severity = "info"
			case 1:
				severity = "warning"
			case 2:
				severity = "error"
			default:
				return nil, fmt.Errorf("yarn run lint: unknown severity %d", msg.Severity)
			}

			ccEntry := codeclimate.Entry{
				Description: msg.RuleID + ": " + msg.Message,
				Severity:    severity,
				Location: codeclimate.Location{
					Path: entry.FilePath,
					Lines: codeclimate.Lines{
						Begin: msg.Line,
					},
				},
			}

			lintEntries = append(lintEntries, ccEntry)
		}
	}

	return lintEntries, nil
}
