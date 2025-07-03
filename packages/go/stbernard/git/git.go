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

package git

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/specterops/bloodhound/packages/go/bhlog/level"
	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

var (
	ErrNoValidSemverFound = errors.New("no valid semver found")
	ErrInvalidConfigValue = errors.New("invalid config value")
)

// ListSubmodulePaths finds all submodules in cwd and returns them as a slice of absolute paths
func ListSubmodulePaths(cwd string, env environment.Environment) ([]string, error) {
	var (
		command  = "git"
		args     = []string{"config", "--file", ".gitmodules", "--get-regexp", "path"}
		subPaths = make([]string, 0, 4)
	)

	if _, err := os.Stat(filepath.Join(cwd, ".gitmodules")); errors.Is(err, os.ErrNotExist) {
		return subPaths, nil
	} else if err != nil {
		return subPaths, fmt.Errorf("stat .gitmodules: %w", err)
	} else if result, err := cmdrunner.Run(command, args, cwd, env); err != nil {
		return subPaths, fmt.Errorf("git submodule names: %w", err)
	} else {
		for _, keyValueStr := range strings.Split(strings.TrimSpace(result.StandardOutput.String()), "\n") {
			keyValue := strings.Split(keyValueStr, " ")

			if len(keyValue) != 2 {
				return subPaths, fmt.Errorf("%w: %s", ErrInvalidConfigValue, keyValueStr)
			} else {
				subPaths = append(subPaths, filepath.Join(cwd, keyValue[1]))
			}
		}

		return subPaths, nil
	}
}

// CheckClean checks if the git repository is clean and returns status as a bool. Codes other than exit 1 are returned as an error
func CheckClean(cwd string, env environment.Environment) (bool, error) {
	cmd := exec.Command("git", "diff-index", "--quiet", "HEAD", "--")
	cmd.Env = env.Slice()
	cmd.Dir = cwd

	if level.GlobalAccepts(slog.LevelDebug) {
		cmd.Stderr = os.Stderr
	}

	slog.Info(fmt.Sprintf("Checking repository clean for %s", cwd))

	// We need to run git status first to ensure we don't hit a cache issue
	if _, err := cmdrunner.Run("git", []string{"status"}, cwd, env); err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}

	if _, err := cmdrunner.Run("git", []string{"diff-index", "--quiet", "HEAD", "--"}, cwd, env); err != nil {
		var errResult *cmdrunner.ExecutionResult

		if errors.As(err, &errResult) && errResult.ReturnCode == 1 {
			return false, nil
		}

		return false, fmt.Errorf("git diff-index: %w", err)
	}

	slog.Info(fmt.Sprintf("Finished checking repository clean for %s", cwd))

	return true, nil
}

// FetchCurrentCommitSHA pulls the SHA for the currently active HEAD and returns it as a string
func FetchCurrentCommitSHA(cwd string, env environment.Environment) (string, error) {
	var (
		command = "git"
		args    = []string{"rev-parse", "HEAD"}
	)

	if result, err := cmdrunner.Run(command, args, cwd, env); err != nil {
		return "", fmt.Errorf("git rev-parse: %w", err)
	} else {
		return strings.TrimSpace(result.StandardOutput.String()), nil
	}
}

// ParseLatestVersionFromTags gets the latest semver tag in the repository
func ParseLatestVersionFromTags(cwd string, env environment.Environment) (semver.Version, error) {
	var (
		version semver.Version
	)

	if versions, err := getAllVersionTags(cwd, env); err != nil {
		return version, fmt.Errorf("get version tags from git: %w", err)
	} else {
		return parseLatestVersion(versions)
	}
}

// parseLatestVersion parses a list of found versions and returns the latest from among them
func parseLatestVersion(versions []string) (semver.Version, error) {
	if len(versions) == 0 {
		return semver.Version{}, ErrNoValidSemverFound
	}

	semversions := make([]*semver.Version, 0, len(versions))

	for _, version := range versions {
		if v, err := semver.NewVersion(version); err != nil {
			// skip if version string isn't valid
			continue
		} else {
			semversions = append(semversions, v)
		}
	}

	if len(semversions) == 0 {
		return semver.Version{}, ErrNoValidSemverFound
	}

	sort.Sort(semver.Collection(semversions))

	return *semversions[len(semversions)-1], nil
}

// getAllVersionTags gets the version tags from git and dumps them into a []string
func getAllVersionTags(cwd string, env environment.Environment) ([]string, error) {
	var (
		output bytes.Buffer
	)

	cmd := exec.Command("git", "tag", "--list", "v*")
	cmd.Env = env.Slice()
	cmd.Dir = cwd
	cmd.Stdout = &output

	if level.GlobalAccepts(slog.LevelDebug) {
		cmd.Stderr = os.Stderr
	}

	slog.Info(fmt.Sprintf("Listing tags for %v", cwd))

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git tag --list v*: %w", err)
	}

	slog.Info(fmt.Sprintf("Finished listing tags for %v", cwd))

	return strings.Split(output.String(), "\n"), nil
}
