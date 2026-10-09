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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestADCSESC16DeltaRetainsEdgesOnReaderFailure(t *testing.T) {
	var (
		testContext = integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
		readerError = errors.New("injected ESC16 enroller filtering failure")
		harness     esc16IntegrationHarness
	)

	testContext.DatabaseTestWithSetup(func(_ *integration.HarnessDetails) error {
		harness = setupESC16IntegrationHarness(testContext, ad.Group)
		// DNS requirements cause an enroller-filtering query. The group remains
		// eligible, so the failure does not represent successful hardening.
		harness.certTemplate.Properties.Set(ad.SubjectAltRequireDNS.String(), true)
		testContext.UpdateNode(harness.certTemplate)
		return nil
	}, func(_ integration.HarnessDetails, database graph.Database) {
		var (
			localGroupData, groupError = adAnalysis.FetchLocalGroupData(t.Context(), database)
			failingDatabase            = &esc16DCAFailureDatabase{Database: database, err: readerError, enrollerID: harness.attacker.ID}
		)
		require.NoError(t, groupError)
		_, _, err := adAnalysis.PostADCS(t.Context(), database, localGroupData)
		require.NoError(t, err)
		originalEdges := fetchESC16IntegrationEdges(t, database)
		require.Len(t, originalEdges, 1)
		originalEdge := originalEdges[0]
		firstSeen, err := originalEdge.Properties.Get(common.FirstSeen.String()).Time()
		require.NoError(t, err)
		require.False(t, firstSeen.IsZero())
		require.NotContains(t, originalEdge.Properties.Map, common.LastSeen.String())

		_, _, err = adAnalysis.PostADCS(t.Context(), failingDatabase, localGroupData)
		assert.ErrorIs(t, err, readerError)
		require.True(t, failingDatabase.failureInjected.Load(), "failure must occur after the tracker read, not during cache initialization")
		retainedEdges := fetchESC16IntegrationEdges(t, database)
		require.Len(t, retainedEdges, 1)
		require.Equal(t, originalEdge.ID, retainedEdges[0].ID)
		require.Equal(t, originalEdge.Properties.Map, retainedEdges[0].Properties.Map)

		_, _, err = adAnalysis.PostADCS(t.Context(), database, localGroupData)
		require.NoError(t, err)
		retriedEdges := fetchESC16IntegrationEdges(t, database)
		require.Len(t, retriedEdges, 1)
		require.Equal(t, originalEdge.ID, retriedEdges[0].ID)
		require.Equal(t, originalEdge.Properties.Map, retriedEdges[0].Properties.Map)
	})
}

// The fault is armed only after the existing DCA relationships have been read
// successfully. Cache-building reads run normally, and no read-count assumption
// is needed to reach the producer phase after tracker initialization.
type esc16DCAFailureDatabase struct {
	graph.Database
	trackerRead     atomic.Bool
	failureInjected atomic.Bool
	err             error
	enrollerID      graph.ID
}

func (s *esc16DCAFailureDatabase) ReadTransaction(ctx context.Context, delegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	return s.Database.ReadTransaction(ctx, func(transaction graph.Transaction) error {
		return delegate(&esc16DCAFailureTransaction{Transaction: transaction, database: s})
	}, options...)
}

type esc16DCAFailureTransaction struct {
	graph.Transaction
	database *esc16DCAFailureDatabase
}

func (s *esc16DCAFailureTransaction) Nodes() graph.NodeQuery {
	return &esc16DCAFailureNodeQuery{NodeQuery: s.Transaction.Nodes(), database: s.database}
}

func (s *esc16DCAFailureTransaction) Relationships() graph.RelationshipQuery {
	return &esc16DCATrackerRelationshipQuery{RelationshipQuery: s.Transaction.Relationships(), database: s.database}
}

type esc16DCATrackerRelationshipQuery struct {
	graph.RelationshipQuery
	database     *esc16DCAFailureDatabase
	trackedKinds bool
}

func (s *esc16DCATrackerRelationshipQuery) Filter(criteria graph.Criteria) graph.RelationshipQuery {
	s.RelationshipQuery = s.RelationshipQuery.Filter(criteria)
	s.trackedKinds = s.trackedKinds || esc16DCACriteriaHasKind(criteria, query.Relationship(), ad.ADCSESC16)
	return s
}

func (s *esc16DCATrackerRelationshipQuery) Filterf(provider graph.CriteriaProvider) graph.RelationshipQuery {
	return s.Filter(provider())
}

func (s *esc16DCATrackerRelationshipQuery) Fetch(delegate func(graph.Cursor[*graph.Relationship]) error) error {
	if err := s.RelationshipQuery.Fetch(delegate); err != nil {
		return err
	} else if s.trackedKinds {
		s.database.trackerRead.Store(true)
	}
	return nil
}

type esc16DCAFailureNodeQuery struct {
	graph.NodeQuery
	database    *esc16DCAFailureDatabase
	faultTarget bool
}

func (s *esc16DCAFailureNodeQuery) Filter(criteria graph.Criteria) graph.NodeQuery {
	s.NodeQuery = s.NodeQuery.Filter(criteria)
	s.faultTarget = s.faultTarget || reflect.DeepEqual(criteria, query.And(
		query.KindIn(query.Node(), ad.User),
		query.InIDs(query.NodeID(), s.database.enrollerID),
	))
	return s
}

func (s *esc16DCAFailureNodeQuery) Filterf(provider graph.CriteriaProvider) graph.NodeQuery {
	return s.Filter(provider())
}

func (s *esc16DCAFailureNodeQuery) Fetch(delegate func(graph.Cursor[*graph.Node]) error, finalCriteria ...graph.Criteria) error {
	if s.faultTarget && s.database.trackerRead.Load() {
		s.database.failureInjected.Store(true)
		return s.database.err
	}
	return s.NodeQuery.Fetch(delegate, finalCriteria...)
}

func esc16DCACriteriaHasKind(criteria graph.Criteria, reference graph.Criteria, kinds ...graph.Kind) bool {
	switch typedCriteria := criteria.(type) {
	case *cypher.KindMatcher:
		return reflect.DeepEqual(typedCriteria.Reference, reference) && typedCriteria.Kinds.ContainsOneOf(kinds...)
	case *cypher.Conjunction:
		for _, expression := range typedCriteria.Expressions {
			if esc16DCACriteriaHasKind(expression, reference, kinds...) {
				return true
			}
		}
	}
	return false
}
