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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/codeclimate"
)

var (
	ErrNonZeroExit = errors.New("non-zero exit status")
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

func Run(jsPaths []string, env []string) ([]codeclimate.Entry, error) {
	var (
		exitError error = nil
		result          = make([]codeclimate.Entry, 0, len(jsPaths))
	)

	log.Infof("Running eslint")

	for _, path := range jsPaths {
		entries, err := runEslint(path, env)
		if errors.Is(err, ErrNonZeroExit) {
			exitError = ErrNonZeroExit
		} else if err != nil {
			return result, fmt.Errorf("failed to run eslint at %v: %w", path, err)
		}
		result = append(result, entries...)
	}

	log.Infof("Completed eslint")

	return result, exitError
}

func runEslint(cwd string, env []string) ([]codeclimate.Entry, error) {
	var (
		result    []codeclimate.Entry
		rawResult []esLintEntry
		outb      bytes.Buffer
	)

	cmd := exec.Command("yarn", "run", "lint", "--format", "json")
	cmd.Env = env
	cmd.Dir = cwd
	cmd.Stdout = &outb

	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		err = ErrNonZeroExit
	} else if err != nil {
		return result, fmt.Errorf("unexpected failure: %w", err)
	}

	if err := json.NewDecoder(&outb).Decode(&rawResult); err != nil {
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

			result = append(result, ccEntry)
		}
	}

	return result, err
}
