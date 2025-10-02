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

package graphopsreplaylog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

// Service provides operations for logging and retrieving graph modification history.
// This service acts as an intermediary between HTTP handlers and both the replay log database
// and the graph database, ensuring all changes are properly logged before execution.
type Service interface {
	// Node operations
	CreateNode(ctx context.Context, objectID string, labels []string, properties map[string]interface{}) error
	DeleteNode(ctx context.Context, objectID string) error

	// Edge operations
	CreateEdge(ctx context.Context, sourceObjectID, targetObjectID, edgeKind string, properties map[string]interface{}) error
	DeleteEdge(ctx context.Context, sourceObjectID, targetObjectID, edgeKind string) error

	// Update operations (stubbed for future implementation)
	UpdateNode(ctx context.Context, objectID string, labels []string, properties map[string]interface{}) error
	UpdateEdge(ctx context.Context, sourceObjectID, targetObjectID, edgeKind string, properties map[string]interface{}) error

	// Retrieve replay log entries
	GetRecentChanges(ctx context.Context, limit int) ([]model.GraphOperationReplayLogEntry, error)

	// Analysis operations
	LogAnalysisStart(ctx context.Context, requestedBy string) error
	LogAnalysisEnd(ctx context.Context, requestedBy string) error

	// Time travel operations
	RollToEntry(ctx context.Context, targetEntryID int64) error
}

// service implements the Service interface
type service struct {
	db      *database.BloodhoundDB
	graphDB graph.Database
}

// NewService creates a new graph operations replay log service.
// This service logs all graph modifications to a database table before applying them to the graph.
func NewService(db *database.BloodhoundDB, graphDB graph.Database) Service {
	return &service{
		db:      db,
		graphDB: graphDB,
	}
}

// CreateNode creates a new node in the graph and logs the operation.
// The replay log entry is persisted first to ensure we have a record even if the graph operation fails.
func (s *service) CreateNode(ctx context.Context, objectID string, labels []string, properties map[string]interface{}) error {
	// Log the operation to the replay log first
	if err := s.logChange(ctx, model.ChangeTypeCreate, model.ObjectTypeNode, objectID, labels, "", "", properties); err != nil {
		return fmt.Errorf("failed to log node creation: %w", err)
	}

	// Perform the actual graph operation
	kinds := make(graph.Kinds, len(labels))
	for i, label := range labels {
		kinds[i] = graph.StringKind(label)
	}

	props := properties
	if props == nil {
		props = make(map[string]interface{})
	}
	props[common.ObjectID.String()] = strings.ToUpper(objectID)
	props[common.LastSeen.String()] = time.Now().UTC()

	err := s.graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
		return batch.UpdateNodeBy(graph.NodeUpdate{
			Node:               graph.PrepareNode(graph.AsProperties(props), kinds...),
			IdentityKind:       kinds[0],
			IdentityProperties: []string{common.ObjectID.String()},
		})
	})

	if err != nil {
		return fmt.Errorf("failed to create node in graph: %w", err)
	}

	return nil
}

// DeleteNode removes a node from the graph and logs the operation.
// This also deletes all edges connected to the node (both incoming and outgoing).
// Each edge deletion is logged separately before the node deletion.
func (s *service) DeleteNode(ctx context.Context, objectID string) error {
	// Structure to hold edge info for logging
	type edgeInfo struct {
		sourceObjectID string
		targetObjectID string
		kind           string
		properties     map[string]interface{}
	}

	var (
		edgesToLog     []edgeInfo
		nodeLabels     []string
		nodeProperties map[string]interface{}
	)

	// First, collect node state and edges (read-only transaction)
	err := s.graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		// Find the node
		var targetNode *graph.Node
		err := tx.Nodes().Filter(
			query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
		).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for n := range cursor.Chan() {
				targetNode = n
				break
			}
			return cursor.Error()
		})

		if err != nil {
			return fmt.Errorf("failed to find node: %w", err)
		}
		if targetNode == nil {
			return fmt.Errorf("node not found")
		}

		// Capture node labels for rollback
		for _, kind := range targetNode.Kinds {
			nodeLabels = append(nodeLabels, kind.String())
		}

		// Capture node properties for rollback
		nodeProperties = make(map[string]interface{})
		for k, v := range targetNode.Properties.Map {
			nodeProperties[k] = v
		}

		// Get all relationships (both incoming and outgoing)
		var relationships []*graph.Relationship
		err = tx.Relationships().Filter(
			query.Or(
				query.Equals(query.StartID(), targetNode.ID),
				query.Equals(query.EndID(), targetNode.ID),
			),
		).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for rel := range cursor.Chan() {
				relationships = append(relationships, rel)
			}
			return cursor.Error()
		})

		if err != nil {
			return fmt.Errorf("failed to fetch relationships: %w", err)
		}

		// For each relationship, get the source/target object IDs and properties
		for _, rel := range relationships {
			var sourceNode, targetNode *graph.Node

			// Get source node
			err = tx.Nodes().Filter(query.Equals(query.NodeID(), rel.StartID)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for n := range cursor.Chan() {
					sourceNode = n
					break
				}
				return cursor.Error()
			})
			if err != nil || sourceNode == nil {
				continue // Skip if we can't find source (happy path only)
			}

			// Get target node
			err = tx.Nodes().Filter(query.Equals(query.NodeID(), rel.EndID)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for n := range cursor.Chan() {
					targetNode = n
					break
				}
				return cursor.Error()
			})
			if err != nil || targetNode == nil {
				continue // Skip if we can't find target (happy path only)
			}

			// Get object IDs from nodes
			sourceObjectID, _ := sourceNode.Properties.Get(common.ObjectID.String()).String()
			targetObjectID, _ := targetNode.Properties.Get(common.ObjectID.String()).String()

			// Store edge info for logging (including properties for rollback)
			edgeProps := make(map[string]interface{})
			for k, v := range rel.Properties.Map {
				edgeProps[k] = v
			}
			edgesToLog = append(edgesToLog, edgeInfo{
				sourceObjectID: sourceObjectID,
				targetObjectID: targetObjectID,
				kind:           rel.Kind.String(),
				properties:     edgeProps,
			})
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to read node edges: %w", err)
	}

	// Now log all edge deletions to the replay log (outside the graph transaction)
	for _, edge := range edgesToLog {
		if err := s.logChange(ctx, model.ChangeTypeDelete, model.ObjectTypeEdge, edge.kind, []string{edge.kind}, edge.sourceObjectID, edge.targetObjectID, edge.properties); err != nil {
			// Happy path: log error but continue (best effort)
			// In production you'd want proper error handling
			continue
		}
	}

	// Log the node deletion to the replay log (with full state for rollback)
	if err := s.logChange(ctx, model.ChangeTypeDelete, model.ObjectTypeNode, objectID, nodeLabels, "", "", nodeProperties); err != nil {
		return fmt.Errorf("failed to log node deletion: %w", err)
	}

	// Finally, delete the node from the graph (this will cascade delete edges in Neo4j automatically)
	err = s.graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Filter(
			query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
		).Delete()
	})

	if err != nil {
		return fmt.Errorf("failed to delete node from graph: %w", err)
	}

	return nil
}

// CreateEdge creates a new edge in the graph and logs the operation.
func (s *service) CreateEdge(ctx context.Context, sourceObjectID, targetObjectID, edgeKind string, properties map[string]interface{}) error {
	// Log the operation to the replay log first
	if err := s.logChange(ctx, model.ChangeTypeCreate, model.ObjectTypeEdge, edgeKind, []string{edgeKind}, sourceObjectID, targetObjectID, properties); err != nil {
		return fmt.Errorf("failed to log edge creation: %w", err)
	}

	// Perform the actual graph operation
	props := properties
	if props == nil {
		props = make(map[string]interface{})
	}
	props[common.LastSeen.String()] = time.Now().UTC()

	err := s.graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
		var src, tgt *graph.Node

		findNode := func(objectID string) (*graph.Node, error) {
			var node *graph.Node
			err := batch.Nodes().Filter(
				query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
			).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for n := range cursor.Chan() {
					node = n
					break
				}
				return cursor.Error()
			})
			return node, err
		}

		var err error
		if src, err = findNode(sourceObjectID); err != nil || src == nil {
			return fmt.Errorf("source node not found")
		}
		if tgt, err = findNode(targetObjectID); err != nil || tgt == nil {
			return fmt.Errorf("target node not found")
		}

		return batch.UpdateRelationshipBy(graph.RelationshipUpdate{
			Relationship:            graph.PrepareRelationship(graph.AsProperties(props), graph.StringKind(edgeKind)),
			Start:                   graph.PrepareNode(graph.AsProperties(graph.PropertyMap{common.ObjectID: strings.ToUpper(sourceObjectID), common.LastSeen: time.Now().UTC()}), src.Kinds...),
			StartIdentityKind:       src.Kinds[0],
			StartIdentityProperties: []string{common.ObjectID.String()},
			End:                     graph.PrepareNode(graph.AsProperties(graph.PropertyMap{common.ObjectID: strings.ToUpper(targetObjectID), common.LastSeen: time.Now().UTC()}), tgt.Kinds...),
			EndIdentityKind:         tgt.Kinds[0],
			EndIdentityProperties:   []string{common.ObjectID.String()},
		})
	})

	if err != nil {
		return fmt.Errorf("failed to create edge in graph: %w", err)
	}

	return nil
}

// DeleteEdge removes an edge from the graph and logs the operation.
func (s *service) DeleteEdge(ctx context.Context, sourceObjectID, targetObjectID, edgeKind string) error {
	var edgeProperties map[string]interface{}

	// First, read the edge properties for rollback (read-only transaction)
	err := s.graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var src, tgt *graph.Node

		findNode := func(objectID string) (*graph.Node, error) {
			var node *graph.Node
			err := tx.Nodes().Filter(
				query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
			).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for n := range cursor.Chan() {
					node = n
					break
				}
				return cursor.Error()
			})
			return node, err
		}

		var err error
		if src, err = findNode(sourceObjectID); err != nil || src == nil {
			return fmt.Errorf("source node not found")
		}
		if tgt, err = findNode(targetObjectID); err != nil || tgt == nil {
			return fmt.Errorf("target node not found")
		}

		// Find the edge and capture its properties
		return tx.Relationships().Filter(
			query.And(
				query.Equals(query.StartID(), src.ID),
				query.Equals(query.EndID(), tgt.ID),
				query.KindIn(query.Relationship(), graph.StringKind(edgeKind)),
			),
		).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for rel := range cursor.Chan() {
				edgeProperties = make(map[string]interface{})
				for k, v := range rel.Properties.Map {
					edgeProperties[k] = v
				}
				break
			}
			return cursor.Error()
		})
	})

	if err != nil {
		return fmt.Errorf("failed to read edge properties: %w", err)
	}

	// Log the operation to the replay log with captured properties
	if err := s.logChange(ctx, model.ChangeTypeDelete, model.ObjectTypeEdge, edgeKind, []string{edgeKind}, sourceObjectID, targetObjectID, edgeProperties); err != nil {
		return fmt.Errorf("failed to log edge deletion: %w", err)
	}

	// Perform the actual graph operation (use WriteTransaction for deletes)
	err = s.graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		var src, tgt *graph.Node

		findNode := func(objectID string) (*graph.Node, error) {
			var node *graph.Node
			err := tx.Nodes().Filter(
				query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
			).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for n := range cursor.Chan() {
					node = n
					break
				}
				return cursor.Error()
			})
			return node, err
		}

		var err error
		if src, err = findNode(sourceObjectID); err != nil || src == nil {
			return fmt.Errorf("source node not found")
		}
		if tgt, err = findNode(targetObjectID); err != nil || tgt == nil {
			return fmt.Errorf("target node not found")
		}

		return tx.Relationships().Filter(
			query.And(
				query.Equals(query.StartID(), src.ID),
				query.Equals(query.EndID(), tgt.ID),
				query.KindIn(query.Relationship(), graph.StringKind(edgeKind)),
			),
		).Delete()
	})

	if err != nil {
		return fmt.Errorf("failed to delete edge from graph: %w", err)
	}

	return nil
}

// UpdateNode is a stub for future implementation of node updates.
func (s *service) UpdateNode(ctx context.Context, objectID string, labels []string, properties map[string]interface{}) error {
	return fmt.Errorf("update operations are not yet implemented")
}

// UpdateEdge is a stub for future implementation of edge updates.
func (s *service) UpdateEdge(ctx context.Context, sourceObjectID, targetObjectID, edgeKind string, properties map[string]interface{}) error {
	return fmt.Errorf("update operations are not yet implemented")
}

// GetRecentChanges retrieves the most recent replay log entries.
// The limit parameter controls how many entries to retrieve (max 100 for this POC).
func (s *service) GetRecentChanges(ctx context.Context, limit int) ([]model.GraphOperationReplayLogEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	var entries []model.GraphOperationReplayLogEntry
	result := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&entries)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve replay log entries: %w", result.Error)
	}

	return entries, nil
}

// logChange writes a replay log entry to the database.
// This is an internal helper method that persists the operation details before execution.
func (s *service) logChange(
	ctx context.Context,
	changeType model.ChangeType,
	objectType model.ObjectType,
	objectID string,
	labels []string,
	sourceObjectID string,
	targetObjectID string,
	properties map[string]interface{},
) error {
	entry := model.GraphOperationReplayLogEntry{
		ChangeType: changeType,
		ObjectType: objectType,
		ObjectID:   objectID,
	}

	// Serialize labels to JSON
	if labels != nil && len(labels) > 0 {
		labelsJSON, err := json.Marshal(labels)
		if err != nil {
			return fmt.Errorf("failed to marshal labels: %w", err)
		}
		entry.Labels = labelsJSON
	}

	// Set edge-specific fields
	if objectType == model.ObjectTypeEdge {
		entry.SourceObjectID = null.StringFrom(sourceObjectID)
		entry.TargetObjectID = null.StringFrom(targetObjectID)
	}

	// Serialize properties to JSON (always set to avoid NULL constraint violation)
	if properties != nil && len(properties) > 0 {
		propsJSON, err := json.Marshal(properties)
		if err != nil {
			return fmt.Errorf("failed to marshal properties: %w", err)
		}
		entry.Properties = propsJSON
	} else {
		// Set empty JSON object for operations without properties (like deletes)
		entry.Properties = []byte("{}")
	}

	// Persist to database
	result := s.db.WithContext(ctx).Create(&entry)

	if result.Error != nil {
		return fmt.Errorf("failed to create replay log entry: %w", result.Error)
	}

	return nil
}

// LogAnalysisStart logs the start of an analysis run to the replay log.
// This creates a marker in the timeline showing when analysis was initiated.
func (s *service) LogAnalysisStart(ctx context.Context, requestedBy string) error {
	properties := map[string]interface{}{
		"requested_by": requestedBy,
	}

	return s.logChange(ctx, model.ChangeTypeAnalysisStart, model.ObjectTypeAnalysis, "analysis", nil, "", "", properties)
}

// LogAnalysisEnd logs the completion of an analysis run to the replay log.
// This creates a marker in the timeline showing when analysis finished.
func (s *service) LogAnalysisEnd(ctx context.Context, requestedBy string) error {
	properties := map[string]interface{}{
		"requested_by": requestedBy,
	}

	return s.logChange(ctx, model.ChangeTypeAnalysisEnd, model.ObjectTypeAnalysis, "analysis", nil, "", "", properties)
}

// RollToEntry moves the graph state to a specific entry ID (logical clock).
// Automatically determines direction (backward or forward) and applies necessary changes.
// Analysis events are skipped as they are no-ops.
func (s *service) RollToEntry(ctx context.Context, targetEntryID int64) error {
	var activeEntries []model.GraphOperationReplayLogEntry
	var rolledBackEntries []model.GraphOperationReplayLogEntry

	// Get entries after target (to potentially roll back)
	if err := s.db.WithContext(ctx).
		Where("id > ? AND rolled_back_at IS NULL", targetEntryID).
		Order("id DESC").
		Find(&activeEntries).Error; err != nil {
		return fmt.Errorf("failed to fetch active entries: %w", err)
	}

	// Get rolled-back entries before target (to potentially re-apply)
	if err := s.db.WithContext(ctx).
		Where("id <= ? AND rolled_back_at IS NOT NULL", targetEntryID).
		Order("id ASC").
		Find(&rolledBackEntries).Error; err != nil {
		return fmt.Errorf("failed to fetch rolled-back entries: %w", err)
	}

	// Roll backward: undo changes after target
	for _, entry := range activeEntries {
		if entry.ObjectType == model.ObjectTypeAnalysis {
			continue
		}
		if err := s.applyInverse(ctx, entry); err != nil {
			return fmt.Errorf("failed to invert entry %d: %w", entry.ID, err)
		}
		if err := s.db.WithContext(ctx).Model(&entry).Update("rolled_back_at", time.Now()).Error; err != nil {
			return fmt.Errorf("failed to mark entry %d as rolled back: %w", entry.ID, err)
		}
	}

	// Roll forward: re-apply rolled-back changes before target
	for _, entry := range rolledBackEntries {
		if entry.ObjectType == model.ObjectTypeAnalysis {
			continue
		}
		if err := s.applyDirect(ctx, entry); err != nil {
			return fmt.Errorf("failed to re-apply entry %d: %w", entry.ID, err)
		}
		if err := s.db.WithContext(ctx).Model(&entry).Update("rolled_back_at", nil).Error; err != nil {
			return fmt.Errorf("failed to clear rollback status for entry %d: %w", entry.ID, err)
		}
	}

	return nil
}

// applyInverse inverts an operation (create→delete, delete→create)
func (s *service) applyInverse(ctx context.Context, entry model.GraphOperationReplayLogEntry) error {
	switch {
	case entry.ChangeType == model.ChangeTypeCreate && entry.ObjectType == model.ObjectTypeNode:
		return s.deleteNodeDirect(ctx, entry.ObjectID)
	case entry.ChangeType == model.ChangeTypeDelete && entry.ObjectType == model.ObjectTypeNode:
		return s.createNodeDirect(ctx, entry)
	case entry.ChangeType == model.ChangeTypeCreate && entry.ObjectType == model.ObjectTypeEdge:
		return s.deleteEdgeDirect(ctx, entry)
	case entry.ChangeType == model.ChangeTypeDelete && entry.ObjectType == model.ObjectTypeEdge:
		return s.createEdgeDirect(ctx, entry)
	}
	return nil
}

// applyDirect re-applies an operation as originally recorded
func (s *service) applyDirect(ctx context.Context, entry model.GraphOperationReplayLogEntry) error {
	switch {
	case entry.ChangeType == model.ChangeTypeCreate && entry.ObjectType == model.ObjectTypeNode:
		return s.createNodeDirect(ctx, entry)
	case entry.ChangeType == model.ChangeTypeDelete && entry.ObjectType == model.ObjectTypeNode:
		return s.deleteNodeDirect(ctx, entry.ObjectID)
	case entry.ChangeType == model.ChangeTypeCreate && entry.ObjectType == model.ObjectTypeEdge:
		return s.createEdgeDirect(ctx, entry)
	case entry.ChangeType == model.ChangeTypeDelete && entry.ObjectType == model.ObjectTypeEdge:
		return s.deleteEdgeDirect(ctx, entry)
	}
	return nil
}

// createNodeDirect creates a node without logging
func (s *service) createNodeDirect(ctx context.Context, entry model.GraphOperationReplayLogEntry) error {
	var labels []string
	if err := json.Unmarshal(entry.Labels, &labels); err != nil {
		return err
	}

	var properties map[string]interface{}
	if err := json.Unmarshal(entry.Properties, &properties); err != nil {
		return err
	}

	kinds := make(graph.Kinds, len(labels))
	for i, label := range labels {
		kinds[i] = graph.StringKind(label)
	}

	props := properties
	if props == nil {
		props = make(map[string]interface{})
	}
	props[common.ObjectID.String()] = strings.ToUpper(entry.ObjectID)

	return s.graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
		return batch.UpdateNodeBy(graph.NodeUpdate{
			Node:               graph.PrepareNode(graph.AsProperties(props), kinds...),
			IdentityKind:       kinds[0],
			IdentityProperties: []string{common.ObjectID.String()},
		})
	})
}

// deleteNodeDirect deletes a node without logging
func (s *service) deleteNodeDirect(ctx context.Context, objectID string) error {
	return s.graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Filter(
			query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
		).Delete()
	})
}

// createEdgeDirect creates an edge without logging
func (s *service) createEdgeDirect(ctx context.Context, entry model.GraphOperationReplayLogEntry) error {
	var labels []string
	if err := json.Unmarshal(entry.Labels, &labels); err != nil {
		return err
	}

	var properties map[string]interface{}
	if err := json.Unmarshal(entry.Properties, &properties); err != nil {
		return err
	}

	edgeKind := labels[0]
	props := properties
	if props == nil {
		props = make(map[string]interface{})
	}
	props[common.LastSeen.String()] = time.Now().UTC()

	return s.graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
		src, err := s.findNode(batch, entry.SourceObjectID.String)
		if err != nil {
			return err
		}
		tgt, err := s.findNode(batch, entry.TargetObjectID.String)
		if err != nil {
			return err
		}

		return batch.UpdateRelationshipBy(graph.RelationshipUpdate{
			Relationship:            graph.PrepareRelationship(graph.AsProperties(props), graph.StringKind(edgeKind)),
			Start:                   graph.PrepareNode(graph.AsProperties(graph.PropertyMap{common.ObjectID: strings.ToUpper(entry.SourceObjectID.String), common.LastSeen: time.Now().UTC()}), src.Kinds...),
			StartIdentityKind:       src.Kinds[0],
			StartIdentityProperties: []string{common.ObjectID.String()},
			End:                     graph.PrepareNode(graph.AsProperties(graph.PropertyMap{common.ObjectID: strings.ToUpper(entry.TargetObjectID.String), common.LastSeen: time.Now().UTC()}), tgt.Kinds...),
			EndIdentityKind:         tgt.Kinds[0],
			EndIdentityProperties:   []string{common.ObjectID.String()},
		})
	})
}

// deleteEdgeDirect deletes an edge without logging
func (s *service) deleteEdgeDirect(ctx context.Context, entry model.GraphOperationReplayLogEntry) error {
	var labels []string
	if err := json.Unmarshal(entry.Labels, &labels); err != nil {
		return err
	}
	edgeKind := labels[0]

	return s.graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		src, err := s.findNode(tx, entry.SourceObjectID.String)
		if err != nil {
			return err
		}
		tgt, err := s.findNode(tx, entry.TargetObjectID.String)
		if err != nil {
			return err
		}

		return tx.Relationships().Filter(
			query.And(
				query.Equals(query.StartID(), src.ID),
				query.Equals(query.EndID(), tgt.ID),
				query.KindIn(query.Relationship(), graph.StringKind(edgeKind)),
			),
		).Delete()
	})
}

// findNode is a helper to find nodes in batch/transaction contexts
func (s *service) findNode(queryable interface {
	Nodes() graph.NodeQuery
}, objectID string) (*graph.Node, error) {
	var node *graph.Node
	err := queryable.Nodes().Filter(
		query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
	).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for n := range cursor.Chan() {
			node = n
			break
		}
		return cursor.Error()
	})
	if err != nil {
		return nil, err
	}
	if node == nil {
		return nil, fmt.Errorf("node %s not found", objectID)
	}
	return node, nil
}
