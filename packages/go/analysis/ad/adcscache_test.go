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
	"errors"
	"testing"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
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

		cache.enterpriseCAsWithQualifyingHosts.Add(eca.ID.Uint64())

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

		// The CA is absent from the qualifying-host bitmap because its only host
		// lives outside the CA's forest.

		result := cache.GetECAHostedChainedDomains()

		assert.NotContains(t, result, eca.ID.Uint64(), "CA with no in-forest host should be skipped")
	})

	t.Run("forest unknown: falls back to host-only gating with no domain filtering", func(t *testing.T) {
		eca, inForestDomain, foreignDomain := adcsForestHostingNodes()
		cache := newForestHostingCache(eca, inForestDomain, foreignDomain)

		cache.enterpriseCAsWithQualifyingHosts.Add(eca.ID.Uint64())

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

		// The CA has no entry in the qualifying-host bitmap.

		result := cache.GetECAHostedChainedDomains()

		assert.NotContains(t, result, eca.ID.Uint64())
	})
}

func TestGetChainedDomains_IgnoresForestHosting(t *testing.T) {
	t.Run("does not apply the hosting-computer guard", func(t *testing.T) {
		eca, inForestDomain, foreignDomain := adcsForestHostingNodes()
		cache := newForestHostingCache(eca, inForestDomain, foreignDomain)

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

func TestEnterpriseCAHostIsEligible(t *testing.T) {
	var (
		inForestDomainSID = "S-1-5-21-100-200-300"
		foreignDomainSID  = "S-1-5-21-400-500-600"
		inForestDomain    = graph.NewNode(1, graph.NewProperties(), ad.Domain)
		foreignDomain     = graph.NewNode(2, graph.NewProperties(), ad.Domain)
		domainsBySID      = map[string]*graph.Node{
			inForestDomainSID: inForestDomain,
			foreignDomainSID:  foreignDomain,
		}
		knownForest = duplexOf(inForestDomain.ID)
		testCases   = []struct {
			name             string
			kind             graph.Kind
			setEnabled       bool
			enabled          any
			setDomainSID     bool
			domainSID        string
			forestDomains    cardinality.Duplex[uint64]
			expectedEligible bool
		}{
			{
				name:             "enabled host qualifies when CA forest is unresolved",
				kind:             ad.Computer,
				setEnabled:       true,
				enabled:          true,
				forestDomains:    nil,
				expectedEligible: true,
			},
			{
				name:             "disabled host does not qualify",
				kind:             ad.Computer,
				setEnabled:       true,
				enabled:          false,
				forestDomains:    nil,
				expectedEligible: false,
			},
			{
				name:             "missing enabled property does not qualify",
				kind:             ad.Computer,
				forestDomains:    nil,
				expectedEligible: false,
			},
			{
				name:             "malformed enabled property does not qualify",
				kind:             ad.Computer,
				setEnabled:       true,
				enabled:          "true",
				forestDomains:    nil,
				expectedEligible: false,
			},
			{
				name:             "enabled in-forest host qualifies",
				kind:             ad.Computer,
				setEnabled:       true,
				enabled:          true,
				setDomainSID:     true,
				domainSID:        inForestDomainSID,
				forestDomains:    knownForest,
				expectedEligible: true,
			},
			{
				name:             "enabled cross-forest host does not qualify",
				kind:             ad.Computer,
				setEnabled:       true,
				enabled:          true,
				setDomainSID:     true,
				domainSID:        foreignDomainSID,
				forestDomains:    knownForest,
				expectedEligible: false,
			},
			{
				name:             "host missing domain SID does not qualify for known forest",
				kind:             ad.Computer,
				setEnabled:       true,
				enabled:          true,
				forestDomains:    knownForest,
				expectedEligible: false,
			},
			{
				name:             "host with unknown domain SID does not qualify for known forest",
				kind:             ad.Computer,
				setEnabled:       true,
				enabled:          true,
				setDomainSID:     true,
				domainSID:        "S-1-5-21-700-800-900",
				forestDomains:    knownForest,
				expectedEligible: false,
			},
			{
				name:             "enabled non-computer does not qualify",
				kind:             ad.User,
				setEnabled:       true,
				enabled:          true,
				setDomainSID:     true,
				domainSID:        inForestDomainSID,
				forestDomains:    knownForest,
				expectedEligible: false,
			},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			properties := graph.NewProperties()
			if testCase.setEnabled {
				properties.Set(common.Enabled.String(), testCase.enabled)
			}
			if testCase.setDomainSID {
				properties.Set(ad.DomainSID.String(), testCase.domainSID)
			}

			host := graph.NewNode(10, properties, testCase.kind)

			assert.Equal(t, testCase.expectedEligible, enterpriseCAHostIsEligible(host, testCase.forestDomains, domainsBySID))
		})
	}
}

func TestQualifyEnterpriseCAHostPathsPropagatesForestResolutionErrors(t *testing.T) {
	var (
		expectedErr  = errors.New("failed forest traversal")
		host         = graph.NewNode(1, graph.NewProperties(), ad.Computer)
		enterpriseCA = graph.NewNode(2, graph.NewProperties(), ad.EnterpriseCA)
		hostingPath  = graph.Path{
			Nodes: []*graph.Node{host, enterpriseCA},
			Edges: []*graph.Relationship{{
				ID:      3,
				Kind:    ad.HostsCAService,
				StartID: host.ID,
				EndID:   enterpriseCA.ID,
			}},
		}
	)

	qualifyingPaths, err := qualifyEnterpriseCAHostPaths(graph.NewPathSet(hostingPath), nil, func(*graph.Node) (cardinality.Duplex[uint64], error) {
		return nil, expectedErr
	})

	require.ErrorIs(t, err, expectedErr)
	assert.Nil(t, qualifyingPaths)
}

func TestQualifyEnterpriseCAHostPathsKeepsOnlyEligibleHostsForExactCA(t *testing.T) {
	const (
		inForestDomainSID = "S-1-5-21-100-200-300"
		foreignDomainSID  = "S-1-5-21-400-500-600"
	)

	var (
		inForestDomain = graph.NewNode(1, graph.NewProperties(), ad.Domain)
		foreignDomain  = graph.NewNode(2, graph.NewProperties(), ad.Domain)
		enterpriseCA   = graph.NewNode(3, graph.NewProperties(), ad.EnterpriseCA)
		validHost      = graph.NewNode(4, graph.NewProperties(), ad.Computer)
		disabledHost   = graph.NewNode(5, graph.NewProperties(), ad.Computer)
		foreignHost    = graph.NewNode(6, graph.NewProperties(), ad.Computer)
		forestDomains  = duplexOf(inForestDomain.ID)
		domainsBySID   = map[string]*graph.Node{
			inForestDomainSID: inForestDomain,
			foreignDomainSID:  foreignDomain,
		}
		forestResolutionCount int
	)

	validHost.Properties.Set(common.Enabled.String(), true)
	validHost.Properties.Set(ad.DomainSID.String(), inForestDomainSID)
	disabledHost.Properties.Set(common.Enabled.String(), false)
	disabledHost.Properties.Set(ad.DomainSID.String(), inForestDomainSID)
	foreignHost.Properties.Set(common.Enabled.String(), true)
	foreignHost.Properties.Set(ad.DomainSID.String(), foreignDomainSID)

	hostingPath := func(id graph.ID, host *graph.Node) graph.Path {
		return graph.Path{
			Nodes: []*graph.Node{host, enterpriseCA},
			Edges: []*graph.Relationship{{
				ID:      id,
				Kind:    ad.HostsCAService,
				StartID: host.ID,
				EndID:   enterpriseCA.ID,
			}},
		}
	}

	qualifyingPaths, err := qualifyEnterpriseCAHostPaths(graph.NewPathSet(
		hostingPath(10, validHost),
		hostingPath(11, disabledHost),
		hostingPath(12, foreignHost),
	), domainsBySID, func(*graph.Node) (cardinality.Duplex[uint64], error) {
		forestResolutionCount++
		return forestDomains, nil
	})

	require.NoError(t, err)
	require.Contains(t, qualifyingPaths, enterpriseCA.ID)
	require.Len(t, qualifyingPaths[enterpriseCA.ID], 1)
	assert.Equal(t, validHost.ID, qualifyingPaths[enterpriseCA.ID][0].Root().ID)
	assert.Equal(t, 1, forestResolutionCount)
}
