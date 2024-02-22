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

package query

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
)

func postgresIndexType(indexType graph.IndexType) string {
	switch indexType {
	case graph.BTreeIndex:
		return pgIndexTypeBTree
	case graph.TextSearchIndex:
		return pgIndexTypeGIN
	default:
		return "NOT SUPPORTED"
	}
}

func parsePostgresIndexType(pgType string) graph.IndexType {
	switch strings.ToLower(pgType) {
	case pgIndexTypeBTree:
		return graph.BTreeIndex
	case pgIndexTypeGIN:
		return graph.TextSearchIndex
	default:
		return graph.UnsupportedIndex
	}
}

func join(values ...string) string {
	return strings.Join(values, "")
}

func formatDropPropertyIndex(indexName string) string {
	return join("drop index if exists ", indexName, ";")
}

func formatDropPropertyConstraint(constraintName string) string {
	return join("drop index if exists ", constraintName, ";")
}

func formatCreatePropertyConstraint(constraintName, tableName, fieldName string, indexType graph.IndexType) string {
	pgIndexType := postgresIndexType(indexType)

	return join("create unique index ", constraintName, " on ", tableName, " using ",
		pgIndexType, " ((", tableName, ".", pgPropertiesColumn, " ->> '", fieldName, "'));")
}

func formatCreatePropertyIndex(indexName, tableName, fieldName string, indexType graph.IndexType) string {
	var (
		pgIndexType  = postgresIndexType(indexType)
		queryPartial = join("create index ", indexName, " on ", tableName, " using ",
			pgIndexType, " ((", tableName, ".", pgPropertiesColumn, " ->> '", fieldName)
	)

	if indexType == graph.TextSearchIndex {
		// GIN text search requires the column to be typed and to contain the tri-gram operation extension
		return join(queryPartial, "'::text) gin_trgm_ops);")
	} else {
		return join(queryPartial, "'));")
	}
}

func formatCreatePartitionTable(name, parent string, graphID int32) string {
	builder := strings.Builder{}

	builder.WriteString("create table ")
	builder.WriteString(name)
	builder.WriteString(" partition of ")
	builder.WriteString(parent)
	builder.WriteString(" for values in (")
	builder.WriteString(strconv.FormatInt(int64(graphID), 10))
	builder.WriteString(")")

	return builder.String()
}

func formatConflictMatcher(propertyNames []string, defaultOnConflict string) string {
	builder := strings.Builder{}
	builder.WriteString("on conflict (")

	if len(propertyNames) > 0 {
		for idx, propertyName := range propertyNames {
			if idx > 0 {
				builder.WriteString(", ")
			}

			builder.WriteString("(properties->>'")
			builder.WriteString(propertyName)
			builder.WriteString("')")
		}
	} else {
		builder.WriteString(defaultOnConflict)
	}

	builder.WriteString(") ")
	return builder.String()
}

func FormatNodeUpsert(graphTarget model.Graph, identityProperties []string) string {
	return join(
		"insert into ", graphTarget.Partitions.Node.Name, " as n ",
		"(graph_id, kind_ids, properties) ",
		"select $1::int4, unnest($2::text[])::int2[], unnest($3::jsonb[]) ",
		formatConflictMatcher(identityProperties, "id, graph_id"),
		"do update set properties = n.properties || excluded.properties, kind_ids = uniq(sort(n.kind_ids || excluded.kind_ids)) ",
		"returning id;",
	)
}

func FormatRelationshipPartitionUpsert(graphTarget model.Graph) string {
	return join(
		"merge into ", graphTarget.Partitions.Edge.Name, " as e ",
		"using (select $1::int4 as gid, unnest($2::int4[]) as sid, unnest($3::int4[]) as eid, unnest($4::int2[]) as kid, unnest($5::jsonb[]) as p) as ei ",
		"on e.start_id = ei.sid and e.end_id = ei.eid and e.kind_id = ei.kid ",
		"when matched then update set properties = e.properties || ei.p ",
		"when not matched then insert (graph_id, start_id, end_id, kind_id, properties) values (ei.gid, ei.sid, ei.eid, ei.kid, ei.p);",
	)
}

type NodeUpdate struct {
	IDFuture *Future[graph.ID]
	Node     *graph.Node
}

// NodeUpdateBatch
//
// TODO: See note below
//
// Some assumptions were made here regarding identity kind matching since this data model does not directly require the
// kind of a node to enforce a constraint
type NodeUpdateBatch struct {
	IdentityProperties []string
	Updates            map[string]*NodeUpdate
}

func NewNodeUpdateBatch() *NodeUpdateBatch {
	return &NodeUpdateBatch{
		Updates: map[string]*NodeUpdate{},
	}
}

func (s *NodeUpdateBatch) Add(update graph.NodeUpdate) (*Future[graph.ID], error) {
	if len(s.IdentityProperties) > 0 && len(update.IdentityProperties) != len(s.IdentityProperties) {
		return nil, fmt.Errorf("node update mixes identity properties with pre-existing updates")
	}

	for _, expectedIdentityProperty := range s.IdentityProperties {
		found := false

		for _, updateIdentityProperty := range update.IdentityProperties {
			if expectedIdentityProperty == updateIdentityProperty {
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("node update mixes identity properties with pre-existing updates")
		}
	}

	if key, err := update.Key(); err != nil {
		return nil, err
	} else {
		update.Node.AddKinds(update.IdentityKind)

		if len(s.IdentityProperties) == 0 {
			s.IdentityProperties = make([]string, len(update.IdentityProperties))
			copy(s.IdentityProperties, update.IdentityProperties)
		}

		if existingUpdate, hasExisting := s.Updates[key]; hasExisting {
			existingUpdate.Node.Merge(update.Node)
			return existingUpdate.IDFuture, nil
		} else {
			newIDFuture := NewFuture(graph.ID(0))

			s.Updates[key] = &NodeUpdate{
				IDFuture: newIDFuture,
				Node:     update.Node,
			}

			return newIDFuture, nil
		}

	}
}

func ValidateNodeUpdateByBatch(updates []graph.NodeUpdate) (*NodeUpdateBatch, error) {
	updateBatch := NewNodeUpdateBatch()

	for _, update := range updates {
		if _, err := updateBatch.Add(update); err != nil {
			return nil, err
		}
	}

	return updateBatch, nil
}

type Future[T any] struct {
	Value T
}

func NewFuture[T any](value T) *Future[T] {
	return &Future[T]{
		Value: value,
	}
}

type RelationshipUpdate struct {
	StartID      *Future[graph.ID]
	EndID        *Future[graph.ID]
	Relationship *graph.Relationship
}

type RelationshipUpdateBatch struct {
	NodeUpdates        *NodeUpdateBatch
	IdentityProperties []string
	Updates            map[string]*RelationshipUpdate
}

func NewRelationshipUpdateBatch() *RelationshipUpdateBatch {
	return &RelationshipUpdateBatch{
		NodeUpdates: NewNodeUpdateBatch(),
		Updates:     map[string]*RelationshipUpdate{},
	}
}

func (s *RelationshipUpdateBatch) Add(update graph.RelationshipUpdate) error {
	if len(s.IdentityProperties) > 0 && len(update.IdentityProperties) != len(s.IdentityProperties) {
		return fmt.Errorf("relationship update mixes identity properties with pre-existing updates")
	}

	for _, expectedIdentityProperty := range s.IdentityProperties {
		found := false

		for _, updateIdentityProperty := range update.IdentityProperties {
			if expectedIdentityProperty == updateIdentityProperty {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("relationship update mixes identity properties with pre-existing updates")
		}
	}

	if startNodeID, err := s.NodeUpdates.Add(graph.NodeUpdate{
		Node:               update.Start,
		IdentityKind:       update.StartIdentityKind,
		IdentityProperties: update.StartIdentityProperties,
	}); err != nil {
		return err
	} else if endNodeID, err := s.NodeUpdates.Add(graph.NodeUpdate{
		Node:               update.End,
		IdentityKind:       update.EndIdentityKind,
		IdentityProperties: update.EndIdentityProperties,
	}); err != nil {
		return err
	} else if key, err := update.Key(); err != nil {
		return err
	} else {
		if len(s.IdentityProperties) == 0 {
			s.IdentityProperties = make([]string, len(update.IdentityProperties))
			copy(s.IdentityProperties, update.IdentityProperties)
		}

		if existingUpdate, hasExisting := s.Updates[key]; hasExisting {
			existingUpdate.Relationship.Merge(update.Relationship)
		} else {
			s.Updates[key] = &RelationshipUpdate{
				StartID:      startNodeID,
				EndID:        endNodeID,
				Relationship: update.Relationship,
			}
		}
	}

	return nil
}

func ValidateRelationshipUpdateByBatch(updates []graph.RelationshipUpdate) (*RelationshipUpdateBatch, error) {
	updateBatch := NewRelationshipUpdateBatch()

	for _, update := range updates {
		if update.Relationship.Kind.Is(ad.SQLAdmin, ad.AllowedToAct) {
			update.Relationship.Kind.Is()
		}

		if err := updateBatch.Add(update); err != nil {
			return nil, err
		}
	}

	return updateBatch, nil
}
