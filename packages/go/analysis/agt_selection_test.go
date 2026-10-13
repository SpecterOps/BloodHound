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

package analysis

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSelectorMemberCacheSharesConcurrentExpansions(t *testing.T) {
	var (
		seeds     = []model.SelectorSeed{{SelectorId: 1, Type: model.SelectorTypeObjectId, Value: "seed"}}
		selectors = func() []model.AssetGroupTagSelector {
			var selectors = make([]model.AssetGroupTagSelector, 20)
			for index := range selectors {
				selectors[index] = model.AssetGroupTagSelector{ID: index + 1, Seeds: seeds}
			}
			return selectors
		}()
		calls   atomic.Int32
		workers sync.WaitGroup
		results = make(chan []model.AssetGroupSelectorNode, 20)
		cache   = newSelectorMemberCache(selectors, nil, func(context.Context, []model.SelectorSeed, model.AssetGroupExpansionMethod) ([]model.AssetGroupSelectorNode, []error) {
			calls.Add(1)
			return []model.AssetGroupSelectorNode{{NodeId: 42, Source: model.AssetGroupSelectorNodeSourceSeed}}, nil
		})
	)
	for index := range 20 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			var clonedSeeds = []model.SelectorSeed{{SelectorId: index + 1, Type: model.SelectorTypeObjectId, Value: "seed"}}
			members, errors := cache.Fetch(context.Background(), clonedSeeds, model.AssetGroupExpansionMethodNone)
			if len(errors) == 0 {
				results <- members
			}
		}()
	}
	workers.Wait()
	close(results)
	require.Equal(t, int32(1), calls.Load())
	require.Len(t, results, 20)
	for result := range results {
		require.Equal(t, graph.ID(42), result[0].NodeId)
	}

	require.Empty(t, cache.results, "completed selections must release the cached members")
	_, errors := cache.Fetch(context.Background(), seeds, model.AssetGroupExpansionMethodAll)
	require.Empty(t, errors)
	require.Equal(t, int32(2), calls.Load(), "expansion methods must remain distinct")
}

func TestSelectorMemberCacheCancellationAndUniqueSelections(t *testing.T) {
	var (
		seeds   = []model.SelectorSeed{{Type: model.SelectorTypeCypher, Value: "query"}}
		started = make(chan struct{})
		release = make(chan struct{})
		done    = make(chan struct{})
		cache   = newSelectorMemberCache([]model.AssetGroupTagSelector{{Seeds: seeds}, {Seeds: seeds}}, nil, func(context.Context, []model.SelectorSeed, model.AssetGroupExpansionMethod) ([]model.AssetGroupSelectorNode, []error) {
			select {
			case <-started:
			default:
				close(started)
			}
			<-release
			return nil, []error{errors.New("fetch failed")}
		})
	)
	go func() {
		defer close(done)
		_, _ = cache.Fetch(context.Background(), seeds, model.AssetGroupExpansionMethodNone)
	}()
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, fetchErrors := cache.Fetch(ctx, seeds, model.AssetGroupExpansionMethodNone)
	require.Len(t, fetchErrors, 1)
	require.ErrorIs(t, fetchErrors[0], context.Canceled)
	close(release)
	<-done
	_, fetchErrors = cache.Fetch(context.Background(), seeds, model.AssetGroupExpansionMethodNone)
	require.EqualError(t, fetchErrors[0], "fetch failed")
	require.Empty(t, cache.results, "failed reads must be retried rather than retained")

	unique := newSelectorMemberCache([]model.AssetGroupTagSelector{{Seeds: seeds}}, nil, func(context.Context, []model.SelectorSeed, model.AssetGroupExpansionMethod) ([]model.AssetGroupSelectorNode, []error) {
		return nil, nil
	})
	_, _ = unique.Fetch(context.Background(), seeds, model.AssetGroupExpansionMethodNone)
	require.Empty(t, unique.results, "single-use expansions must not be retained")
}

func TestSelectorMemberCacheBypassesTemporalQueries(t *testing.T) {
	for _, query := range []string{
		"MATCH (n) WHERE n.lastlogon < datetime() RETURN n",
		"MATCH (n) WHERE n.lastlogon < DATETIME /* comment */ () RETURN n",
		"MATCH (n) WHERE n.date < date() RETURN n",
		"MATCH (n) WHERE n.lastlogon < localdatetime () RETURN n",
		"MATCH (n) WHERE n.time < localtime() RETURN n",
		"MATCH (n) WHERE n.time < time() RETURN n",
		"MATCH (n {name: 'datetime'}) RETURN n",
	} {
		t.Run(query, func(t *testing.T) {
			var (
				seeds = []model.SelectorSeed{{Type: model.SelectorTypeCypher, Value: query}}
				calls int
				cache = newSelectorMemberCache([]model.AssetGroupTagSelector{{Seeds: seeds}, {Seeds: seeds}}, nil, func(context.Context, []model.SelectorSeed, model.AssetGroupExpansionMethod) ([]model.AssetGroupSelectorNode, []error) {
					calls++
					return nil, nil
				})
			)
			_, _ = cache.Fetch(context.Background(), seeds, model.AssetGroupExpansionMethodNone)
			_, _ = cache.Fetch(context.Background(), seeds, model.AssetGroupExpansionMethodNone)
			require.Equal(t, 2, calls, "time-dependent selectors must execute independently")
			require.Empty(t, cache.results)
		})
	}
	require.True(t, selectorMembersCacheable([]model.SelectorSeed{{Type: model.SelectorTypeCypher, Value: "MATCH (n:Group) WHERE n.datetime_property = 1 RETURN n"}}))
	require.True(t, selectorMembersCacheable([]model.SelectorSeed{{Type: model.SelectorTypeObjectId, Value: "datetime"}}))
}

func TestCachedSelectorMembersKeepCertificationAndFilteringPerSelector(t *testing.T) {
	var (
		ctx      = context.Background()
		database = mocks.NewMockDatabase(gomock.NewController(t))
		members  = []model.AssetGroupSelectorNode{
			{NodeId: 1, Source: model.AssetGroupSelectorNodeSourceSeed, NodeEnvironmentId: "first"},
			{NodeId: 2, Source: model.AssetGroupSelectorNodeSourceSeed, NodeEnvironmentId: "second"},
		}
		fetch = func(context.Context, []model.SelectorSeed, model.AssetGroupExpansionMethod) ([]model.AssetGroupSelectorNode, []error) {
			return members, nil
		}
		manual = model.AssetGroupSelectorNode{SelectorId: 1, NodeId: 1, Certified: model.AssetGroupCertificationManual, CertifiedBy: null.StringFrom("reviewer")}
	)
	database.EXPECT().GetSelectorNodesBySelectorIds(ctx, 1).Return([]model.AssetGroupSelectorNode{manual}, nil)
	database.EXPECT().UpdateSelectorNodes(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, nodes []model.AssetGroupSelectorNode) error {
		require.Len(t, nodes, 1)
		require.Equal(t, manual.Certified, nodes[0].Certified)
		require.Equal(t, manual.CertifiedBy, nodes[0].CertifiedBy)
		return nil
	})
	require.Empty(t, selectNodesWithFetch(ctx, database, model.AssetGroupTagSelector{ID: 1, AutoCertify: model.SelectorAutoCertifyMethodAllMembers}, func(node model.AssetGroupSelectorNode) bool { return node.NodeEnvironmentId == "first" }, fetch, model.AssetGroupExpansionMethodAll))
	database.EXPECT().GetSelectorNodesBySelectorIds(ctx, 2).Return([]model.AssetGroupSelectorNode{}, nil)
	database.EXPECT().InsertSelectorNodes(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, nodes []model.AssetGroupSelectorNode) error {
		require.Len(t, nodes, 1)
		require.Equal(t, graph.ID(2), nodes[0].NodeId)
		require.Equal(t, model.AssetGroupCertificationAuto, nodes[0].Certified)
		return nil
	})
	require.Empty(t, selectNodesWithFetch(ctx, database, model.AssetGroupTagSelector{ID: 2, AutoCertify: model.SelectorAutoCertifyMethodAllMembers}, func(node model.AssetGroupSelectorNode) bool { return node.NodeEnvironmentId == "second" }, fetch, model.AssetGroupExpansionMethodAll))
	require.Zero(t, members[0].SelectorId)
	require.Zero(t, members[1].SelectorId)
	require.NotEqual(t, model.AssetGroupCertificationManual, members[0].Certified)
}
