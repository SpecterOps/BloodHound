// Copyright 2023 Specter Ops, Inc.
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

package azure_test

import (
	"context"
	"github.com/specterops/bloodhound/dawgs/graph"
	schema "github.com/specterops/bloodhound/graphschema"
	azure2 "github.com/specterops/bloodhound/src/analysis/azure"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAnalysisAzure_GraphStats(t *testing.T) {
	testCtx := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())
	testCtx.DatabaseTest(func(harness integration.HarnessDetails, db graph.Database) {

		_, agg, err := azure2.GraphStats(context.TODO(), testCtx.Graph.Database)
		require.Nil(t, err)
		assert.NotZero(t, agg.Tenants)
		assert.NotZero(t, agg.Users)
		assert.NotZero(t, agg.Groups)
		assert.NotZero(t, agg.Apps)
		assert.NotZero(t, agg.ServicePrincipals)
		assert.NotZero(t, agg.Devices)
		assert.NotZero(t, agg.ManagementGroups)
		assert.NotZero(t, agg.Subscriptions)
		assert.NotZero(t, agg.ResourceGroups)
		assert.NotZero(t, agg.VMs)
		assert.NotZero(t, agg.KeyVaults)
		assert.NotZero(t, agg.Relationships)
	})
}
