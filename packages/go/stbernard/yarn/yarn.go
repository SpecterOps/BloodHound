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

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// InstallWorkspaceDeps runs yarn install for a given list of jsPaths
func InstallWorkspaceDeps(jsPaths []string, env environment.Environment) error {
	var (
		command = "yarn"
		args    = []string{"install"}
	)

	for _, path := range jsPaths {
		if err := cmdrunner.RunAtPathWithEnv(command, args, path, env); err != nil {
			return fmt.Errorf("yarn install at %v: %w", path, err)
		}
	}

	return nil
}

// BuildWorkspace runs yarn build for the current working directory
func BuildWorkspace(cwd string, env environment.Environment) error {
	var (
		command = "yarn"
		args    = []string{"build"}
	)

	if err := cmdrunner.RunAtPathWithEnv(command, args, cwd, env); err != nil {
		return fmt.Errorf("yarn build at %v: %w", cwd, err)
	} else {
		return nil
	}
}
