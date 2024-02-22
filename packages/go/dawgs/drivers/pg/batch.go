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

package pg

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cypher/backend/pgsql"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	sql "github.com/specterops/bloodhound/dawgs/drivers/pg/query"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"strconv"
	"strings"
	"sync/atomic"
)

type Int2ArrayEncoder struct {
	buffer *bytes.Buffer
}

func (s *Int2ArrayEncoder) Encode(values []int16) string {
	s.buffer.Reset()
	s.buffer.WriteRune('{')

	for idx, value := range values {
		if idx > 0 {
			s.buffer.WriteRune(',')
		}

		s.buffer.WriteString(strconv.Itoa(int(value)))
	}

	s.buffer.WriteRune('}')
	return s.buffer.String()
}

type batch struct {
	ctx                        context.Context
	innerTransaction           *transaction
	schemaManager              *SchemaManager
	nodeDeletionBuffer         []graph.ID
	relationshipDeletionBuffer []graph.ID
	nodeCreateBuffer           []*graph.Node
	nodeUpdateByBuffer         []graph.NodeUpdate
	relationshipCreateBuffer   []*graph.Relationship
	relationshipUpdateByBuffer []graph.RelationshipUpdate
	batchWriteSize             int
	kindIDEncoder              Int2ArrayEncoder
}

func newBatch(ctx context.Context, conn *pgxpool.Conn, schemaManager *SchemaManager, cfg *Config) (*batch, error) {
	if tx, err := newTransaction(ctx, conn, schemaManager, cfg); err != nil {
		return nil, err
	} else {
		return &batch{
			ctx:              ctx,
			schemaManager:    schemaManager,
			innerTransaction: tx,
			batchWriteSize:   cfg.BatchWriteSize,
			kindIDEncoder: Int2ArrayEncoder{
				buffer: &bytes.Buffer{},
			},
		}, nil
	}
}

func (s *batch) WithGraph(schema graph.Graph) graph.Batch {
	s.innerTransaction.WithGraph(schema)
	return s
}

func (s *batch) CreateNode(node *graph.Node) error {
	s.nodeCreateBuffer = append(s.nodeCreateBuffer, node)
	return s.tryFlush(s.batchWriteSize)
}

func (s *batch) Nodes() graph.NodeQuery {
	return s.innerTransaction.Nodes()
}

func (s *batch) Relationships() graph.RelationshipQuery {
	return s.innerTransaction.Relationships()
}

func (s *batch) UpdateNodeBy(update graph.NodeUpdate) error {
	s.nodeUpdateByBuffer = append(s.nodeUpdateByBuffer, update)
	return s.tryFlush(s.batchWriteSize)
}

func (s *batch) flushNodeDeleteBuffer() error {
	if _, err := s.innerTransaction.tx.Exec(s.ctx, deleteNodeWithIDStatement, s.nodeDeletionBuffer); err != nil {
		return err
	}

	s.nodeDeletionBuffer = s.nodeDeletionBuffer[:0]
	return nil
}

func (s *batch) flushRelationshipDeleteBuffer() error {
	if _, err := s.innerTransaction.tx.Exec(s.ctx, deleteEdgeWithIDStatement, s.relationshipDeletionBuffer); err != nil {
		return err
	}

	s.relationshipDeletionBuffer = s.relationshipDeletionBuffer[:0]
	return nil
}

func (s *batch) flushNodeCreateBuffer() error {
	var (
		withoutIDs = false
		withIDs    = false
	)

	for _, node := range s.nodeCreateBuffer {
		if node.ID == 0 || node.ID == graph.UnregisteredNodeID {
			withoutIDs = true
		} else {
			withIDs = true
		}

		if withIDs && withoutIDs {
			return fmt.Errorf("batch may not mix preset node IDs with entries that require an auto-generated ID")
		}
	}

	if withoutIDs {
		return s.flushNodeCreateBufferWithoutIDs()
	}

	return s.flushNodeCreateBufferWithIDs()
}

func (s *batch) flushNodeCreateBufferWithIDs() error {
	var (
		numCreates    = len(s.nodeCreateBuffer)
		nodeIDs       = make([]uint32, numCreates)
		kindIDSlices  = make([]string, numCreates)
		kindIDEncoder = Int2ArrayEncoder{
			buffer: &bytes.Buffer{},
		}
		properties = make([]pgtype.JSONB, numCreates)
	)

	for idx, nextNode := range s.nodeCreateBuffer {
		nodeIDs[idx] = nextNode.ID.Uint32()

		if mappedKindIDs, missingKinds := s.schemaManager.MapKinds(nextNode.Kinds); len(missingKinds) > 0 {
			return fmt.Errorf("unable to map kinds %v", missingKinds)
		} else {
			kindIDSlices[idx] = kindIDEncoder.Encode(mappedKindIDs)
		}

		if propertiesJSONB, err := pgsql.PropertiesToJSONB(nextNode.Properties); err != nil {
			return err
		} else {
			properties[idx] = propertiesJSONB
		}
	}

	if graphTarget, err := s.innerTransaction.getTargetGraph(); err != nil {
		return err
	} else if _, err := s.innerTransaction.tx.Exec(s.ctx, createNodeWithIDBatchStatement, graphTarget.ID, nodeIDs, kindIDSlices, properties); err != nil {
		return err
	}

	s.nodeCreateBuffer = s.nodeCreateBuffer[:0]
	return nil
}

func (s *batch) flushNodeCreateBufferWithoutIDs() error {
	var (
		numCreates    = len(s.nodeCreateBuffer)
		kindIDSlices  = make([]string, numCreates)
		kindIDEncoder = Int2ArrayEncoder{
			buffer: &bytes.Buffer{},
		}
		properties = make([]pgtype.JSONB, numCreates)
	)

	for idx, nextNode := range s.nodeCreateBuffer {
		if mappedKindIDs, missingKinds := s.schemaManager.MapKinds(nextNode.Kinds); len(missingKinds) > 0 {
			return fmt.Errorf("unable to map kinds %v", missingKinds)
		} else {
			kindIDSlices[idx] = kindIDEncoder.Encode(mappedKindIDs)
		}

		if propertiesJSONB, err := pgsql.PropertiesToJSONB(nextNode.Properties); err != nil {
			return err
		} else {
			properties[idx] = propertiesJSONB
		}
	}

	if graphTarget, err := s.innerTransaction.getTargetGraph(); err != nil {
		return err
	} else if _, err := s.innerTransaction.tx.Exec(s.ctx, createNodeWithoutIDBatchStatement, graphTarget.ID, kindIDSlices, properties); err != nil {
		return err
	}

	s.nodeCreateBuffer = s.nodeCreateBuffer[:0]
	return nil
}

func (s *batch) flushNodeUpsertBatch(updates *sql.NodeUpdateBatch) error {
	parameters := NewNodeUpsertParameters(len(updates.Updates))

	if err := parameters.AppendAll(updates, s.schemaManager, s.kindIDEncoder); err != nil {
		return err
	}

	if graphTarget, err := s.innerTransaction.getTargetGraph(); err != nil {
		return err
	} else {
		query := sql.FormatNodeUpsert(graphTarget, updates.IdentityProperties)

		if rows, err := s.innerTransaction.tx.Query(s.ctx, query, parameters.Format(graphTarget)...); err != nil {
			return err
		} else {
			defer rows.Close()

			idFutureIndex := 0

			for rows.Next() {
				if err := rows.Scan(&parameters.IDFutures[idFutureIndex].Value); err != nil {
					return err
				}

				idFutureIndex++
			}
		}
	}

	return nil
}

func (s *batch) tryFlushNodeUpdateByBuffer() error {
	if updates, err := sql.ValidateNodeUpdateByBatch(s.nodeUpdateByBuffer); err != nil {
		return err
	} else if err := s.flushNodeUpsertBatch(updates); err != nil {
		return err
	}

	s.nodeUpdateByBuffer = s.nodeUpdateByBuffer[:0]
	return nil
}

type NodeUpsertParameters struct {
	IDFutures    []*sql.Future[graph.ID]
	KindIDSlices []string
	Properties   []pgtype.JSONB
}

func NewNodeUpsertParameters(size int) *NodeUpsertParameters {
	return &NodeUpsertParameters{
		IDFutures:    make([]*sql.Future[graph.ID], 0, size),
		KindIDSlices: make([]string, 0, size),
		Properties:   make([]pgtype.JSONB, 0, size),
	}
}

func (s *NodeUpsertParameters) Format(graphTarget model.Graph) []any {
	return []any{
		graphTarget.ID,
		s.KindIDSlices,
		s.Properties,
	}
}

func (s *NodeUpsertParameters) Append(update *sql.NodeUpdate, schemaManager *SchemaManager, kindIDEncoder Int2ArrayEncoder) error {
	s.IDFutures = append(s.IDFutures, update.IDFuture)

	if mappedKindIDs, missingKinds := schemaManager.MapKinds(update.Node.Kinds); len(missingKinds) > 0 {
		return fmt.Errorf("unable to map kinds %v", missingKinds)
	} else {
		s.KindIDSlices = append(s.KindIDSlices, kindIDEncoder.Encode(mappedKindIDs))
	}

	if propertiesJSONB, err := pgsql.PropertiesToJSONB(update.Node.Properties); err != nil {
		return err
	} else {
		s.Properties = append(s.Properties, propertiesJSONB)
	}

	return nil
}

func (s *NodeUpsertParameters) AppendAll(updates *sql.NodeUpdateBatch, schemaManager *SchemaManager, kindIDEncoder Int2ArrayEncoder) error {
	for _, nextUpdate := range updates.Updates {
		if err := s.Append(nextUpdate, schemaManager, kindIDEncoder); err != nil {
			return err
		}
	}

	return nil
}

type RelationshipUpdateByParameters struct {
	StartIDs   []graph.ID
	EndIDs     []graph.ID
	KindIDs    []int16
	Properties []pgtype.JSONB
}

func NewRelationshipUpdateByParameters(size int) *RelationshipUpdateByParameters {
	return &RelationshipUpdateByParameters{
		StartIDs:   make([]graph.ID, 0, size),
		EndIDs:     make([]graph.ID, 0, size),
		KindIDs:    make([]int16, 0, size),
		Properties: make([]pgtype.JSONB, 0, size),
	}
}

func (s *RelationshipUpdateByParameters) Format(graphTarget model.Graph) []any {
	return []any{
		graphTarget.ID,
		s.StartIDs,
		s.EndIDs,
		s.KindIDs,
		s.Properties,
	}
}

func (s *RelationshipUpdateByParameters) Append(update *sql.RelationshipUpdate, schemaManager *SchemaManager) error {
	s.StartIDs = append(s.StartIDs, update.StartID.Value)
	s.EndIDs = append(s.EndIDs, update.EndID.Value)

	if mappedKindID, mapped := schemaManager.MapKind(update.Relationship.Kind); !mapped {
		return fmt.Errorf("unable to map kind %s", update.Relationship.Kind)
	} else {
		s.KindIDs = append(s.KindIDs, mappedKindID)
	}

	if propertiesJSONB, err := pgsql.PropertiesToJSONB(update.Relationship.Properties); err != nil {
		return err
	} else {
		s.Properties = append(s.Properties, propertiesJSONB)
	}
	return nil
}

func (s *RelationshipUpdateByParameters) AppendAll(updates *sql.RelationshipUpdateBatch, schemaManager *SchemaManager) error {
	for _, nextUpdate := range updates.Updates {
		if err := s.Append(nextUpdate, schemaManager); err != nil {
			return err
		}
	}

	return nil
}

var numRels = &atomic.Int64{}

func (s *batch) flushRelationshipUpdateByBuffer(updates *sql.RelationshipUpdateBatch) error {
	if err := s.flushNodeUpsertBatch(updates.NodeUpdates); err != nil {
		return err
	}

	parameters := NewRelationshipUpdateByParameters(len(updates.Updates))

	if err := parameters.AppendAll(updates, s.schemaManager); err != nil {
		return err
	}

	numRels.Add(int64(len(parameters.Properties)))

	if graphTarget, err := s.innerTransaction.getTargetGraph(); err != nil {
		return err
	} else {
		query := sql.FormatRelationshipPartitionUpsert(graphTarget)

		if _, err := s.innerTransaction.tx.Exec(s.ctx, query, parameters.Format(graphTarget)...); err != nil {
			return err
		}
	}

	return nil
}

func (s *batch) tryFlushRelationshipUpdateByBuffer() error {
	if updateBatch, err := sql.ValidateRelationshipUpdateByBatch(s.relationshipUpdateByBuffer); err != nil {
		return err
	} else if err := s.flushRelationshipUpdateByBuffer(updateBatch); err != nil {
		return err
	}

	s.relationshipUpdateByBuffer = s.relationshipUpdateByBuffer[:0]
	return nil
}

type relationshipCreateBatch struct {
	startIDs         []uint32
	endIDs           []uint32
	edgeKindIDs      []int16
	edgePropertyBags []pgtype.JSONB
}

func newRelationshipCreateBatch(size int) *relationshipCreateBatch {
	return &relationshipCreateBatch{
		startIDs:         make([]uint32, 0, size),
		endIDs:           make([]uint32, 0, size),
		edgeKindIDs:      make([]int16, 0, size),
		edgePropertyBags: make([]pgtype.JSONB, 0, size),
	}
}

func (s *relationshipCreateBatch) Add(startID, endID uint32, edgeKindID int16) {
	s.startIDs = append(s.startIDs, startID)
	s.edgeKindIDs = append(s.edgeKindIDs, edgeKindID)
	s.endIDs = append(s.endIDs, endID)
}

func (s *relationshipCreateBatch) EncodeProperties(edgePropertiesBatch []*graph.Properties) error {
	for _, edgeProperties := range edgePropertiesBatch {
		if propertiesJSONB, err := pgsql.PropertiesToJSONB(edgeProperties); err != nil {
			return err
		} else {
			s.edgePropertyBags = append(s.edgePropertyBags, propertiesJSONB)
		}
	}

	return nil
}

type relationshipCreateBatchBuilder struct {
	keyToEdgeID             map[string]uint32
	relationshipUpdateBatch *relationshipCreateBatch
	edgePropertiesIndex     map[uint32]int
	edgePropertiesBatch     []*graph.Properties
}

func newRelationshipCreateBatchBuilder(size int) *relationshipCreateBatchBuilder {
	return &relationshipCreateBatchBuilder{
		keyToEdgeID:             map[string]uint32{},
		relationshipUpdateBatch: newRelationshipCreateBatch(size),
		edgePropertiesIndex:     map[uint32]int{},
	}
}

func (s *relationshipCreateBatchBuilder) Build() (*relationshipCreateBatch, error) {
	return s.relationshipUpdateBatch, s.relationshipUpdateBatch.EncodeProperties(s.edgePropertiesBatch)
}

func (s *relationshipCreateBatchBuilder) Add(kindMapper KindMapper, edge *graph.Relationship) error {
	keyBuilder := strings.Builder{}

	keyBuilder.WriteString(edge.StartID.String())
	keyBuilder.WriteString(edge.EndID.String())
	keyBuilder.WriteString(edge.Kind.String())

	key := keyBuilder.String()

	if existingPropertiesIdx, hasExisting := s.keyToEdgeID[key]; hasExisting {
		s.edgePropertiesBatch[existingPropertiesIdx].Merge(edge.Properties)
	} else {
		var (
			startID        = edge.StartID.Uint32()
			edgeID         = edge.ID.Uint32()
			endID          = edge.EndID.Uint32()
			edgeProperties = edge.Properties.Clone()
		)

		if edgeKindID, hasKind := kindMapper.MapKind(edge.Kind); !hasKind {
			return fmt.Errorf("unable to map kind %s", edge.Kind)
		} else {
			s.relationshipUpdateBatch.Add(startID, endID, edgeKindID)
		}

		s.keyToEdgeID[key] = edgeID

		s.edgePropertiesBatch = append(s.edgePropertiesBatch, edgeProperties)
		s.edgePropertiesIndex[edgeID] = len(s.edgePropertiesBatch) - 1
	}

	return nil
}

func (s *batch) flushRelationshipCreateBuffer() error {
	batchBuilder := newRelationshipCreateBatchBuilder(len(s.relationshipCreateBuffer))

	for _, nextRel := range s.relationshipCreateBuffer {
		if err := batchBuilder.Add(s.schemaManager, nextRel); err != nil {
			return err
		}
	}

	if createBatch, err := batchBuilder.Build(); err != nil {
		return err
	} else if graphTarget, err := s.innerTransaction.getTargetGraph(); err != nil {
		return err
	} else if _, err := s.innerTransaction.tx.Exec(s.ctx, createEdgeBatchStatement, graphTarget.ID, createBatch.startIDs, createBatch.endIDs, createBatch.edgeKindIDs, createBatch.edgePropertyBags); err != nil {
		log.Infof("Num merged property bags: %d - Num edge keys: %d - StartID batch size: %d", len(batchBuilder.edgePropertiesIndex), len(batchBuilder.keyToEdgeID), len(batchBuilder.relationshipUpdateBatch.startIDs))
		return err
	}

	s.relationshipCreateBuffer = s.relationshipCreateBuffer[:0]
	return nil
}

func (s *batch) tryFlush(batchWriteSize int) error {
	if len(s.nodeUpdateByBuffer) > batchWriteSize {
		if err := s.tryFlushNodeUpdateByBuffer(); err != nil {
			return err
		}
	}

	if len(s.relationshipUpdateByBuffer) > batchWriteSize {
		if err := s.tryFlushRelationshipUpdateByBuffer(); err != nil {
			return err
		}
	}

	if len(s.relationshipCreateBuffer) > batchWriteSize {
		if err := s.flushRelationshipCreateBuffer(); err != nil {
			return err
		}
	}

	if len(s.nodeCreateBuffer) > batchWriteSize {
		if err := s.flushNodeCreateBuffer(); err != nil {
			return err
		}
	}

	if len(s.nodeDeletionBuffer) > batchWriteSize {
		if err := s.flushNodeDeleteBuffer(); err != nil {
			return err
		}
	}

	if len(s.relationshipDeletionBuffer) > batchWriteSize {
		if err := s.flushRelationshipDeleteBuffer(); err != nil {
			return err
		}
	}

	return nil
}

func (s *batch) CreateRelationship(relationship *graph.Relationship) error {
	s.relationshipCreateBuffer = append(s.relationshipCreateBuffer, relationship)
	return s.tryFlush(s.batchWriteSize)
}

func (s *batch) CreateRelationshipByIDs(startNodeID, endNodeID graph.ID, kind graph.Kind, properties *graph.Properties) error {
	return s.CreateRelationship(&graph.Relationship{
		StartID:    startNodeID,
		EndID:      endNodeID,
		Kind:       kind,
		Properties: properties,
	})
}

func (s *batch) UpdateRelationshipBy(update graph.RelationshipUpdate) error {
	s.relationshipUpdateByBuffer = append(s.relationshipUpdateByBuffer, update)
	return s.tryFlush(s.batchWriteSize)
}

func (s *batch) Commit() error {
	if err := s.tryFlush(0); err != nil {
		return err
	}

	return s.innerTransaction.Commit()
}

func (s *batch) DeleteNode(id graph.ID) error {
	s.nodeDeletionBuffer = append(s.nodeDeletionBuffer, id)
	return s.tryFlush(s.batchWriteSize)
}

func (s *batch) DeleteRelationship(id graph.ID) error {
	s.relationshipDeletionBuffer = append(s.relationshipDeletionBuffer, id)
	return s.tryFlush(s.batchWriteSize)
}

func (s *batch) Close() {
	s.innerTransaction.Close()
}
