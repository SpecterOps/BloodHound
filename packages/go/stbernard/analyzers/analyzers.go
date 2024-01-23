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
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/golang"
	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/js"
)

var (
	ErrSeverityExit = errors.New("high severity linter result")
)

func Run(cwd string, modPaths []string, jsPaths []string, env []string) (string, error) {
	var (
		severityError bool
	)

	golint, err := golang.Run(cwd, modPaths, env)
	if errors.Is(err, golang.ErrNonZeroExit) {
		log.Debug().Msg("Ignoring golangci-lint exit code")
	} else if err != nil {
		return "", fmt.Errorf("golangci-lint: %w", err)
	}

	eslint, err := js.Run(jsPaths, env)
	if errors.Is(err, js.ErrNonZeroExit) {
		log.Debug().Msg("Ignoring eslint exit code")
	} else if err != nil {
		return "", fmt.Errorf("eslint: %w", err)
	}

	codeClimateReport := append(golint, eslint...)

	for idx, entry := range codeClimateReport {
		// We're using err == nil here because we want to do nothing if an error occurs
		if path, err := filepath.Rel(cwd, entry.Location.Path); err != nil {
			log.Debug().Fault(err).Msg("File path is either already relative or cannot be relative to workspace root")
		} else {
			codeClimateReport[idx].Location.Path = path
		}

		if entry.Severity == "error" {
			severityError = true
		}
	}

	if jsonBytes, err := json.MarshalIndent(codeClimateReport, "", "    "); err != nil {
		return "", fmt.Errorf("could not marshal code climate report: %w", err)
	} else if severityError {
		return string(jsonBytes), ErrSeverityExit
	} else {
		return string(jsonBytes), nil
	}
}
