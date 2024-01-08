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

package neo4j

import (
	"testing"

	neo4j_core "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/vendormocks/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	Entity     = graph.StringKind("Base")
	Group      = graph.StringKind("Group")
	User       = graph.StringKind("User")
	Computer   = graph.StringKind("Computer")
	HasSession = graph.StringKind("HasSession")
)

func TestNeo4jTransaction_UpdateNodeBy(t *testing.T) {
	var (
		mockCtl         = gomock.NewController(t)
		resultMock      = neo4j.NewMockResult(mockCtl)
		transactionMock = neo4j.NewMockTransaction(mockCtl)
		tx              = &batchTransaction{
			innerTx: &neo4jTransaction{
				innerTx: transactionMock,
			},
			batchWriteSize: 10,
		}

		nodeProperties = map[string]any{
			"objectid": "1-2-3",
			"name":     "a_name",
		}
		node = graph.PrepareNode(graph.AsProperties(nodeProperties), Entity, User)

		expectedIdentityProperties = map[string]any{
			"p": []map[string]any{
				nodeProperties,
				nodeProperties,
				nodeProperties,
			},
		}
	)

	resultMock.EXPECT().Err().Return(nil)
	transactionMock.EXPECT().Run(`unwind $p as p merge (n:Base {objectid:p.objectid}) set n += p, n:Base, n:User;`, expectedIdentityProperties).Return(resultMock, nil)
	transactionMock.EXPECT().Commit().Return(nil)

	require.Nil(t, tx.UpdateNodeBy(graph.NodeUpdate{
		Node:               node,
		IdentityKind:       Entity,
		IdentityProperties: []string{"objectid"},
	}))

	require.Nil(t, tx.UpdateNodeBy(graph.NodeUpdate{
		Node:               node,
		IdentityKind:       Entity,
		IdentityProperties: []string{"objectid"},
	}))

	require.Nil(t, tx.UpdateNodeBy(graph.NodeUpdate{
		Node:               node,
		IdentityKind:       Entity,
		IdentityProperties: []string{"objectid"},
	}))

	tx.Commit()
}

func TestNeo4jTransaction_UpdateNodeBy_Batch(t *testing.T) {
	var (
		mockCtl         = gomock.NewController(t)
		resultMock      = neo4j.NewMockResult(mockCtl)
		transactionMock = neo4j.NewMockTransaction(mockCtl)
		tx              = &batchTransaction{
			innerTx: &neo4jTransaction{
				innerTx: transactionMock,
			},
			batchWriteSize: 10,
		}

		nodeProperties = map[string]any{
			"objectid": "1-2-3",
			"name":     "a_name",
		}

		expectedUserProperties = map[string]any{
			"p": []map[string]any{
				nodeProperties,
				nodeProperties,
			},
		}

		expectedGroupProperties = map[string]any{
			"p": []map[string]any{
				nodeProperties,
			},
		}
	)

	resultMock.EXPECT().Err().Return(nil).Times(2)
	transactionMock.EXPECT().Run(`unwind $p as p merge (n:Base {objectid:p.objectid}) set n += p, n:Base, n:User;`, expectedUserProperties).Return(resultMock, nil)
	transactionMock.EXPECT().Run(`unwind $p as p merge (n:Base {objectid:p.objectid}) set n += p, n:Base, n:Group;`, expectedGroupProperties).Return(resultMock, nil)
	transactionMock.EXPECT().Commit().Return(nil)

	require.Nil(t, tx.UpdateNodeBy(graph.NodeUpdate{
		Node:               graph.PrepareNode(graph.AsProperties(nodeProperties), Entity, User),
		IdentityKind:       Entity,
		IdentityProperties: []string{"objectid"},
	}))

	require.Nil(t, tx.UpdateNodeBy(graph.NodeUpdate{
		Node:               graph.PrepareNode(graph.AsProperties(nodeProperties), Entity, User),
		IdentityKind:       Entity,
		IdentityProperties: []string{"objectid"},
	}))

	require.Nil(t, tx.UpdateNodeBy(graph.NodeUpdate{
		Node:               graph.PrepareNode(graph.AsProperties(nodeProperties), Entity, Group),
		IdentityKind:       Entity,
		IdentityProperties: []string{"objectid"},
	}))

	tx.Commit()
}

func TestNeo4jTransaction_UpdateRelationshipBy_Batch(t *testing.T) {
	const (
		batchWriteSize = 1
		writeFlushSize = batchWriteSize * 2
	)

	var (
		mockCtl         = gomock.NewController(t)
		resultMock      = neo4j.NewMockResult(mockCtl)
		transactionMock = neo4j.NewMockTransaction(mockCtl)
		sessionMock     = neo4j.NewMockSession(mockCtl)

		tx = &batchTransaction{
			innerTx: &neo4jTransaction{
				session:        sessionMock,
				innerTx:        nil,
				writeFlushSize: writeFlushSize,
				batchWriteSize: batchWriteSize,
			},
			batchWriteSize: batchWriteSize,
		}
	)

	submitf := func() {
		require.Nil(t, tx.UpdateRelationshipBy(graph.RelationshipUpdate{
			Relationship:            graph.PrepareRelationship(graph.NewProperties(), HasSession),
			IdentityProperties:      []string{"objectid"},
			Start:                   graph.PrepareNode(graph.NewProperties(), Entity, User),
			StartIdentityKind:       Entity,
			StartIdentityProperties: []string{"objectid"},
			End:                     graph.PrepareNode(graph.NewProperties(), Entity, Computer),
			EndIdentityKind:         Entity,
			EndIdentityProperties:   []string{"objectid"},
		}))
	}

	// Batch should submit batch queries with a set of payloads for each
	sessionMock.EXPECT().BeginTransaction(gomock.Any()).Return(transactionMock, nil)
	resultMock.EXPECT().Err().Return(nil).Times(writeFlushSize / batchWriteSize)
	transactionMock.EXPECT().Run(gomock.Any(), gomock.Any()).Return(resultMock, nil).Times(writeFlushSize / batchWriteSize)
	transactionMock.EXPECT().Commit().Return(nil)

	for i := 0; i < writeFlushSize; i++ {
		submitf()
	}

	// Expect that a new transaction is opened and then closed with commit after the final submission and call to commit
	sessionMock.EXPECT().BeginTransaction(gomock.Any()).Return(transactionMock, nil)
	resultMock.EXPECT().Err().Return(nil)
	transactionMock.EXPECT().Run(`unwind $p as p merge (s:Base {objectid:p.s.objectid}) merge (e:Base {objectid:p.e.objectid}) merge (s)-[r:HasSession {objectid:p.r.objectid}]->(e) set s += p.s, e += p.e, r += p.r, s:Base, s:User, e:Base, e:Computer, s.lastseen = datetime({timezone: 'UTC'}), e.lastseen = datetime({timezone: 'UTC'});`, gomock.Any()).Return(resultMock, nil)
	transactionMock.EXPECT().Commit().Return(nil)

	submitf()
	tx.Commit()
}

func TestNeo4jTransaction_UpdateRelationshipBy(t *testing.T) {
	var (
		mockCtl         = gomock.NewController(t)
		resultMock      = neo4j.NewMockResult(mockCtl)
		transactionMock = neo4j.NewMockTransaction(mockCtl)
		tx              = &batchTransaction{
			innerTx: &neo4jTransaction{
				innerTx: transactionMock,
			},
			batchWriteSize: 10,
		}

		relationshipProperties = map[string]any{
			"objectid": "a-b-c",
		}

		startNodeProperties = map[string]any{
			"objectid": "1-2-3",
			"name":     "a_name",
		}

		endNodeProperties = map[string]any{
			"objectid": "2-3-4",
			"name":     "a_name",
		}

		expectedPayloa = map[string]any{
			"p": []map[string]any{{
				"r": map[string]any{
					"objectid": "a-b-c",
				},
				"s": map[string]any{
					"name":     "a_name",
					"objectid": "1-2-3",
				},
				"e": map[string]any{
					"name":     "a_name",
					"objectid": "2-3-4",
				},
			}, {
				"r": map[string]any{
					"objectid": "a-b-c",
				},
				"s": map[string]any{
					"name":     "a_name",
					"objectid": "1-2-3",
				},
				"e": map[string]any{
					"name":     "a_name",
					"objectid": "2-3-4",
				},
			}},
		}
	)

	resultMock.EXPECT().Err().Return(nil)
	transactionMock.EXPECT().Run(`unwind $p as p merge (s:Base {objectid:p.s.objectid}) merge (e:Base {objectid:p.e.objectid}) merge (s)-[r:HasSession {objectid:p.r.objectid}]->(e) set s += p.s, e += p.e, r += p.r, s:Base, s:User, e:Base, e:Computer, s.lastseen = datetime({timezone: 'UTC'}), e.lastseen = datetime({timezone: 'UTC'});`, expectedPayloa).Return(resultMock, nil)
	transactionMock.EXPECT().Commit().Return(nil)

	require.Nil(t, tx.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship:            graph.PrepareRelationship(graph.AsProperties(relationshipProperties), HasSession),
		IdentityProperties:      []string{"objectid"},
		Start:                   graph.PrepareNode(graph.AsProperties(startNodeProperties), Entity, User),
		StartIdentityKind:       Entity,
		StartIdentityProperties: []string{"objectid"},
		End:                     graph.PrepareNode(graph.AsProperties(endNodeProperties), Entity, Computer),
		EndIdentityKind:         Entity,
		EndIdentityProperties:   []string{"objectid"},
	}))

	require.Nil(t, tx.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship:            graph.PrepareRelationship(graph.AsProperties(relationshipProperties), HasSession),
		IdentityProperties:      []string{"objectid"},
		Start:                   graph.PrepareNode(graph.AsProperties(startNodeProperties), Entity, User),
		StartIdentityKind:       Entity,
		StartIdentityProperties: []string{"objectid"},
		End:                     graph.PrepareNode(graph.AsProperties(endNodeProperties), Entity, Computer),
		EndIdentityKind:         Entity,
		EndIdentityProperties:   []string{"objectid"},
	}))

	tx.Commit()
}

func TestNeo4jTransaction_CreateNode(t *testing.T) {
	var (
		mockCtl         = gomock.NewController(t)
		resultMock      = neo4j.NewMockResult(mockCtl)
		transactionMock = neo4j.NewMockTransaction(mockCtl)
		tx              = neo4jTransaction{
			innerTx: transactionMock,
		}

		properties = map[string]any{
			"prop": "value",
		}

		expectedProperties = map[string]any{
			"p0": map[string]any{
				"prop": "value",
			},
		}
	)

	transactionMock.EXPECT().Run(`create (n:User $p0) return n`, expectedProperties).Return(resultMock, nil)
	transactionMock.EXPECT().Commit()

	resultMock.EXPECT().Err().Return(nil)
	resultMock.EXPECT().Next().Return(true)
	resultMock.EXPECT().Record().Return(&neo4j_core.Record{
		Values: []any{
			neo4j_core.Node{
				Id:     0,
				Labels: []string{"User"},
				Props:  properties,
			},
		},
		Keys: []string{
			"n",
		},
	})

	_, err := tx.CreateNode(graph.AsProperties(properties), User)
	require.Nil(t, err)
}
