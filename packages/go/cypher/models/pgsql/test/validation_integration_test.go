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

//go:build manual_integration

package test

import (
	"context"
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/stretchr/testify/require"
)

func TestTranslationTestCases(t *testing.T) {
	var (
		textCtx, done = context.WithCancel(context.Background())
		cfg, err      = utils.LoadIntegrationTestConfig()
	)

	defer done()

	require.Nil(t, err)

	if connection, err := dawgs.Open(context.TODO(), pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: size.Gibibyte,
		DriverCfg:             cfg.Database.PostgreSQLConnectionString(),
	}); err != nil {
		t.Fatalf("Failed opening database connection: %v", err)
	} else if pgConnection, typeOK := connection.(*pg.Driver); !typeOK {
		t.Fatalf("Invalid connection type: %T", connection)
	} else {
		defer connection.Close(textCtx)

		graphSchema := schema.DefaultGraphSchema()

		graphSchema.Graphs[0].Nodes = graphSchema.DefaultGraph.Nodes.Add(
			graph.StringKind("NodeKind1"),
			graph.StringKind("NodeKind2"),
		)

		graphSchema.Graphs[0].Edges = graphSchema.DefaultGraph.Edges.Add(
			graph.StringKind("EdgeKind1"),
			graph.StringKind("EdgeKind2"),
		)

		if err := connection.AssertSchema(textCtx, graphSchema); err != nil {
			t.Fatalf("Failed asserting graph schema: %v", err)
		}

		casesRun := 0

		if testCases, err := ReadTranslationTestCases(); err != nil {
			t.Fatal(err)
		} else {
			for _, testCase := range testCases {
				t.Run(testCase.Name, func(t *testing.T) {
					defer func() {
						if err := recover(); err != nil {
							debug.PrintStack()
							t.Error(err)
						}
					}()

					testCase.AssertLive(textCtx, t, pgConnection)
				})

				casesRun += 1
			}
		}

		fmt.Printf("Validated %d test cases\n", casesRun)
	}
}
