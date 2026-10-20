// Copyright 2026 Specter Ops, Inc.
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

package js

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/packages/go/stbernard/analyzers/codeclimate"
	"github.com/specterops/bloodhound/packages/go/stbernard/environment"
	"github.com/stretchr/testify/require"
)

func TestDecodeDiagnostics(t *testing.T) {
	var input = `{"diagnostics":[
		{"code":"eslint(no-unused-vars)","severity":"error","message":"unused","filename":"src/example.ts","labels":[{"span":{"line":12}}]},
		{"code":"react(only-export-components)","severity":"warning","message":"export","filename":"/absolute/component.tsx","labels":[{"span":{"line":4}}]},
		{"code":"example","severity":"advice","message":"suggestion","filename":"src/example.ts","labels":[]}
	]}`
	entries, err := decodeDiagnostics(strings.NewReader(input), "/workspace")
	require.NoError(t, err)
	require.Equal(t, []codeclimate.Entry{
		{Description: "eslint(no-unused-vars): unused", Severity: codeclimate.SeverityMajor, Location: codeclimate.Location{Path: "/workspace/src/example.ts", Lines: codeclimate.Lines{Begin: 12}}},
		{Description: "react(only-export-components): export", Severity: codeclimate.SeverityMinor, Location: codeclimate.Location{Path: "/absolute/component.tsx", Lines: codeclimate.Lines{Begin: 4}}},
		{Description: "example: suggestion", Severity: codeclimate.SeverityInfo, Location: codeclimate.Location{Path: "/workspace/src/example.ts", Lines: codeclimate.Lines{Begin: 1}}},
	}, entries)
}

func TestDecodeDiagnosticsInvalidOutput(t *testing.T) {
	var inputs = []string{
		`not json`,
		`{}`,
		`{"diagnostics":null}`,
		`{"diagnostics":[{"severity":"unknown","filename":"example.ts"}]}`,
		`{"diagnostics":[{"severity":"error"}]}`,
		`[]`,
	}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			_, err := decodeDiagnostics(strings.NewReader(input), "/workspace")
			require.Error(t, err)
		})
	}
}

func TestRunExitHandling(t *testing.T) {
	var tests = []struct {
		name        string
		output      string
		exit        string
		wantError   bool
		wantEntries int
	}{
		{name: "clean", output: `{"diagnostics":[]}`, exit: "0"},
		{name: "lint failure", output: `{"diagnostics":[{"code":"rule","message":"message","severity":"error","filename":"src/file.ts","labels":[{"span":{"line":3}}]}]}`, exit: "1", wantEntries: 1},
		{name: "configuration failure", output: `{"diagnostics":[]}`, exit: "1", wantError: true},
		{name: "crash", output: `{"diagnostics":[]}`, exit: "2", wantError: true},
		{name: "invalid output", output: "not json", exit: "0", wantError: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var (
				directory = t.TempDir()
				command   = filepath.Join(directory, "yarn")
				script    = "#!/bin/sh\n[ \"$*\" = \"run lint --format json\" ] || exit 2\nprintf '%s' '" + test.output + "'\nexit " + test.exit + "\n"
			)
			require.NoError(t, os.WriteFile(command, []byte(script), 0o700))
			entries, err := runLint(directory, environment.Environment{environment.YarnCmdVarName: command})
			if test.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, entries, test.wantEntries)
			}
		})
	}
}
