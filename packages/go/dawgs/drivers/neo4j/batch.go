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
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

type createRelationshipByIDs struct {
	startID    graph.ID
	endID      graph.ID
	kind       graph.Kind
	properties *graph.Properties
}

type batchTransaction struct {
	innerTx                    *neo4jTransaction
	nodeDeletionBuffer         []graph.ID
	relationshipDeletionBuffer []graph.ID
	nodeUpdateByBuffer         []graph.NodeUpdate
	relationshipCreateBuffer   []createRelationshipByIDs
	relationshipUpdateByBuffer []graph.RelationshipUpdate
	batchWriteSize             int
}

func (s *batchTransaction) CreateNode(node *graph.Node) error {
	_, err := s.innerTx.CreateNode(node.Properties, node.Kinds...)
	return err
}

func (s *batchTransaction) CreateRelationship(relationship *graph.Relationship) error {
	return s.CreateRelationshipByIDs(relationship.StartID, relationship.EndID, relationship.Kind, relationship.Properties)
}

func (s *batchTransaction) WithGraph(graphSchema graph.Graph) graph.Batch {
	return s
}

func (s *batchTransaction) Nodes() graph.NodeQuery {
	return NewNodeQuery(s.innerTx.ctx, s)
}

func (s *batchTransaction) Relationships() graph.RelationshipQuery {
	return NewRelationshipQuery(s.innerTx.ctx, s)
}

func (s *batchTransaction) UpdateNodeBy(update graph.NodeUpdate) error {
	if s.nodeUpdateByBuffer = append(s.nodeUpdateByBuffer, update); len(s.nodeUpdateByBuffer) >= s.batchWriteSize {
		return s.flushNodeUpdates()
	}

	return nil
}

func (s *batchTransaction) UpdateRelationshipBy(update graph.RelationshipUpdate) error {
	if s.relationshipUpdateByBuffer = append(s.relationshipUpdateByBuffer, update); len(s.relationshipUpdateByBuffer) >= s.batchWriteSize {
		return s.flushRelationshipUpdates()
	}

	return nil
}

func (s *batchTransaction) DeleteNodes(ids []graph.ID) error {
	return s.innerTx.DeleteNodesBySlice(ids)
}

func (s *batchTransaction) DeleteRelationships(ids []graph.ID) error {
	return s.innerTx.DeleteRelationshipsBySlice(ids)
}

func (s *batchTransaction) Commit() error {
	if len(s.nodeUpdateByBuffer) > 0 {
		if err := s.flushNodeUpdates(); err != nil {
			return err
		}
	}

	if len(s.relationshipCreateBuffer) > 0 {
		if err := s.flushRelationshipCreation(); err != nil {
			return err
		}
	}

	if len(s.relationshipUpdateByBuffer) > 0 {
		if err := s.flushRelationshipUpdates(); err != nil {
			return err
		}
	}

	if len(s.nodeDeletionBuffer) > 0 {
		if err := s.flushNodeDeletions(); err != nil {
			return err
		}
	}

	if len(s.relationshipDeletionBuffer) > 0 {
		if err := s.flushRelationshipDeletions(); err != nil {
			return err
		}
	}

	return s.innerTx.Commit()
}

func (s *batchTransaction) Close() error {
	return s.innerTx.Close()
}

func (s *batchTransaction) UpdateNode(target *graph.Node) error {
	return s.innerTx.UpdateNode(target)
}

func (s *batchTransaction) CreateRelationshipByIDs(startNodeID, endNodeID graph.ID, kind graph.Kind, properties *graph.Properties) error {
	nextUpdate := createRelationshipByIDs{
		startID:    startNodeID,
		endID:      endNodeID,
		kind:       kind,
		properties: properties,
	}

	if s.relationshipCreateBuffer = append(s.relationshipCreateBuffer, nextUpdate); len(s.relationshipCreateBuffer) >= s.batchWriteSize {
		return s.flushRelationshipCreation()
	}

	return nil
}

func (s *batchTransaction) DeleteNode(id graph.ID) error {
	if s.nodeDeletionBuffer = append(s.nodeDeletionBuffer, id); len(s.nodeDeletionBuffer) >= s.batchWriteSize {
		return s.flushNodeDeletions()
	}

	return nil
}

func (s *batchTransaction) DeleteRelationship(id graph.ID) error {
	if s.relationshipDeletionBuffer = append(s.relationshipDeletionBuffer, id); len(s.relationshipDeletionBuffer) >= s.batchWriteSize {
		return s.flushRelationshipDeletions()
	}

	return nil
}

func (s *batchTransaction) UpdateRelationship(relationship *graph.Relationship) error {
	return s.innerTx.UpdateRelationship(relationship)
}

func (s *batchTransaction) Raw(cypher string, params map[string]any) graph.Result {
	return s.innerTx.Raw(cypher, params)
}

type relationshipCreateByIDBatch struct {
	numRelationships int
	queryParameters  map[string]any
}

func cypherBuildRelationshipCreateByIDBatch(updates []createRelationshipByIDs) ([]string, []relationshipCreateByIDBatch) {
	var (
		queries         []string
		queryParameters []relationshipCreateByIDBatch

		output           = strings.Builder{}
		updatesByRelKind = map[graph.Kind][]createRelationshipByIDs{}
	)

	for _, update := range updates {
		updatesByRelKind[update.kind] = append(updatesByRelKind[update.kind], update)
	}

	for kind, batchJobs := range updatesByRelKind {
		output.WriteString("unwind $p as p match (s) where id(s) = p.s match(e) where id(e) = p.e merge (s)-[r:")
		output.WriteString(kind.String())
		output.WriteString("]->(e) set r += p.p")

		nextQueryParameters := make([]map[string]any, len(batchJobs))

		for idx, batchJob := range batchJobs {
			nextQueryParameters[idx] = map[string]any{
				"s": batchJob.startID,
				"e": batchJob.endID,
				"p": batchJob.properties.Map,
			}
		}

		queries = append(queries, output.String())
		queryParameters = append(queryParameters, relationshipCreateByIDBatch{
			numRelationships: len(nextQueryParameters),
			queryParameters: map[string]any{
				"p": nextQueryParameters,
			},
		})

		output.Reset()
	}

	return queries, queryParameters
}

func (s *batchTransaction) flushRelationshipCreation() error {
	statements, batches := cypherBuildRelationshipCreateByIDBatch(s.relationshipCreateBuffer)

	for parameterIdx, statement := range statements {
		nextBatch := batches[parameterIdx]

		if result := s.innerTx.runAndLog(statement, nextBatch.queryParameters, nextBatch.numRelationships); result.Error() != nil {
			return result.Error()
		}
	}

	s.relationshipCreateBuffer = s.relationshipCreateBuffer[:0]
	return nil
}

func (s *batchTransaction) flushRelationshipDeletions() error {
	buffer := s.relationshipDeletionBuffer
	s.relationshipDeletionBuffer = s.relationshipDeletionBuffer[:0]

	return s.DeleteRelationships(buffer)
}

func (s *batchTransaction) flushNodeUpdates() error {
	buffer := s.nodeUpdateByBuffer
	s.nodeUpdateByBuffer = s.nodeUpdateByBuffer[:0]

	return s.innerTx.updateNodesBy(buffer...)
}

func (s *batchTransaction) flushRelationshipUpdates() error {
	buffer := s.relationshipUpdateByBuffer
	s.relationshipUpdateByBuffer = s.relationshipUpdateByBuffer[:0]

	return s.innerTx.updateRelationshipsBy(buffer...)
}

func (s *batchTransaction) flushNodeDeletions() error {
	buffer := s.nodeDeletionBuffer
	s.nodeDeletionBuffer = s.nodeDeletionBuffer[:0]

	return s.innerTx.DeleteNodesBySlice(buffer)
}

func newBatchOperation(ctx context.Context, session neo4j.Session, cfg graph.TransactionConfig, writeFlushSize int, batchWriteSize int, traversalMemoryLimit size.Size) *batchTransaction {
	return &batchTransaction{
		innerTx:                    newTransaction(ctx, session, cfg, writeFlushSize, batchWriteSize, traversalMemoryLimit),
		batchWriteSize:             batchWriteSize,
		nodeDeletionBuffer:         make([]graph.ID, 0, batchWriteSize),
		relationshipDeletionBuffer: make([]graph.ID, 0, batchWriteSize),
		nodeUpdateByBuffer:         make([]graph.NodeUpdate, 0, batchWriteSize),
		relationshipUpdateByBuffer: make([]graph.RelationshipUpdate, 0, batchWriteSize),
	}
}
