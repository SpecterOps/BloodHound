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
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
)

type selectorMemberFetcher func(context.Context, []model.SelectorSeed, model.AssetGroupExpansionMethod) ([]model.AssetGroupSelectorNode, []error)

// These DAWGS temporal functions can read the transaction's current time.
// A conservative token match also bypasses caching for fixed arguments or
// quoted occurrences, keeping each selector's query evaluation independent.
var volatileSelectorFunctionPattern = regexp.MustCompile(`(?i)\b(date|time|localtime|datetime|localdatetime)\b`)

func selectorMembersCacheable(seeds []model.SelectorSeed) bool {
	for _, seed := range seeds {
		if seed.Type == model.SelectorTypeCypher && volatileSelectorFunctionPattern.MatchString(seed.Value) {
			return false
		}
	}
	return true
}

func fetchSelectorMembers(ctx context.Context, parameters appcfg.AGTParameters, graphDB graph.Database, primaryDisplayKinds graphschema.PrimaryDisplayKinds, seeds []model.SelectorSeed, expansionMethod model.AssetGroupExpansionMethod) ([]model.AssetGroupSelectorNode, []error) {
	var (
		nodes, errors = FetchNodesFromSeeds(ctx, parameters, graphDB, seeds, expansionMethod, -1)
		members       = make([]model.AssetGroupSelectorNode, 0, len(nodes))
	)
	if len(errors) > 0 {
		return nil, errors
	}
	for _, node := range nodes {
		members = append(members, buildSelectorNode(primaryDisplayKinds, *node, model.AssetGroupTagSelector{}, model.AssetGroupCertificationPending, null.String{}))
	}
	return members, errors
}

type selectorMemberResult struct {
	ready   chan struct{}
	members []model.AssetGroupSelectorNode
	errors  []error
}

// Only repeated selections retain results. Cached members contain immutable
// values rather than graph nodes, and certification remains selector-specific.
type selectorMemberCache struct {
	lock          sync.Mutex
	duplicateKeys map[string]bool
	remaining     map[string]int
	results       map[string]*selectorMemberResult
	fetch         selectorMemberFetcher
}

func selectorMemberKey(seeds []model.SelectorSeed, expansionMethod model.AssetGroupExpansionMethod) string {
	var key strings.Builder
	fmt.Fprintf(&key, "%d;", expansionMethod)
	// Preserve seed order: a node can be a seed for one selection and an
	// expanded member for another. Selector IDs do not affect expansion.
	for _, seed := range seeds {
		fmt.Fprintf(&key, "%d:%d:%s;", seed.Type, len(seed.Value), seed.Value)
	}
	return key.String()
}

func newSelectorMemberCache(selectors []model.AssetGroupTagSelector, expansionByTagID map[int]model.AssetGroupExpansionMethod, fetch selectorMemberFetcher) *selectorMemberCache {
	var (
		counts = make(map[string]int)
		cache  = &selectorMemberCache{
			duplicateKeys: make(map[string]bool),
			remaining:     counts,
			results:       make(map[string]*selectorMemberResult),
			fetch:         fetch,
		}
	)
	for _, selector := range selectors {
		if !selectorMembersCacheable(selector.Seeds) {
			continue
		}
		key := selectorMemberKey(selector.Seeds, expansionByTagID[selector.AssetGroupTagId])
		counts[key]++
		cache.duplicateKeys[key] = counts[key] > 1
	}
	return cache
}

func (s *selectorMemberCache) Fetch(ctx context.Context, seeds []model.SelectorSeed, expansionMethod model.AssetGroupExpansionMethod) ([]model.AssetGroupSelectorNode, []error) {
	var key = selectorMemberKey(seeds, expansionMethod)
	if err := ctx.Err(); err != nil {
		return nil, []error{err}
	}
	if !s.duplicateKeys[key] {
		return s.fetch(ctx, seeds, expansionMethod)
	}

	s.lock.Lock()
	if s.remaining[key] <= 0 {
		s.lock.Unlock()
		return s.fetch(ctx, seeds, expansionMethod)
	}
	if result, exists := s.results[key]; exists {
		s.lock.Unlock()
		select {
		case <-result.ready:
			s.release(key, result)
			return result.members, result.errors
		case <-ctx.Done():
			s.release(key, result)
			return nil, []error{ctx.Err()}
		}
	}
	result := &selectorMemberResult{ready: make(chan struct{})}
	s.results[key] = result
	s.lock.Unlock()
	result.members, result.errors = s.fetch(ctx, seeds, expansionMethod)
	if len(result.errors) > 0 || ctx.Err() != nil {
		// Failed or canceled reads may succeed on a later attempt. Publish the
		// current error to waiting selectors without retaining it in the cache.
		s.lock.Lock()
		delete(s.results, key)
		s.lock.Unlock()
	}
	close(result.ready)
	s.release(key, result)
	return result.members, result.errors
}

func (s *selectorMemberCache) release(key string, result *selectorMemberResult) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.remaining[key]--
	if s.remaining[key] <= 0 && s.results[key] == result {
		delete(s.results, key)
	}
}
