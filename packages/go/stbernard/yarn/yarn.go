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

package yarn

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/specterops/bloodhound/log"
)

func InstallWorkspaceDeps(jsPaths []string, env []string) error {
	for _, path := range jsPaths {
		if err := yarnInstall(path, env); err != nil {
			return fmt.Errorf("failed to run yarn install at %v: %w", path, err)
		}
	}

	return nil
}

func yarnInstall(path string, env []string) error {
	cmd := exec.Command("yarn", "install")
	cmd.Env = env
	cmd.Dir = path
	if log.GlobalAccepts(log.LevelDebug) {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}

	log.Infof("Running yarn install for %v", path)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yarn install: %w", err)
	} else {
		log.Infof("Finished yarn install for %v", path)
		return nil
	}
}
