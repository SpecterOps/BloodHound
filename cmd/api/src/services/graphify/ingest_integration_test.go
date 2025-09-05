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

//go:build serial_integration
// +build serial_integration

package graphify_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

// verifies that files that bypassed validation controls due to being uploaded as zips receive validation attention in the datapipe,
// and that invalid files are not ingested into the graph
func Test_ReadFileForIngest(t *testing.T) {
	var (
		ingestSchema, _ = upload.LoadIngestSchema()
		validReader     = bytes.NewReader([]byte(`{"graph":{"nodes":[{"id": "1234", "kinds": ["kindA","kindB"],"properties":{"true": true,"hello":"world"}}]}}`))
		// invalidReader simulates reading a file that doesn't pass jsonschema validation against the nodes schema.
		// ReadFileForIngest() should kick out, ingesting no graph data
		invalidReader = bytes.NewReader([]byte(`{"graph":{"nodes": [{"id":1234}]}}`))
		readOptions   = graphify.ReadOptions{
			IngestSchema:       ingestSchema,
			FileType:           model.FileTypeZip,
			RegisterSourceKind: func(k graph.Kind) error { return nil }, // stub this out
		}
	)

	t.Run("happy path. a file uploaded as a zip passes validation and is written to the graph", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.BatchTest(func(harness integration.HarnessDetails, batch graph.Batch) {
			ingestContext := graphify.NewIngestContext(testContext.Context(), graphify.WithBatchUpdater(batch))

			err := graphify.ReadFileForIngest(ingestContext, validReader, readOptions)
			require.Nil(t, err)

		}, func(details integration.HarnessDetails, tx graph.Transaction) {

			err := tx.Nodes().
				Filter(query.Equals(query.Property(query.Node(), "objectid"), "1234")).
				Fetch(func(cursor graph.Cursor[*graph.Node]) error {
					numNodes := 0
					for node := range cursor.Chan() {
						// assert kinds were added correctly
						require.Contains(t, node.Kinds, graph.StringKind("kindA"))
						require.Contains(t, node.Kinds, graph.StringKind("kindB"))

						// assert properties were saved correctly
						booleanProperty, _ := node.Properties.Get("true").Bool()
						require.Equal(t, true, booleanProperty)
						stringProperty, _ := node.Properties.Get("hello").String()
						require.Equal(t, "world", stringProperty)

						numNodes++
					}

					// assert 1 node was ingested
					require.Equal(t, 1, numNodes)
					return nil
				})

			require.Nil(t, err)
		})
	})

	t.Run("failure path. a file uploaded as a zip fails validation and nothing is written to the graph", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error { return nil }, func(harness integration.HarnessDetails, db graph.Database) {
			_ = db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
				ingestContext := graphify.NewIngestContext(testContext.Context(), graphify.WithBatchUpdater(batch))

				err := graphify.ReadFileForIngest(ingestContext, invalidReader, readOptions)
				require.NotNil(t, err)
				var report upload.ValidationReport
				if errors.As(err, &report) {
					// verify nodes[0] caused a validation error
					require.Len(t, report.ValidationErrors, 1)
				}
				return nil
			})

			// assert that zero nodes exist
			_ = db.ReadTransaction(testContext.Context(), func(tx graph.Transaction) error {
				numNodes, err := tx.Nodes().Count()
				require.Nil(t, err)
				require.Equal(t, int64(0), numNodes)
				return nil
			})

		})
	})
}
