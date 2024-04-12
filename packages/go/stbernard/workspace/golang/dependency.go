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
	"fmt"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// InstallGolangCiLint runs go install for the currently supported golangci-lint version
func InstallGolangCiLint(path string, env environment.Environment) error {
	var (
		command = "go"
		args    = []string{"install", "github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2"}
	)

	if err := cmdrunner.Run(command, args, path, env); err != nil {
		return fmt.Errorf("golangci-lint install: %w", err)
	}

	return nil
}
