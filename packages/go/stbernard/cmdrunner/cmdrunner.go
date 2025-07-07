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

package cmdrunner

import (
	"bytes"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// ExecutionResult data structure that represents the result of running a command and captures information about
// an executed command's output.
//
// The combined output buffer preserves the chronological order of both streams at time of collection from their
// respective file descriptors.
type ExecutionResult struct {
	Command        string
	Arguments      []string
	Path           string
	StandardOutput *bytes.Buffer
	ErrorOutput    *bytes.Buffer
	CombinedOutput *bytes.Buffer
	ReturnCode     int
}

func (s *ExecutionResult) Error() string {
	return "command execution failed: " + s.Command
}

func newExecutionResult(command string, args []string, path string) *ExecutionResult {
	return &ExecutionResult{
		Command:        command,
		Arguments:      args,
		Path:           path,
		StandardOutput: &bytes.Buffer{},
		ErrorOutput:    &bytes.Buffer{},
		CombinedOutput: &bytes.Buffer{},
		ReturnCode:     0,
	}
}

func prepareCommand(command string, args []string, path string, env environment.Environment) (*exec.Cmd, *ExecutionResult) {
	var (
		cmd    = exec.Command(command, args...)
		result = newExecutionResult(command, args, path)
	)

	cmd.Dir = path
	cmd.Env = env.Slice()
	cmd.Stdout = io.MultiWriter(result.StandardOutput, result.CombinedOutput)
	cmd.Stderr = io.MultiWriter(result.ErrorOutput, result.CombinedOutput)

	return cmd, result
}

func shortCommandString(command string, args []string) string {
	if len(args) > 0 {
		return command + " " + args[0]
	}

	return command
}

// logCommand outputs command execution intent into the log with a short version of the command and its arguments. The
// returned closure will emit the result of the executed command along with more detailed information including
// elapsed run time to debug output.
func logCommand(result *ExecutionResult) func() {
	var (
		commandStr = shortCommandString(result.Command, result.Arguments)
		started    = time.Now()
	)

	slog.Info("exec", slog.String("command", commandStr))

	return func() {
		var (
			formattedArgs = strings.Join(result.Arguments, " ")
			elapsed       = time.Since(started)
		)

		if result.ReturnCode != 0 {
			if _, err := io.Copy(os.Stderr, result.ErrorOutput); err != nil {
				slog.Error("failed to copy result to stderr", slog.String("error", err.Error()))
			}
		}

		slog.Debug("exec result",
			slog.String("command", commandStr),
			slog.String("args", formattedArgs),
			slog.String("path", result.Path),
			slog.Int("return_code", result.ReturnCode),
			slog.Int64("elapsed_ms", elapsed.Milliseconds()),
		)
	}
}

// Run a command with ars and environment variables set at a specified path
func Run(command string, args []string, path string, env environment.Environment) (*ExecutionResult, error) {
	cmd, result := prepareCommand(command, args, path, env)

	defer logCommand(result)()

	// Pull the return code from the error, if possible
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			// Update the return code and return the result as the error instead
			result.ReturnCode = exitErr.ExitCode()
			return nil, result
		}

		// Likely a system fault that prevented the command from ever running
		return nil, err
	}

	return result, nil
}
