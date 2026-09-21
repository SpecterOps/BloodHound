// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package prsize

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

const (
	Name                = "pr-size"
	Usage               = "Check pull request changed-line limits"
	maximumChangedLines = 1000
)

var (
	ErrPullRequestTooLarge = errors.New("pull request exceeds the changed-line limit")
	excludedPatterns       = []*regexp.Regexp{
		regexp.MustCompile(`(^|/)docs?/`),
		regexp.MustCompile(`(^|/)(architecture|rfc|test|tests|testdata|fixtures|__tests__|__mocks__|mocks|vendormocks)/`),
		regexp.MustCompile(`(^|/)\.yarn/(cache|unplugged|releases)/`),
		regexp.MustCompile(`(^|/)\.(npm|pnpm-store)/`),
		regexp.MustCompile(`(^|/)vendor/`),
		regexp.MustCompile(`(^|/)(go\.sum|yarn\.lock|package-lock\.json|npm-shrinkwrap\.json|pnpm-lock\.yaml)$`),
		regexp.MustCompile(`(^|/)openapi/doc/openapi\.json$`),
		regexp.MustCompile(`(_generated|\.generated|\.gen)\.[^/]+$`),
		regexp.MustCompile(`\.pb\.go$`),
		regexp.MustCompile(`(^|/)wire_gen\.go$`),
		regexp.MustCompile(`\.(md|mdx|rst|adoc|puml|plantuml|c4)$`),
		regexp.MustCompile(`(_test\.(go|[cm]?[jt]sx?)|\.(spec|test|e2e|integration)\.[^/]+)$`),
	}
	excludedPaths = map[string]struct{}{".yarn/build-state.yml": {}, ".yarn/install-state.gz": {}}
)

type command struct {
	env  environment.Environment
	base string
	head string
}

type result struct {
	changedLines  int
	countedFiles  []fileChange
	excludedFiles []fileChange
	binaryFiles   []string
}

type fileChange struct {
	path         string
	additions    int
	deletions    int
	changedLines int
}

func Create(env environment.Environment) *command {
	return &command{env: env}
}

func (s *command) Name() string  { return Name }
func (s *command) Usage() string { return Usage }

func (s *command) Parse(commandIndex int) error {
	var command = flag.NewFlagSet(Name, flag.ExitOnError)

	command.StringVar(&s.base, "base", "", "base revision")
	command.StringVar(&s.head, "head", "", "head revision")
	if err := command.Parse(os.Args[commandIndex+1:]); err != nil {
		return fmt.Errorf("parsing %s command: %w", Name, err)
	}
	if s.base == "" || s.head == "" {
		return fmt.Errorf("parsing %s command: both --base and --head are required", Name)
	}
	return nil
}

func (s *command) Run() error {
	var (
		paths workspace.WorkspacePaths
		err   error
	)

	if paths, err = workspace.FindPaths(s.env); err != nil {
		return fmt.Errorf("finding workspace paths: %w", err)
	}
	resultValue, err := s.count(paths.Root)
	if err != nil {
		return err
	}
	if err := writeResult(os.Stdout, resultValue); err != nil {
		return fmt.Errorf("writing pull request size result: %w", err)
	}
	if resultValue.changedLines > maximumChangedLines {
		return fmt.Errorf("%w: %d counted changed lines (limit: %d)", ErrPullRequestTooLarge, resultValue.changedLines, maximumChangedLines)
	}
	return nil
}

func (s *command) count(root string) (result, error) {
	var plan = cmdrunner.ExecutionPlan{
		Command: "git",
		Args:    []string{"diff", "--no-ext-diff", "--no-renames", "--numstat", "-z", s.base + "..." + s.head},
		Path:    root,
		Env:     s.env.Slice(),
	}

	commandResult, err := cmdrunner.Run(context.TODO(), plan)
	if err != nil {
		return result{}, fmt.Errorf("calculating pull request diff: %w", err)
	}
	return countNumstat(commandResult.StandardOutput.Bytes())
}

func countNumstat(output []byte) (result, error) {
	var resultValue result

	for len(output) > 0 {
		var record []byte
		record, output, _ = bytes.Cut(output, []byte{0})
		if len(record) == 0 {
			continue
		}
		fields := bytes.SplitN(record, []byte{'\t'}, 3)
		if len(fields) != 3 {
			return result{}, fmt.Errorf("parsing git numstat record")
		}
		path := string(fields[2])
		if bytes.Equal(fields[0], []byte("-")) || bytes.Equal(fields[1], []byte("-")) {
			resultValue.binaryFiles = append(resultValue.binaryFiles, path)
			continue
		}
		additions, err := strconv.Atoi(string(fields[0]))
		if err != nil {
			return result{}, fmt.Errorf("parsing additions for %q: %w", path, err)
		}
		deletions, err := strconv.Atoi(string(fields[1]))
		if err != nil {
			return result{}, fmt.Errorf("parsing deletions for %q: %w", path, err)
		}
		file := fileChange{
			path:         path,
			additions:    additions,
			deletions:    deletions,
			changedLines: additions + deletions,
		}
		if isExcluded(path) {
			resultValue.excludedFiles = append(resultValue.excludedFiles, file)
		} else {
			resultValue.changedLines += file.changedLines
			resultValue.countedFiles = append(resultValue.countedFiles, file)
		}
	}
	sort.Slice(resultValue.countedFiles, func(firstIndex, secondIndex int) bool {
		return resultValue.countedFiles[firstIndex].changedLines > resultValue.countedFiles[secondIndex].changedLines
	})
	sort.Slice(resultValue.excludedFiles, func(firstIndex, secondIndex int) bool {
		return resultValue.excludedFiles[firstIndex].changedLines > resultValue.excludedFiles[secondIndex].changedLines
	})
	return resultValue, nil
}

func writeResult(writer io.Writer, resultValue result) error {
	if _, err := fmt.Fprintf(writer, "Pull request size: %d counted changed lines (limit: %d; excluded text files: %d; binary files: %d)\n", resultValue.changedLines, maximumChangedLines, len(resultValue.excludedFiles), len(resultValue.binaryFiles)); err != nil {
		return err
	}
	if err := writeFileChanges(writer, "Largest counted files", resultValue.countedFiles, 10); err != nil {
		return err
	}
	if err := writeFileChanges(writer, "Excluded files", resultValue.excludedFiles, 0); err != nil {
		return err
	}
	if len(resultValue.binaryFiles) > 0 {
		if _, err := fmt.Fprintln(writer, "Binary files (not line-counted):"); err != nil {
			return err
		}
		for _, path := range resultValue.binaryFiles {
			if _, err := fmt.Fprintf(writer, "  %s\n", path); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeFileChanges(writer io.Writer, heading string, files []fileChange, limit int) error {
	if len(files) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(writer, heading+":"); err != nil {
		return err
	}
	for index, file := range files {
		if limit > 0 && index == limit {
			break
		}
		if _, err := fmt.Fprintf(writer, "  %s: +%d / -%d (%d)\n", file.path, file.additions, file.deletions, file.changedLines); err != nil {
			return err
		}
	}
	return nil
}

func isExcluded(path string) bool {
	if _, found := excludedPaths[path]; found {
		return true
	}
	for _, pattern := range excludedPatterns {
		if pattern.MatchString(path) {
			return true
		}
	}
	return false
}
