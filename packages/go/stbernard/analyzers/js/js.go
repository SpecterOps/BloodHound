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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/codeclimate"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

type oxlintOutput struct {
	Diagnostics []oxlintDiagnostic `json:"diagnostics"`
}

type eslintEntry struct {
	FilePath string `json:"filePath"`
	Messages []struct {
		RuleID   string `json:"ruleId"`
		Severity int    `json:"severity"`
		Message  string `json:"message"`
		Line     uint64 `json:"line"`
	} `json:"messages"`
}

type oxlintDiagnostic struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Filename string `json:"filename"`
	Labels   []struct {
		Span struct {
			Line uint64 `json:"line"`
		} `json:"span"`
	} `json:"labels"`
}

// Run each workspace's lint script and collect CodeClimate entries.
func Run(jsPaths []string, env environment.Environment) (codeclimate.SeverityMap, error) {
	var result = make([]codeclimate.Entry, 0, len(jsPaths))

	slog.Info("Running JavaScript lint")
	for _, path := range jsPaths {
		if entries, err := runLint(path, env); err != nil {
			return nil, fmt.Errorf("running lint at %v: %w", path, err)
		} else {
			result = append(result, entries...)
		}
	}
	slog.Info("Completed JavaScript lint")
	return codeclimate.NewSeverityMap(result), nil
}

func runLint(path string, env environment.Environment) ([]codeclimate.Entry, error) {
	var executionPlan = cmdrunner.ExecutionPlan{
		Command:        env[environment.YarnCmdVarName],
		Args:           []string{"run", "lint", "--format", "json"},
		Path:           path,
		Env:            env.Slice(),
		SuppressErrors: true,
	}

	result, err := cmdrunner.Run(context.TODO(), executionPlan)
	if err != nil && !errors.Is(err, cmdrunner.ErrCmdExecutionFailed) {
		return nil, fmt.Errorf("yarn run lint: %w", err)
	}
	entries, decodeError := decodeDiagnostics(result.StandardOutput, path)
	if decodeError != nil {
		return nil, fmt.Errorf("decoding lint output: %w; stderr: %s", decodeError, result.ErrorOutput.String())
	}
	// Exit 1 with diagnostics is an ordinary lint failure. Empty output or other
	// exit codes indicate a tooling failure and must not be treated as a clean run.
	if err != nil && (result.ReturnCode != 1 || len(entries) == 0) {
		return nil, fmt.Errorf("lint failed: %w; stderr: %s", err, result.ErrorOutput.String())
	}
	return entries, nil
}

func decodeDiagnostics(reader io.Reader, workspacePath string) ([]codeclimate.Entry, error) {
	var (
		rawOutput json.RawMessage
		output    oxlintOutput
		entries   []codeclimate.Entry
	)
	if err := json.NewDecoder(reader).Decode(&rawOutput); err != nil {
		return nil, err
	}
	// Keep supporting ESLint workspaces while BHE and BHCE migrate separately.
	if len(rawOutput) > 0 && rawOutput[0] == '[' {
		return decodeESLintDiagnostics(rawOutput)
	}
	if err := json.Unmarshal(rawOutput, &output); err != nil {
		return nil, err
	}
	if output.Diagnostics == nil {
		return nil, errors.New("missing diagnostics array")
	}
	for _, diagnostic := range output.Diagnostics {
		var (
			severity codeclimate.Severity
			line     uint64 = 1
			path            = diagnostic.Filename
		)
		switch diagnostic.Severity {
		case "advice":
			severity = codeclimate.SeverityInfo
		case "warning":
			severity = codeclimate.SeverityMinor
		case "error":
			severity = codeclimate.SeverityMajor
		default:
			return nil, fmt.Errorf("unknown severity %q", diagnostic.Severity)
		}
		for _, label := range diagnostic.Labels {
			if label.Span.Line > 0 {
				line = label.Span.Line
				break
			}
		}
		if path == "" {
			return nil, errors.New("diagnostic missing filename")
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(workspacePath, path)
		}
		entries = append(entries, codeclimate.Entry{
			Description: diagnostic.Code + ": " + diagnostic.Message,
			Severity:    severity,
			Location: codeclimate.Location{
				Path:  path,
				Lines: codeclimate.Lines{Begin: line},
			},
		})
	}
	return entries, nil
}

func decodeESLintDiagnostics(rawOutput json.RawMessage) ([]codeclimate.Entry, error) {
	var (
		output     []eslintEntry
		entries    []codeclimate.Entry
		severities = [...]codeclimate.Severity{
			codeclimate.SeverityInfo, codeclimate.SeverityMinor, codeclimate.SeverityMajor,
		}
	)
	if err := json.Unmarshal(rawOutput, &output); err != nil {
		return nil, err
	}
	for _, file := range output {
		for _, message := range file.Messages {
			if message.Severity < 0 || message.Severity > 2 {
				return nil, fmt.Errorf("unknown ESLint severity %d", message.Severity)
			}
			entries = append(entries, codeclimate.Entry{
				Description: message.RuleID + ": " + message.Message,
				Severity:    severities[message.Severity],
				Location: codeclimate.Location{
					Path:  file.FilePath,
					Lines: codeclimate.Lines{Begin: message.Line},
				},
			})
		}
	}
	return entries, nil
}
