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

//go:build integration

package ad_test

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cypher/models/cypher"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func TestNTLMADCSDeltaRetainsEdgesOnEnumerationFailure(t *testing.T) {
	var (
		testContext      = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		enumerationError = errors.New("injected NTLM computer enumeration failure")
	)

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.NTLMCoerceAndRelayNTLMToADCSRPC.Setup(testContext)
		harness.NTLMCoerceAndRelayNTLMToADCSRPC.EnterpriseCA1.Properties.Set(ad.HasVulnerableEndpoint.String(), true)
		harness.NTLMCoerceAndRelayNTLMToADCSRPC.EnterpriseCA1.Properties.Set(ad.HTTPEnrollmentEndpoints.String(), []string{"https://test.com"})
		testContext.UpdateNode(harness.NTLMCoerceAndRelayNTLMToADCSRPC.EnterpriseCA1)
		harness.NTLMCoerceAndRelayNTLMToADCSRPC.Computer.Properties.Set(ad.WebClientRunning.String(), true)
		testContext.UpdateNode(harness.NTLMCoerceAndRelayNTLMToADCSRPC.Computer)
		return nil
	}, func(_ integration.HarnessDetails, database graph.Database) {
		var (
			localGroupData, cache, prereqError = FetchADCSPrereqs(database)
			failingDatabase                    = &ntlmDCAFailureDatabase{Database: database, err: enumerationError}
		)
		require.NoError(t, prereqError)
		_, err := adAnalysis.PostNTLM(t.Context(), database, localGroupData, cache, true)
		require.NoError(t, err)
		originalEdges := fetchNTLMADCSDeltaEdges(t, database)
		require.Len(t, originalEdges, 2)
		for _, originalEdge := range originalEdges {
			firstSeen, err := originalEdge.Properties.Get(common.FirstSeen.String()).Time()
			require.NoError(t, err)
			require.False(t, firstSeen.IsZero())
			require.NotContains(t, originalEdge.Properties.Map, common.LastSeen.String())
		}

		_, err = adAnalysis.PostNTLM(t.Context(), failingDatabase, localGroupData, cache, true)
		require.ErrorIs(t, err, enumerationError)
		require.True(t, failingDatabase.failureInjected.Load(), "failure must occur after the tracker read, not during cache initialization")
		retainedEdges := fetchNTLMADCSDeltaEdges(t, database)
		require.Len(t, retainedEdges, len(originalEdges))
		for relationshipKind, originalEdge := range originalEdges {
			require.Equal(t, originalEdge.ID, retainedEdges[relationshipKind].ID)
			require.Equal(t, originalEdge.Properties.Map, retainedEdges[relationshipKind].Properties.Map)
		}

		_, err = adAnalysis.PostNTLM(t.Context(), database, localGroupData, cache, true)
		require.NoError(t, err)
		retriedEdges := fetchNTLMADCSDeltaEdges(t, database)
		require.Len(t, retriedEdges, len(originalEdges))
		for relationshipKind, originalEdge := range originalEdges {
			require.Equal(t, originalEdge.ID, retriedEdges[relationshipKind].ID)
			require.Equal(t, originalEdge.Properties.Map, retriedEdges[relationshipKind].Properties.Map)
		}
	})
}

// The fault is armed only after the existing DCA relationships have been read
// successfully. Cache-building reads run normally, and no read-count assumption
// is needed to reach the producer phase after tracker initialization.
type ntlmDCAFailureDatabase struct {
	graph.Database
	trackerRead     atomic.Bool
	failureInjected atomic.Bool
	err             error
}

func (s *ntlmDCAFailureDatabase) ReadTransaction(ctx context.Context, delegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	return s.Database.ReadTransaction(ctx, func(transaction graph.Transaction) error {
		return delegate(&ntlmDCAFailureTransaction{Transaction: transaction, database: s})
	}, options...)
}

type ntlmDCAFailureTransaction struct {
	graph.Transaction
	database *ntlmDCAFailureDatabase
}

func (s *ntlmDCAFailureTransaction) Nodes() graph.NodeQuery {
	return &ntlmDCAFailureNodeQuery{NodeQuery: s.Transaction.Nodes(), database: s.database}
}

func (s *ntlmDCAFailureTransaction) Relationships() graph.RelationshipQuery {
	return &ntlmDCATrackerRelationshipQuery{RelationshipQuery: s.Transaction.Relationships(), database: s.database}
}

type ntlmDCATrackerRelationshipQuery struct {
	graph.RelationshipQuery
	database     *ntlmDCAFailureDatabase
	trackedKinds bool
}

func (s *ntlmDCATrackerRelationshipQuery) Filter(criteria graph.Criteria) graph.RelationshipQuery {
	s.RelationshipQuery = s.RelationshipQuery.Filter(criteria)
	s.trackedKinds = s.trackedKinds || ntlmDCACriteriaHasKind(criteria, query.Relationship(), ad.CoerceAndRelayNTLMToADCS, ad.CoerceAndRelayNTLMToADCSRPC)
	return s
}

func (s *ntlmDCATrackerRelationshipQuery) Filterf(provider graph.CriteriaProvider) graph.RelationshipQuery {
	return s.Filter(provider())
}

func (s *ntlmDCATrackerRelationshipQuery) Fetch(delegate func(graph.Cursor[*graph.Relationship]) error) error {
	if err := s.RelationshipQuery.Fetch(delegate); err != nil {
		return err
	} else if s.trackedKinds {
		s.database.trackerRead.Store(true)
	}
	return nil
}

type ntlmDCAFailureNodeQuery struct {
	graph.NodeQuery
	database    *ntlmDCAFailureDatabase
	faultTarget bool
}

func (s *ntlmDCAFailureNodeQuery) Filter(criteria graph.Criteria) graph.NodeQuery {
	s.NodeQuery = s.NodeQuery.Filter(criteria)
	s.faultTarget = s.faultTarget || ntlmDCACriteriaHasKind(criteria, query.Node(), ad.Computer)
	return s
}

func (s *ntlmDCAFailureNodeQuery) Filterf(provider graph.CriteriaProvider) graph.NodeQuery {
	return s.Filter(provider())
}

func (s *ntlmDCAFailureNodeQuery) Fetch(delegate func(graph.Cursor[*graph.Node]) error, finalCriteria ...graph.Criteria) error {
	if s.faultTarget && s.database.trackerRead.Load() {
		s.database.failureInjected.Store(true)
		return s.database.err
	}
	return s.NodeQuery.Fetch(delegate, finalCriteria...)
}

func ntlmDCACriteriaHasKind(criteria graph.Criteria, reference graph.Criteria, kinds ...graph.Kind) bool {
	switch typedCriteria := criteria.(type) {
	case *cypher.KindMatcher:
		return reflect.DeepEqual(typedCriteria.Reference, reference) && typedCriteria.Kinds.ContainsOneOf(kinds...)
	case *cypher.Conjunction:
		for _, expression := range typedCriteria.Expressions {
			if ntlmDCACriteriaHasKind(expression, reference, kinds...) {
				return true
			}
		}
	}
	return false
}
