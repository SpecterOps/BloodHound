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
