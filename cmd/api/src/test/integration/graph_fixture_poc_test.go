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

//go:build integration
// +build integration

package integration_test

import (
	"embed"
	"io/fs"
	"path/filepath"
	"testing"

	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/src/test/integration"
)

//go:embed harnesses
var harnesses embed.FS

func TestGraphLoader(t *testing.T) {
	// This first part is a programmatic setup of a 'fileSet' which will
	// quickly roll all the harness.json files from the test harnesses into
	// a map for testing.

	// In a real scenario, you might have several fixture sources that you
	// want to pull fixtures from. You could programmatically generate a file
	// set, or you could build it manually. That would look something like this:

	// //go:embed fixtures1
	// var fixtures1 embed.FS
	// //go:embed fixtures2
	// var fixtures2 embed.FS

	// var fileSet = map[fs.FS][]string{
	// 	&fixtures1: {
	// 		"fixtures1/test_domain_fixture.json",
	// 		"fixtures1/test_domain_members_fixture.json",
	// 	},
	// 	&fixtures2: {
	// 		"fixtures2/test_adcs_templates_fixture.json",
	// 		"fixtures2/test_adcs_members_fixture.json",
	// 	},
	// }

	var fileSet = make(map[fs.FS][]string)
	fileSet[&harnesses] = make([]string, 0)

	if entries, err := fs.ReadDir(harnesses, "harnesses"); err != nil {
		t.Fatal(err)
	} else {
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			fileSet[&harnesses] = append(fileSet[&harnesses], "harnesses/"+entry.Name())
		}
	}

	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	graphCtx := testContext.Graph

	if err := lab.LoadGraphFixtureFiles(graphCtx.Database, fileSet); err != nil {
		t.Fatal(err)
	}
}
