// Copyright 2026 Specter Ops, Inc.
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

package post_test

import (
	"context"
	"errors"
	"testing"

	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	azureAnalysis "github.com/specterops/bloodhound/packages/go/analysis/azure"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/cypher/models/cypher"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util/size"
	"github.com/stretchr/testify/require"
)

func TestFilteredRelationshipSink_CallersCompleteOnlySuccessfulGenerations(t *testing.T) {
	var processors = []struct {
		name         string
		trackedKind  graph.Kind
		producerKind graph.Kind
		process      func(context.Context, graph.Database) (*post.AtomicPostProcessingStats, error)
	}{
		{
			name: "owns enumeration", trackedKind: ad.Owns, producerKind: ad.OwnsRaw,
			process: func(ctx context.Context, database graph.Database) (*post.AtomicPostProcessingStats, error) {
				return adAnalysis.PostOwnsAndWriteOwner(ctx, database, &adAnalysis.LocalGroupData{})
			},
		},
		{
			name: "write owner enumeration", trackedKind: ad.WriteOwner, producerKind: ad.WriteOwnerRaw,
			process: func(ctx context.Context, database graph.Database) (*post.AtomicPostProcessingStats, error) {
				return adAnalysis.PostOwnsAndWriteOwner(ctx, database, &adAnalysis.LocalGroupData{})
			},
		},
		{
			name: "app role assignments", trackedKind: azure.AZMGAddSecret, producerKind: azure.Contains,
			process: azureAnalysis.AppRoleAssignments,
		},
		{
			name: "add owner", trackedKind: azure.AddOwner, producerKind: azure.Contains,
			process: azureAnalysis.CreateAZAddOwnerEdge,
		},
		{
			name: "user role assignments", trackedKind: azure.ResetPassword, producerKind: azure.Contains,
			process: azureAnalysis.UserRoleAssignments,
		},
	}
	for _, processor := range processors {
		t.Run(processor.name, func(t *testing.T) {
			for _, outcome := range []string{"producer failure", "successful empty generation", "writer failure", "pruning failure"} {
				t.Run(outcome, func(t *testing.T) {
					var (
						failure  = errors.New(outcome)
						database = &sinkCallerTestDatabase{
							sinkTestDatabase: &sinkTestDatabase{relationships: map[graph.ID]bool{77: true}},
							trackedKind:      processor.trackedKind, producerKind: processor.producerKind,
						}
					)
					switch outcome {
					case "producer failure":
						database.producerError = failure
					case "writer failure":
						database.flushError = failure
					case "pruning failure":
						database.deleteError = failure
					}

					_, err := processor.process(t.Context(), database)
					if outcome == "successful empty generation" {
						require.NoError(t, err)
						require.Empty(t, database.relationships, "a complete empty result must remove obsolete edges")
					} else {
						require.ErrorIs(t, err, failure)
						require.True(t, database.relationships[77], "incomplete processing must retain the existing edge")
					}
					if outcome == "producer failure" {
						require.True(t, database.failureInjected, "the failure must occur in production after loading the tracker")
					}
				})
			}
		})
	}
}

func TestFilteredRelationshipSink_UserRoleAssignmentsRetainEdgesOnInvalidGroupProperty(t *testing.T) {
	var database = &sinkCallerTestDatabase{
		sinkTestDatabase: &sinkTestDatabase{relationships: map[graph.ID]bool{77: true}},
		trackedKind:      azure.ResetPassword,
		invalidGroup:     true,
	}
	_, err := azureAnalysis.UserRoleAssignments(t.Context(), database)
	require.Error(t, err)
	require.True(t, database.relationships[77], "an invalid AddMembers input must not prune existing role edges")
}

// Query kinds select the producer fault so prerequisite and tracker reads still succeed.
type sinkCallerTestDatabase struct {
	*sinkTestDatabase
	trackedKind     graph.Kind
	producerKind    graph.Kind
	producerError   error
	failureInjected bool
	invalidGroup    bool
}

func (s *sinkCallerTestDatabase) ReadTransaction(_ context.Context, delegate graph.TransactionDelegate, _ ...graph.TransactionOption) error {
	return delegate(&sinkCallerTransaction{database: s})
}

type sinkCallerTransaction struct {
	graph.Transaction
	database *sinkCallerTestDatabase
}

func (s *sinkCallerTransaction) Nodes() graph.NodeQuery {
	return &sinkCallerNodeQuery{}
}

func (s *sinkCallerTransaction) Relationships() graph.RelationshipQuery {
	return &sinkCallerRelationshipQuery{database: s.database}
}

func (s *sinkCallerTransaction) GraphQueryMemoryLimit() size.Size {
	return 0
}

type sinkCallerRelationshipQuery struct {
	graph.RelationshipQuery
	database *sinkCallerTestDatabase
	criteria graph.Criteria
}

func (s *sinkCallerRelationshipQuery) Filter(criteria graph.Criteria) graph.RelationshipQuery {
	s.criteria = criteria
	return s
}

func (s *sinkCallerRelationshipQuery) Filterf(provider graph.CriteriaProvider) graph.RelationshipQuery {
	return s.Filter(provider())
}

func (s *sinkCallerRelationshipQuery) producerFailure() error {
	if s.database.producerError != nil && sinkCallerCriteriaHasKind(s.criteria, s.database.producerKind) {
		s.database.failureInjected = true
		return s.database.producerError
	}
	return nil
}

func (s *sinkCallerRelationshipQuery) Fetch(delegate func(graph.Cursor[*graph.Relationship]) error) error {
	if err := s.producerFailure(); err != nil {
		return err
	}
	if sinkCallerCriteriaHasKind(s.criteria, s.database.trackedKind) {
		return delegate(newSinkCallerCursor(graph.NewRelationship(77, 1, 2, graph.NewProperties(), s.database.trackedKind)))
	}
	return delegate(newSinkCallerCursor[*graph.Relationship]())
}

func (s *sinkCallerRelationshipQuery) FetchIDs(delegate func(graph.Cursor[graph.ID]) error) error {
	return delegate(newSinkCallerCursor[graph.ID]())
}

func (s *sinkCallerRelationshipQuery) FetchDirection(_ graph.Direction, delegate func(graph.Cursor[graph.DirectionalResult]) error) error {
	if err := s.producerFailure(); err != nil {
		return err
	}
	if s.database.invalidGroup && sinkCallerCriteriaHasKind(s.criteria, azure.Contains) && sinkCallerCriteriaHasKind(s.criteria, azure.Group) {
		var properties = graph.NewProperties()
		properties.Set(azure.IsAssignableToRole.String(), 42)
		return delegate(newSinkCallerCursor(graph.DirectionalResult{
			Node: graph.NewNode(5, properties, azure.Group),
		}))
	}
	return delegate(newSinkCallerCursor[graph.DirectionalResult]())
}

type sinkCallerNodeQuery struct {
	graph.NodeQuery
	criteria graph.Criteria
}

func (s *sinkCallerNodeQuery) Filter(criteria graph.Criteria) graph.NodeQuery {
	s.criteria = criteria
	return s
}

func (s *sinkCallerNodeQuery) Filterf(provider graph.CriteriaProvider) graph.NodeQuery {
	return s.Filter(provider())
}

func (s *sinkCallerNodeQuery) Fetch(delegate func(graph.Cursor[*graph.Node]) error, _ ...graph.Criteria) error {
	if sinkCallerCriteriaHasKind(s.criteria, azure.Tenant) {
		return delegate(newSinkCallerCursor(graph.NewNode(3, graph.NewProperties(), azure.Tenant)))
	}
	return delegate(newSinkCallerCursor[*graph.Node]())
}

func (s *sinkCallerNodeQuery) FetchIDs(delegate func(graph.Cursor[graph.ID]) error) error {
	return delegate(newSinkCallerCursor[graph.ID]())
}

func sinkCallerCriteriaHasKind(criteria graph.Criteria, kind graph.Kind) bool {
	switch typedCriteria := criteria.(type) {
	case *cypher.KindMatcher:
		return typedCriteria.Kinds.ContainsOneOf(kind)
	case *cypher.Conjunction:
		for _, expression := range typedCriteria.Expressions {
			if sinkCallerCriteriaHasKind(expression, kind) {
				return true
			}
		}
	}
	return false
}

type sinkCallerCursor[T any] struct {
	values chan T
}

func newSinkCallerCursor[T any](values ...T) *sinkCallerCursor[T] {
	var valueChannel = make(chan T, len(values))
	for _, value := range values {
		valueChannel <- value
	}
	close(valueChannel)
	return &sinkCallerCursor[T]{values: valueChannel}
}

func (s *sinkCallerCursor[T]) Chan() chan T { return s.values }
func (s *sinkCallerCursor[T]) Error() error { return nil }
func (s *sinkCallerCursor[T]) Close()       {}
