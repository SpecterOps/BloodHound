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
	"context"
	"encoding/json"
	"fmt"
	"github.com/specterops/bloodhound/dawgs/drivers"
	"github.com/specterops/bloodhound/log"
	"sort"
	"strings"

	"github.com/specterops/bloodhound/dawgs/query/neo4j"
	"github.com/specterops/bloodhound/dawgs/util/size"

	neo4j_core "github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
)

const (
	DefaultBatchWriteSize = 20_000
	DefaultWriteFlushSize = DefaultBatchWriteSize * 5

	// DefaultConcurrentConnections defines the default number of concurrent graph database connections allowed.
	DefaultConcurrentConnections = 50
)

type innerTransaction interface {
	Raw(cypher string, params map[string]any) graph.Result
}

type neo4jTransaction struct {
	cfg                  graph.TransactionConfig
	ctx                  context.Context
	session              neo4j_core.Session
	innerTx              neo4j_core.Transaction
	writes               int
	writeFlushSize       int
	batchWriteSize       int
	traversalMemoryLimit size.Size
}

func (s *neo4jTransaction) WithGraph(graphSchema graph.Graph) graph.Transaction {
	// Neo4j does not support multiple graph namespaces within the same database. While Neo4j enterprise supports
	// multiple databases this is not the same. Graph namespaces could be hacked using labels but this then requires
	// a material change in how labels are applied and therefore was not plumbed.
	//
	// This has no material effect on the usage of the database: the schema is the same for all graph namespaces.
	return s
}

func (s *neo4jTransaction) Query(query string, parameters map[string]any) graph.Result {
	return s.Raw(query, parameters)
}

func (s *neo4jTransaction) updateRelationshipsBy(updates ...graph.RelationshipUpdate) error {
	var (
		numUpdates                       = len(updates)
		statements, queryParameterArrays = cypherBuildRelationshipUpdateQueryBatch(updates)
	)

	for parameterIdx, stmt := range statements {
		propertyBags := queryParameterArrays[parameterIdx]
		chunkMap := make([]map[string]any, 0, s.batchWriteSize)

		for _, val := range propertyBags {
			chunkMap = append(chunkMap, val)

			if len(chunkMap) == s.batchWriteSize {
				if result := s.Raw(stmt, map[string]any{
					"p": chunkMap,
				}); result.Error() != nil {
					return result.Error()
				}

				chunkMap = chunkMap[:0]
			}
		}

		if len(chunkMap) > 0 {
			if result := s.Raw(stmt, map[string]any{
				"p": chunkMap,
			}); result.Error() != nil {
				return result.Error()
			}
		}
	}

	return s.logWrites(numUpdates)
}

func (s *neo4jTransaction) UpdateRelationshipBy(update graph.RelationshipUpdate) error {
	return s.updateRelationshipsBy(update)
}

func (s *neo4jTransaction) updateNodesBy(updates ...graph.NodeUpdate) error {
	var (
		numUpdates                     = len(updates)
		statements, queryParameterMaps = cypherBuildNodeUpdateQueryBatch(updates)
	)

	for parameterIdx, stmt := range statements {
		if result := s.Raw(stmt, queryParameterMaps[parameterIdx]); result.Error() != nil {
			return fmt.Errorf("update nodes by error on statement (%s): %s", stmt, result.Error())
		}
	}

	return s.logWrites(numUpdates)
}

func (s *neo4jTransaction) UpdateNodeBy(update graph.NodeUpdate) error {
	return s.updateNodesBy(update)
}

func newTransaction(ctx context.Context, session neo4j_core.Session, cfg graph.TransactionConfig, writeFlushSize int, batchWriteSize int, traversalMemoryLimit size.Size) *neo4jTransaction {
	if traversalMemoryLimit == 0 {
		traversalMemoryLimit = 2 * size.Gibibyte
	}

	return &neo4jTransaction{
		cfg:                  cfg,
		ctx:                  ctx,
		session:              session,
		writeFlushSize:       writeFlushSize,
		batchWriteSize:       batchWriteSize,
		traversalMemoryLimit: traversalMemoryLimit,
	}
}

func (s *neo4jTransaction) flushTx() error {
	defer func() {
		s.innerTx = nil
	}()

	if err := s.innerTx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *neo4jTransaction) currentTx() neo4j_core.Transaction {
	if s.innerTx == nil {
		if newTx, err := s.session.BeginTransaction(neo4j_core.WithTxTimeout(s.cfg.Timeout)); err != nil {
			return newErrorTransactionWrapper(err)
		} else {
			s.innerTx = newTx
		}
	}

	return s.innerTx
}

func (s *neo4jTransaction) logWrites(writes int) error {
	if s.writes += writes; s.writes >= s.writeFlushSize {
		if err := s.flushTx(); err != nil {
			return err
		}

		s.writes = 0
	}

	return nil
}

func (s *neo4jTransaction) runAndLog(stmt string, params map[string]any, numWrites int) graph.Result {
	result := s.Raw(stmt, params)

	if result.Error() == nil {
		if err := s.logWrites(numWrites); err != nil {
			return NewResult(stmt, err, nil)
		}
	}

	return result
}

func (s *neo4jTransaction) updateNode(updatedNode *graph.Node) error {
	queryBuilder := neo4j.NewQueryBuilder(query.SinglePartQuery(
		query.Where(
			query.Equals(query.NodeID(), updatedNode.ID),
		),

		query.Updatef(func() graph.Criteria {
			var (
				properties       = updatedNode.Properties
				updateStatements []graph.Criteria
			)

			if addedKinds := updatedNode.AddedKinds; len(addedKinds) > 0 {
				updateStatements = append(updateStatements, query.AddKinds(query.Node(), addedKinds))
			}

			if deletedKinds := updatedNode.DeletedKinds; len(deletedKinds) > 0 {
				updateStatements = append(updateStatements, query.DeleteKinds(query.Node(), deletedKinds))
			}

			if modifiedProperties := properties.ModifiedProperties(); len(modifiedProperties) > 0 {
				updateStatements = append(updateStatements, query.SetProperties(query.Node(), modifiedProperties))
			}

			if deletedProperties := properties.DeletedProperties(); len(deletedProperties) > 0 {
				updateStatements = append(updateStatements, query.DeleteProperties(query.Node(), deletedProperties...))
			}

			return updateStatements
		}),
	))

	if err := queryBuilder.Prepare(); err != nil {
		return err
	} else if cypherQuery, err := queryBuilder.Render(); err != nil {
		return graph.NewError(cypherQuery, err)
	} else if result := s.Raw(cypherQuery, queryBuilder.Parameters); result.Error() != nil {
		return result.Error()
	}

	return nil
}

func (s *neo4jTransaction) createNode(properties *graph.Properties, kinds ...graph.Kind) (*graph.Node, error) {
	queryBuilder := neo4j.NewQueryBuilder(query.SinglePartQuery(
		query.Create(
			query.NodePattern(
				kinds,
				query.Parameter(properties.Map),
			),
		),

		query.Returning(
			query.Node(),
		),
	))

	if err := queryBuilder.Prepare(); err != nil {
		return nil, err
	} else if statement, err := queryBuilder.Render(); err != nil {
		return nil, err
	} else if result := s.Raw(statement, queryBuilder.Parameters); result.Error() != nil {
		return nil, result.Error()
	} else if !result.Next() {
		return nil, graph.ErrNoResultsFound
	} else {
		var node graph.Node
		return &node, result.Scan(&node)
	}
}

func (s *neo4jTransaction) createRelationshipByIDs(startNodeID, endNodeID graph.ID, kind graph.Kind, properties *graph.Properties) (*graph.Relationship, error) {
	queryBuilder := neo4j.NewQueryBuilder(query.SinglePartQuery(
		query.Where(
			query.And(
				query.Equals(query.StartID(), startNodeID),
				query.Equals(query.EndID(), endNodeID),
			),
		),
		query.Create(
			query.Start(),
			query.RelationshipPattern(kind, query.Parameter(properties.Map), graph.DirectionOutbound),
			query.End(),
		),

		query.Returning(
			query.Relationship(),
		),
	))

	if err := queryBuilder.Prepare(); err != nil {
		return nil, err
	} else if statement, err := queryBuilder.Render(); err != nil {
		return nil, err
	} else if result := s.Raw(statement, queryBuilder.Parameters); result.Error() != nil {
		return nil, result.Error()
	} else if !result.Next() {
		return nil, graph.ErrNoResultsFound
	} else {
		var relationship graph.Relationship
		return &relationship, result.Scan(&relationship)
	}
}

func (s *neo4jTransaction) Raw(stmt string, params map[string]any) graph.Result {
	const maxParametersToRender = 12

	if drivers.IsQueryAnalysisEnabled() {
		var (
			parametersWritten = 0
			prettyParameters  strings.Builder
			sortedKeys        []string
		)

		if len(params) > maxParametersToRender {
			sortedKeys = make([]string, 0, maxParametersToRender)
		} else {
			sortedKeys = make([]string, 0, len(params))
		}

		for key := range params {
			if sortedKeys = append(sortedKeys, key); len(sortedKeys) >= maxParametersToRender {
				break
			}
		}

		sort.Strings(sortedKeys)

		for _, key := range sortedKeys {
			value := params[key]

			if parametersWritten++; parametersWritten >= maxParametersToRender {
				break
			} else if parametersWritten > 1 {
				prettyParameters.WriteString(", ")
			}

			prettyParameters.WriteString(key)
			prettyParameters.WriteString(":")

			if marshalledValue, err := json.Marshal(value); err != nil {
				log.Errorf("Unable to marshal query parameter %s", key)
			} else {
				prettyParameters.Write(marshalledValue)
			}
		}

		log.Info().Str("dawgs_db_driver", DriverName).Msgf("%s - %s", stmt, prettyParameters.String())
	}

	driverResult, err := s.currentTx().Run(stmt, params)
	return NewResult(stmt, err, driverResult)
}

func (s *neo4jTransaction) Nodes() graph.NodeQuery {
	return NewNodeQuery(s.ctx, s)
}

func (s *neo4jTransaction) Relationships() graph.RelationshipQuery {
	return NewRelationshipQuery(s.ctx, s)
}

func (s *neo4jTransaction) DeleteNodesBySlice(ids []graph.ID) error {
	return s.runAndLog(cypherDeleteNodesByID, map[string]any{
		idListParameterName: ids,
	}, len(ids)).Error()
}

func (s *neo4jTransaction) DeleteRelationshipsBySlice(ids []graph.ID) error {
	return s.runAndLog(cypherDeleteRelationshipsByID, map[string]any{
		"p": ids,
	}, len(ids)).Error()
}

func (s *neo4jTransaction) Commit() error {
	if s.innerTx != nil {
		txRef := s.innerTx
		s.innerTx = nil

		return txRef.Commit()
	}

	return nil
}

func (s *neo4jTransaction) Close() error {
	if s.innerTx != nil {
		txRef := s.innerTx
		s.innerTx = nil

		return txRef.Close()
	}

	return nil
}

func (s *neo4jTransaction) CreateNode(properties *graph.Properties, kinds ...graph.Kind) (*graph.Node, error) {
	if node, err := s.createNode(properties, kinds...); err != nil {
		return nil, err
	} else {
		return node, s.logWrites(1)
	}
}

func (s *neo4jTransaction) UpdateNode(target *graph.Node) error {
	if err := s.updateNode(target); err != nil {
		return err
	}

	return s.logWrites(1)
}

func (s *neo4jTransaction) CreateRelationship(startNode, endNode *graph.Node, kind graph.Kind, properties *graph.Properties) (*graph.Relationship, error) {
	return s.CreateRelationshipByIDs(startNode.ID, endNode.ID, kind, properties)
}

func (s *neo4jTransaction) CreateRelationshipByIDs(startNodeID, endNodeID graph.ID, kind graph.Kind, properties *graph.Properties) (*graph.Relationship, error) {
	if rel, err := s.createRelationshipByIDs(startNodeID, endNodeID, kind, properties); err != nil {
		return nil, err
	} else {
		return rel, s.logWrites(1)
	}
}

func (s *neo4jTransaction) DeleteNode(id graph.ID) error {
	return s.runAndLog(cypherDeleteNodeByID, map[string]any{
		idParameterName: id,
	}, 1).Error()
}

func (s *neo4jTransaction) DeleteRelationship(id graph.ID) error {
	return s.runAndLog(cypherDeleteRelationshipByID, map[string]any{
		idParameterName: id,
	}, 1).Error()
}

func (s *neo4jTransaction) UpdateRelationship(relationship *graph.Relationship) error {
	queryBuilder := neo4j.NewQueryBuilder(query.SinglePartQuery(
		query.Where(
			query.Equals(query.RelationshipID(), relationship.ID),
		),

		query.Updatef(func() graph.Criteria {
			var (
				properties       = relationship.Properties
				updateStatements []graph.Criteria
			)

			if modifiedProperties := properties.ModifiedProperties(); len(modifiedProperties) > 0 {
				updateStatements = append(updateStatements, query.SetProperties(query.Relationship(), modifiedProperties))
			}

			if deletedProperties := properties.DeletedProperties(); len(deletedProperties) > 0 {
				updateStatements = append(updateStatements, query.DeleteProperties(query.Relationship(), deletedProperties...))
			}

			return updateStatements
		}),
	))

	if err := queryBuilder.Prepare(); err != nil {
		return err
	} else if cypherQuery, err := queryBuilder.Render(); err != nil {
		return graph.NewError(cypherQuery, err)
	} else {
		return s.runAndLog(cypherQuery, queryBuilder.Parameters, 1).Error()
	}
}

func (s *neo4jTransaction) TraversalMemoryLimit() size.Size {
	return s.traversalMemoryLimit
}
