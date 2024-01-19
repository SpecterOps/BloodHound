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

package v2_test

import (
	"context"
	"encoding/json"
	schema "github.com/specterops/bloodhound/graphschema"
	"testing"

	"github.com/specterops/bloodhound/analysis/azure"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/common"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestGetAZEntityInformation(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, schema.DefaultGraphSchema())

	testContext.TransactionalTest(func(harness integration.HarnessDetails, tx graph.Transaction) {
		objectID, err := harness.AZGroupMembership.Group.Properties.Get(common.ObjectID.String()).String()
		require.Nil(t, err)

		groupInformation, err := v2.GetAZEntityInformation(context.Background(), testContext.Graph.Database, "groups", objectID, true)
		require.Nil(t, err)

		groupInformationJSON, err := json.Marshal(groupInformation)
		require.Nil(t, err)

		groupDetails := azure.GroupDetails{}
		require.Nil(t, json.Unmarshal(groupInformationJSON, &groupDetails))
		require.Equal(t, 3, groupDetails.GroupMembers)
	})
}
