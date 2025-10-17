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
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	maxShortArgs = 2
)

var ErrCmdExecutionFailed = errors.New("command execution failed")

// ExecutionPlan is a configuration struct for Run
type ExecutionPlan struct {
	Command        string
	Args           []string
	Path           string
	Env            []string
	SuppressErrors bool
}

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

func newExecutionResult(ep ExecutionPlan) *ExecutionResult {
	return &ExecutionResult{
		Command:        ep.Command,
		Arguments:      ep.Args,
		Path:           ep.Path,
		StandardOutput: &bytes.Buffer{},
		ErrorOutput:    &bytes.Buffer{},
		CombinedOutput: &bytes.Buffer{},
		ReturnCode:     0,
	}
}

// Run a command with args and environment variables set at a specified path.
func Run(ctx context.Context, ep ExecutionPlan) (*ExecutionResult, error) {
	result := newExecutionResult(ep)
	cmd := exec.Command(ep.Command, ep.Args...)
	cmd.Dir = ep.Path
	cmd.Env = ep.Env
	cmd.Stdout = io.MultiWriter(result.StandardOutput, result.CombinedOutput)
	cmd.Stderr = io.MultiWriter(result.ErrorOutput, result.CombinedOutput)

	defer logCommand(result, ep.SuppressErrors)()

	debugEnabled := slog.Default().Enabled(ctx, slog.LevelDebug)

	if debugEnabled {
		cmd.Stdout = io.MultiWriter(cmd.Stdout, os.Stderr)
		cmd.Stderr = io.MultiWriter(cmd.Stderr, os.Stderr)
	}

	if err := cmd.Run(); err != nil {
		if exitErr := new(exec.ExitError); errors.As(err, &exitErr) {
			result.ReturnCode = exitErr.ExitCode()
			if !debugEnabled && !ep.SuppressErrors {
				fmt.Fprint(os.Stderr, result.CombinedOutput)
			}
			return result, fmt.Errorf("%w: %w", ErrCmdExecutionFailed, err)
		}
		return result, fmt.Errorf("command failed to run: %w", err)
	}

	return result, nil
}

func shortCommandString(command string, args []string, limit int) string {
	var sb strings.Builder

	sb.WriteString(command)

	for i, arg := range args {
		if i < limit {
			sb.WriteString(" ")
			sb.WriteString(arg)
		}
	}

	return sb.String()
}

// logCommand outputs command execution intent into the log with a short version of the command and its arguments. The
// returned closure will emit the result of the executed command along with more detailed information including
// elapsed run time to debug output.
func logCommand(result *ExecutionResult, suppressErrors bool) func() {
	commandStr := shortCommandString(result.Command, result.Arguments, maxShortArgs)
	started := time.Now()

	slog.Info("Exec", slog.String("command", commandStr))

	return func() {
		elapsed := time.Since(started)
		logLevel := slog.LevelDebug
		timeUnit := elapsed.Milliseconds()

		if result.ReturnCode != 0 && !suppressErrors {
			logLevel = slog.LevelError
		}

		slog.Log(context.TODO(), logLevel, "Exec result",
			slog.String("command", result.Command),
			slog.String("args", strings.Join(result.Arguments, " ")),
			slog.String("path", result.Path),
			slog.Int("return_code", result.ReturnCode),
			slog.Int64("elapsed_ms", timeUnit),
		)
	}
}
