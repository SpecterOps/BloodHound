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

package redoc

import (
	"fmt"
	"path/filepath"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

func GenerateOpenAPIDoc(projectPath string, submodules []string, env environment.Environment) error {

	// Either we are in the `bhce` submodule or we must find it
	var basePath = projectPath
	for _, submodule := range submodules {
		if filepath.Base(submodule) == "bhce" {
			basePath = submodule
			break
		}
	}

	var (
		srcPath    = filepath.Join(basePath, "packages", "go", "openapi")
		inputPath  = filepath.Join(srcPath, "src/openapi.yaml")
		outputPath = filepath.Join(srcPath, "doc/openapi.json")
		command    = "npx"
		args       = []string{"@redocly/cli@1.18.1", "bundle", inputPath, "--output", outputPath}
	)

	if err := cmdrunner.Run(command, args, srcPath, env); err != nil {
		return fmt.Errorf("generate openapi docs: %w", err)
	}

	return nil
}
