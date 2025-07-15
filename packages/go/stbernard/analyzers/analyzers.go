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

package analyzers

import (
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/codeclimate"
	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/golang"
	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/js"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/specterops/bloodhound/packages/go/stbernard/workspace"
)

var (
	ErrSeverityExit = errors.New("high severity linter result")
)

func outputSeverityMap(results codeclimate.SeverityMap, outputAllSeverity bool) {
	for _, nextSeverity := range results.SortedSeverities() {
		if nextSeverity.Priority() > codeclimate.SeverityMinor.Priority() || outputAllSeverity {
			fileEntries := results[nextSeverity]

			// Issues of this severity should be output directly
			for _, entries := range fileEntries {
				for _, entry := range entries {
					fmt.Printf("%s:%d - %s\n", entry.Location.Path, entry.Location.Lines.Begin, entry.Description)
				}
			}
		}
	}
}

// Run all registered analyzers and collects the results into a CodeClimate-like structure
//
// If one or more entries have a severity of "error" this function returns an error stating
// that a high severity result was found.
func Run(paths workspace.WorkspacePaths, env environment.Environment, outputAllSeverity bool) error {
	if golintResults, err := golang.Run(paths.Root, paths.GoModules, env); err != nil {
		return fmt.Errorf("golangci-lint: %w", err)
	} else if eslintResults, err := js.Run(paths.YarnWorkspaces, env); err != nil {
		return fmt.Errorf("eslint: %w", err)
	} else {
		outputSeverityMap(golintResults, outputAllSeverity)
		outputSeverityMap(eslintResults, outputAllSeverity)

		// Check to see if any high severity issues were identified after output
		combinedResults := codeclimate.CombineSeverityMaps(golintResults, eslintResults)

		// Any finding with a priority greater than Minor is considered a high severity finding
		if combinedResults.HasGreaterSeverity(codeclimate.SeverityMinor) {
			return ErrSeverityExit
		}
	}

	return nil
}
