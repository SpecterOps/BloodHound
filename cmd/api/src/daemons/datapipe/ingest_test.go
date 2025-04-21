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

package datapipe_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEinNodeProperties(t *testing.T) {
	var (
		nowUTC     = time.Now().UTC()
		objectID   = "objectid"
		properties = map[string]any{
			datapipe.ReconcileProperty:      false,
			common.Name.String():            "name",
			common.OperatingSystem.String(): "temple",
			ad.DistinguishedName.String():   "distinguished-name",
		}
		normalizedProperties = datapipe.NormalizeEinNodeProperties(properties, objectID, nowUTC)
	)

	assert.Nil(t, normalizedProperties[datapipe.ReconcileProperty])
	assert.NotNil(t, normalizedProperties[common.LastSeen.String()])
	assert.Equal(t, "OBJECTID", normalizedProperties[common.ObjectID.String()])
	assert.Equal(t, "NAME", normalizedProperties[common.Name.String()])
	assert.Equal(t, "DISTINGUISHED-NAME", normalizedProperties[ad.DistinguishedName.String()])
	assert.Equal(t, "TEMPLE", normalizedProperties[common.OperatingSystem.String()])
}

func Test_Monday(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.GenericIngest.Setup(testContext)
		return nil
	},
		func(harness integration.HarnessDetails, db graph.Database) {
			ingestibleRel := ein.IngestibleRelationship{
				SourceProperty: "name",
				Source:         "name a",
				// SourceKind:     ad.User,
				TargetProperty: "name",
				Target:         "name b",
				TargetKind:     ad.Computer,
			}

			db.BatchOperation(testContext.Context(), func(batch graph.Batch) error {
				// OPEN QUESTION:
				// how does this wok when sourceKind/targetKind are nil --> ANSWER: panics
				// expect that source will be not-nil result
				err := batch.Nodes().Filterf(func() graph.Criteria {
					var (
						sourceCriteria, targetCriteria []graph.Criteria
					)
					sourceCriteria = append(sourceCriteria, query.Equals(query.NodeProperty(ingestibleRel.SourceProperty), strings.ToUpper(ingestibleRel.Source)))
					if ingestibleRel.SourceKind != nil {
						sourceCriteria = append(sourceCriteria, query.Kind(query.Node(), ingestibleRel.SourceKind))
					}
					source := query.And(sourceCriteria...)

					targetCriteria = append(targetCriteria, query.Equals(query.NodeProperty(ingestibleRel.TargetProperty), strings.ToUpper(ingestibleRel.Target)))
					if ingestibleRel.TargetKind != nil {
						targetCriteria = append(targetCriteria, query.Kind(query.Node(), ingestibleRel.TargetKind))
					}

					target := query.And(targetCriteria...)

					return query.Or(source, target)
				}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
					for node := range cursor.Chan() {
						fmt.Println(node)
					}

					return nil
				})

				require.Nil(t, err)

				return nil
			})
		})
}
