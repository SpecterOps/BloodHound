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

package datapipe

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nodeDeleteCall records the arguments of a single DeleteNodesByKinds invocation as kind strings so tests can assert
// on the translated set-based deletes without depending on graph.Kind identity.
type nodeDeleteCall struct {
	includeAny []string
	excludeAny []string
}

// fakeGraphDeleter is a test double for graph.Database that additionally implements both nodesByKindDeleter and
// relationshipsByKindDeleter. Because graph.AsDriver type-asserts the database against the requested interface, a value
// of this type is dispatched down the set-based deletion path in DeleteCollectedGraphData. The embedded graph.Database
// is left nil: only the set-based delete methods are expected to be called, and any accidental use of another method
// will panic and surface the mistake.
type fakeGraphDeleter struct {
	graph.Database

	nodeCalls         []nodeDeleteCall
	relationshipCalls [][]string
	nodeErr           error
	relationshipErr   error
}

func (s *fakeGraphDeleter) DeleteNodesByKinds(_ context.Context, includeAny graph.Kinds, excludeAny graph.Kinds) error {
	s.nodeCalls = append(s.nodeCalls, nodeDeleteCall{includeAny: includeAny.Strings(), excludeAny: excludeAny.Strings()})
	return s.nodeErr
}

func (s *fakeGraphDeleter) DeleteRelationshipsByKinds(_ context.Context, kinds graph.Kinds) error {
	s.relationshipCalls = append(s.relationshipCalls, kinds.Strings())
	return s.relationshipErr
}

func TestDeleteCollectedGraphData_SetBasedNodeDeletion(t *testing.T) {
	var (
		ctx         = context.Background()
		sourceKinds = graph.Kinds{graph.StringKind("Base"), graph.StringKind("AZBase"), graph.StringKind("GithubBase")}
	)

	type testCase struct {
		name          string
		deleteRequest model.AnalysisRequest
		expectedCalls []nodeDeleteCall
	}

	testCases := []testCase{
		{
			name:          "DeleteAllGraph deletes every node except MigrationData",
			deleteRequest: model.AnalysisRequest{DeleteAllGraph: true},
			expectedCalls: []nodeDeleteCall{
				{includeAny: nil, excludeAny: []string{common.MigrationData.String()}},
			},
		},
		{
			name:          "DeleteSourceKinds deletes only the requested kinds and preserves MigrationData",
			deleteRequest: model.AnalysisRequest{DeleteSourceKinds: []string{"AZBase", "GithubBase"}},
			expectedCalls: []nodeDeleteCall{
				{includeAny: []string{"AZBase", "GithubBase"}, excludeAny: []string{common.MigrationData.String()}},
			},
		},
		{
			name:          "DeleteSourcelessGraph deletes nodes lacking any source kind and preserves MigrationData",
			deleteRequest: model.AnalysisRequest{DeleteSourcelessGraph: true},
			expectedCalls: []nodeDeleteCall{
				{includeAny: nil, excludeAny: append([]string{common.MigrationData.String()}, sourceKinds.Strings()...)},
			},
		},
		{
			name:          "DeleteSourcelessGraph and DeleteSourceKinds issue a union of two deletes",
			deleteRequest: model.AnalysisRequest{DeleteSourcelessGraph: true, DeleteSourceKinds: []string{"AZBase", "GithubBase"}},
			expectedCalls: []nodeDeleteCall{
				{includeAny: []string{"AZBase", "GithubBase"}, excludeAny: []string{common.MigrationData.String()}},
				{includeAny: nil, excludeAny: append([]string{common.MigrationData.String()}, sourceKinds.Strings()...)},
			},
		},
		{
			name:          "DeleteAllGraph takes precedence over DeleteSourceKinds and DeleteSourcelessGraph",
			deleteRequest: model.AnalysisRequest{DeleteAllGraph: true, DeleteSourcelessGraph: true, DeleteSourceKinds: []string{"AZBase"}},
			expectedCalls: []nodeDeleteCall{
				{includeAny: nil, excludeAny: []string{common.MigrationData.String()}},
			},
		},
		{
			name:          "empty request performs no node deletes",
			deleteRequest: model.AnalysisRequest{},
			expectedCalls: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			deleter := &fakeGraphDeleter{}

			err := DeleteCollectedGraphData(ctx, deleter, testCase.deleteRequest, sourceKinds)
			require.NoError(t, err)

			require.Len(t, deleter.nodeCalls, len(testCase.expectedCalls))
			for i, expected := range testCase.expectedCalls {
				assert.ElementsMatch(t, expected.includeAny, deleter.nodeCalls[i].includeAny, "includeAny for call %d", i)
				assert.ElementsMatch(t, expected.excludeAny, deleter.nodeCalls[i].excludeAny, "excludeAny for call %d", i)
			}
			assert.Empty(t, deleter.relationshipCalls)
		})
	}
}

func TestDeleteCollectedGraphData_SetBasedRelationshipDeletion(t *testing.T) {
	var (
		ctx         = context.Background()
		sourceKinds = graph.Kinds{graph.StringKind("Base")}
	)

	t.Run("DeleteRelationships issues a single set-based edge delete and no node deletes", func(t *testing.T) {
		deleter := &fakeGraphDeleter{}

		deleteRequest := model.AnalysisRequest{DeleteRelationships: []string{"MemberOf", "HasSession"}}

		err := DeleteCollectedGraphData(ctx, deleter, deleteRequest, sourceKinds)
		require.NoError(t, err)

		assert.Empty(t, deleter.nodeCalls)
		require.Len(t, deleter.relationshipCalls, 1)
		assert.ElementsMatch(t, []string{"MemberOf", "HasSession"}, deleter.relationshipCalls[0])
	})

	t.Run("node and relationship deletes are both dispatched when requested together", func(t *testing.T) {
		deleter := &fakeGraphDeleter{}

		deleteRequest := model.AnalysisRequest{
			DeleteSourceKinds:   []string{"AZBase"},
			DeleteRelationships: []string{"MemberOf"},
		}

		err := DeleteCollectedGraphData(ctx, deleter, deleteRequest, sourceKinds)
		require.NoError(t, err)

		require.Len(t, deleter.nodeCalls, 1)
		assert.ElementsMatch(t, []string{"AZBase"}, deleter.nodeCalls[0].includeAny)
		assert.ElementsMatch(t, []string{common.MigrationData.String()}, deleter.nodeCalls[0].excludeAny)
		require.Len(t, deleter.relationshipCalls, 1)
		assert.ElementsMatch(t, []string{"MemberOf"}, deleter.relationshipCalls[0])
	})
}

func TestDeleteCollectedGraphData_SetBasedErrors(t *testing.T) {
	var (
		ctx         = context.Background()
		sourceKinds = graph.Kinds{graph.StringKind("Base")}
		sentinelErr = errors.New("boom")
	)

	t.Run("node delete error is wrapped", func(t *testing.T) {
		deleter := &fakeGraphDeleter{nodeErr: sentinelErr}

		err := DeleteCollectedGraphData(ctx, deleter, model.AnalysisRequest{DeleteAllGraph: true}, sourceKinds)
		require.ErrorIs(t, err, sentinelErr)
		assert.Contains(t, err.Error(), "error deleting graph nodes")
	})

	t.Run("relationship delete error is wrapped", func(t *testing.T) {
		deleter := &fakeGraphDeleter{relationshipErr: sentinelErr}

		err := DeleteCollectedGraphData(ctx, deleter, model.AnalysisRequest{DeleteRelationships: []string{"MemberOf"}}, sourceKinds)
		require.ErrorIs(t, err, sentinelErr)
		assert.Contains(t, err.Error(), "error deleting graph edges")
	})
}

func TestDrainAndDelete(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes every id received from the channel", func(t *testing.T) {
		var (
			ids     = []graph.ID{graph.ID(1), graph.ID(2), graph.ID(3)}
			deleted []graph.ID
		)

		inC := make(chan graph.ID, len(ids))
		for _, id := range ids {
			inC <- id
		}
		close(inC)

		err := drainAndDelete(ctx, inC, func(id graph.ID) error {
			deleted = append(deleted, id)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, ids, deleted)
	})

	t.Run("returns the first delete error and stops draining", func(t *testing.T) {
		var (
			sentinelErr = errors.New("delete failed")
			attempts    int
		)

		inC := make(chan graph.ID, 3)
		inC <- graph.ID(1)
		inC <- graph.ID(2)
		inC <- graph.ID(3)
		close(inC)

		err := drainAndDelete(ctx, inC, func(id graph.ID) error {
			attempts++
			return sentinelErr
		})

		require.ErrorIs(t, err, sentinelErr)
		assert.Equal(t, 1, attempts)
	})

	t.Run("returns nil for an already drained channel", func(t *testing.T) {
		inC := make(chan graph.ID)
		close(inC)

		err := drainAndDelete(ctx, inC, func(id graph.ID) error {
			t.Fatal("delete should not be called for an empty channel")
			return nil
		})

		require.NoError(t, err)
	})
}
