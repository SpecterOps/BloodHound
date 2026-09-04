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

package ad

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func duplexOf(ids ...graph.ID) cardinality.Duplex[uint64] {
	bitmap := cardinality.NewBitmap64()
	for _, id := range ids {
		bitmap.Add(id.Uint64())
	}
	return bitmap
}

// newForestHostingCache builds an ADCSCache where the CA has a valid cert chain
// to both domains, so only the hosting-computer gate can distinguish whether the
// CA should be processed.
func newForestHostingCache(eca, inForestDomain, foreignDomain *graph.Node) *ADCSCache {
	cache := NewADCSCache()

	cache.enterpriseCertAuthorities = []*graph.Node{eca}
	cache.domains = []*graph.Node{inForestDomain, foreignDomain}

	for _, domain := range cache.domains {
		cache.rootCAForChainValid[domain.ID] = duplexOf(eca.ID)
		cache.authStoreForChainValid[domain.ID] = duplexOf(eca.ID)
	}

	return cache
}

func adcsForestHostingNodes() (eca, inForestDomain, foreignDomain *graph.Node) {
	eca = graph.NewNode(10, graph.NewProperties(), ad.EnterpriseCA)
	inForestDomain = graph.NewNode(1, graph.NewProperties(), ad.Domain)
	foreignDomain = graph.NewNode(2, graph.NewProperties(), ad.Domain)
	return eca, inForestDomain, foreignDomain
}

func TestGetECAHostedChainedDomains_ForestHosting(t *testing.T) {
	t.Run("forest known and host in forest: keeps all chained domains", func(t *testing.T) {
		eca, inForestDomain, foreignDomain := adcsForestHostingNodes()
		cache := newForestHostingCache(eca, inForestDomain, foreignDomain)

		cache.hasHostingComputer[eca.ID] = true
		cache.hasInForestHostingComputer[eca.ID] = true

		result := cache.GetECAHostedChainedDomains()

		require.Contains(t, result, eca.ID.Uint64())
		chains := result[eca.ID.Uint64()]
		assert.True(t, chains.Domains.Contains(inForestDomain.ID.Uint64()), "in-forest domain should survive")
		assert.True(t, chains.Domains.Contains(foreignDomain.ID.Uint64()), "foreign-forest chained domain should survive")
		assert.Equal(t, uint64(2), chains.Domains.Cardinality())
	})

	t.Run("forest known but only a cross-forest hosting computer: CA is dropped entirely", func(t *testing.T) {
		eca, inForestDomain, foreignDomain := adcsForestHostingNodes()
		cache := newForestHostingCache(eca, inForestDomain, foreignDomain)

		// A hosting computer exists, but it lives outside the CA's forest.
		cache.hasHostingComputer[eca.ID] = true
		cache.hasInForestHostingComputer[eca.ID] = false

		result := cache.GetECAHostedChainedDomains()

		assert.NotContains(t, result, eca.ID.Uint64(), "CA with no in-forest host should be skipped")
	})

	t.Run("forest unknown: falls back to host-only gating with no domain filtering", func(t *testing.T) {
		eca, inForestDomain, foreignDomain := adcsForestHostingNodes()
		cache := newForestHostingCache(eca, inForestDomain, foreignDomain)

		// No hasInForestHostingComputer entry => forest unknown.
		cache.hasHostingComputer[eca.ID] = true

		result := cache.GetECAHostedChainedDomains()

		require.Contains(t, result, eca.ID.Uint64())
		chains := result[eca.ID.Uint64()]
		assert.True(t, chains.Domains.Contains(inForestDomain.ID.Uint64()))
		assert.True(t, chains.Domains.Contains(foreignDomain.ID.Uint64()), "fallback should preserve prior behavior")
		assert.Equal(t, uint64(2), chains.Domains.Cardinality())
	})

	t.Run("forest unknown and no hosting computer: CA is dropped", func(t *testing.T) {
		eca, inForestDomain, foreignDomain := adcsForestHostingNodes()
		cache := newForestHostingCache(eca, inForestDomain, foreignDomain)

		// hasHostingComputer absent/false and forest unknown.

		result := cache.GetECAHostedChainedDomains()

		assert.NotContains(t, result, eca.ID.Uint64())
	})
}

func TestGetChainedDomains_IgnoresForestHosting(t *testing.T) {
	t.Run("does not apply the hosting-computer guard", func(t *testing.T) {
		eca, inForestDomain, foreignDomain := adcsForestHostingNodes()
		cache := newForestHostingCache(eca, inForestDomain, foreignDomain)
		cache.hasInForestHostingComputer[eca.ID] = false

		result := cache.GetChainedDomains()

		require.Contains(t, result, eca.ID.Uint64())
		chains := result[eca.ID.Uint64()]
		assert.True(t, chains.Domains.Contains(inForestDomain.ID.Uint64()))
		assert.True(t, chains.Domains.Contains(foreignDomain.ID.Uint64()))
	})

	t.Run("forest unknown: keeps every chained domain", func(t *testing.T) {
		eca, inForestDomain, foreignDomain := adcsForestHostingNodes()
		cache := newForestHostingCache(eca, inForestDomain, foreignDomain)

		result := cache.GetChainedDomains()

		require.Contains(t, result, eca.ID.Uint64())
		chains := result[eca.ID.Uint64()]
		assert.Equal(t, uint64(2), chains.Domains.Cardinality())
	})
}
