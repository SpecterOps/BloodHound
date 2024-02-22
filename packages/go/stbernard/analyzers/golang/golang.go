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
	"path"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/codeclimate"
	"github.com/specterops/bloodhound/slicesext"
)

var (
	ErrNonZeroExit = errors.New("non-zero exit status")
)

func InstallGolangCiLint(env []string) error {
	cmd := exec.Command("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2")
	cmd.Env = env
	if log.GlobalAccepts(log.LevelDebug) {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}

	log.Infof("Running golangci-lint install")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install golanci-lint: %w", err)
	} else {
		log.Infof("Successfully installed golangci-lint")
		return nil
	}

}

func Run(cwd string, modPaths []string, env []string) ([]codeclimate.Entry, error) {
	var (
		result []codeclimate.Entry
		args   = []string{"run", "--out-format", "code-climate", "--config", ".golangci.json", "--"}
		outb   bytes.Buffer
		errb   bytes.Buffer
	)

	args = append(args, slicesext.Map(modPaths, func(modPath string) string {
		return path.Join(modPath, "...")
	})...)

	cmd := exec.Command("golangci-lint")
	cmd.Env = env
	cmd.Dir = cwd
	cmd.Stderr = &errb
	cmd.Stdout = &outb
	cmd.Args = append(cmd.Args, args...)

	log.Infof("Running golangci-lint")

	err := cmd.Run()
	if _, ok := err.(*exec.ExitError); ok {
		err = ErrNonZeroExit
	} else if err != nil {
		return result, fmt.Errorf("unexpected failure: %w", err)
	}

	log.Infof("Completed golangci-lint")

	if err := json.NewDecoder(&outb).Decode(&result); err != nil {
		log.Errorf("Failed to get valid output from golangci-lint. Error output: %s", errb)
		return result, fmt.Errorf("failed to decode output: %w", err)
	}

	return result, err
}
