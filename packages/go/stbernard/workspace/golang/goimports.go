// Copyright 2025 Specter Ops, Inc.
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
	"os"

	"github.com/specterops/bloodhound/packages/go/stbernard/cmdrunner"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
)

// Running checks for unused go imports and formatting .go files
func RunGoImports(env environment.Environment) error {
	rootDir, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd := "goimports"
	args := []string{"-w", rootDir}

	if err := cmdrunner.Run(cmd, args, rootDir, env); err != nil {
		return err
	}

	return nil
}
