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

package migrations_test

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func TestVersion_730_Migration(t *testing.T) {
	t.Run("Migration_v730 Success", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.Version730_Migration.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			err := migrations.Version_730_Migration(context.Background(), db)
			require.Nil(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				computers, err := ops.FetchNodes(tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)))

				require.Nil(t, err)

				for _, computer := range computers {
					if computer.ID == harness.Version730_Migration.Computer1.ID {
						smbSigning, err := computer.Properties.Get(ad.SMBSigning.String()).Bool()
						require.Nil(t, err)
						require.True(t, smbSigning)
					} else {
						_, err := computer.Properties.Get(ad.SMBSigning.String()).Bool()
						require.Error(t, err)
						require.True(t, errors.Is(err, graph.ErrPropertyNotFound))
					}
				}

				return nil
			})
		})
	})
}

func TestVersion_900_Migration(t *testing.T) {
	t.Run("Migration_v9.0.0 Success", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.Version900_Migration_Harness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			err := migrations.Version_900_Migration(context.Background(), db)
			require.Nil(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				computers, err := ops.FetchNodes(tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)))

				require.Nil(t, err)

				for _, computer := range computers {
					environmentID, err := computer.Properties.Get(graphschema.EnvironmentIDKey).String()
					require.Nil(t, err)
					require.Equal(t, "ENV-001", environmentID)

					deletedProperty, err := computer.Properties.Get("environment_id").String()
					require.Error(t, err)
					require.True(t, errors.Is(err, graph.ErrPropertyNotFound))
					require.Empty(t, deletedProperty)
				}

				return nil
			})
		})
	})
}

func TestVersion_910_Migration(t *testing.T) {
	t.Run("Migration_v9.1.0 Success", func(t *testing.T) {
		testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
			harness.Version910_Migration_Harness.Setup(testContext)
			return nil
		}, func(harness integration.HarnessDetails, db graph.Database) {
			err := migrations.Version_910_Migration(context.Background(), db)
			require.Nil(t, err)

			db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
				nodes, err := ops.FetchNodes(tx.Nodes())
				require.Nil(t, err)

				for _, node := range nodes {
					switch node.ID {
					case harness.Version910_Migration_Harness.ADNode.ID:
						require.True(t, node.Kinds.ContainsOneOf(ad.Group))
					case harness.Version910_Migration_Harness.AZNode.ID:
						require.False(t, node.Kinds.ContainsOneOf(ad.Group))
					case harness.Version910_Migration_Harness.OGNode.ID:
						require.False(t, node.Kinds.ContainsOneOf(ad.Group))
					}
				}
				return nil
			})
		})
	})
}
