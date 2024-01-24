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

package tests

import (
	"context"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/common"
)

func RunSuite(db graph.Database, driverName string) (TestSuite, error) {
	if err := db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		return tx.Nodes().Delete()
	}); err != nil {
		return TestSuite{}, err
	}

	// Clear IDs
	StartNodeIDs = make([]graph.ID, SimpleRelationshipsToCreate)
	EndNodeIDs = make([]graph.ID, SimpleRelationshipsToCreate)
	RelationshipIDs = make([]graph.ID, SimpleRelationshipsToCreate)

	// Setup and run the test suite
	suite := TestSuite{
		Name: driverName,
	}

	suite.NewTestCase("Node and Relationship Creation", NodeAndRelationshipCreationTest)
	suite.NewTestCase("Batch Node and Relationship Creation", BatchNodeAndRelationshipCreationTest)

	suite.NewTestCase("Fetch Nodes by ID", FetchNodesByID)
	suite.NewTestCase("Fetch Nodes by Filter Item", FetchNodesByProperty(common.ObjectID.String(), SimpleRelationshipsToCreate/4))
	suite.NewTestCase("Fetch Nodes by Indexed Item", FetchNodesByProperty(common.Name.String(), SimpleRelationshipsToCreate/4))
	suite.NewTestCase("Fetch Nodes by Slice of Filter Properties", FetchNodesByPropertySlice(common.ObjectID.String()))
	suite.NewTestCase("Fetch Nodes by Slice of Indexed Properties", FetchNodesByPropertySlice(common.Name.String()))

	suite.NewTestCase("Node Update", NodeUpdateTests)

	suite.NewTestCase("Fetch Relationships by ID", FetchRelationshipsByID)
	suite.NewTestCase("Fetch Relationships by Filter Item", FetchRelationshipsByProperty(common.Name.String()))
	suite.NewTestCase("Fetch Relationships by Slice of Filter Properties", FetchRelationshipsByPropertySlice)
	suite.NewTestCase("Fetch Relationships by Indexed Start Node Item", FetchRelationshipByStartNodeProperty)

	suite.NewTestCase("Fetch Directional Result by Indexed Start Node Item", FetchDirectionalResultByStartNodeProperty)

	suite.NewTestCase("Batch Delete Nodes by ID", BatchDeleteEndNodesByID)
	suite.NewTestCase("Delete Nodes by Slice of IDs", DeleteStartNodesByIDSlice)

	return suite, suite.Execute(db)
}
