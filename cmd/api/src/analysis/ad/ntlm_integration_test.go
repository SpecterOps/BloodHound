// Copyright 2024 Specter Ops, Inc.
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
// +build integration

package ad_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/cardinality"

	"github.com/specterops/bloodhound/analysis"
	ad2 "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestPostNTLMRelayADCS(t *testing.T) {
	//TODO: Add some negative tests here
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.NTLMCoerceAndRelayNTLMToADCS.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToADCS")
		_, _, domains, authenticatedUsers, err := fetchNTLMPrereqs(db)

		require.NoError(t, err)

		for _, domain := range domains {
			innerDomain := domain
			computerCache, err := fetchComputerCache(db, innerDomain)
			require.NoError(t, err)

			err = ad2.PostCoerceAndRelayNTLMToADCS(context.Background(), db, operation, authenticatedUsers, computerCache)
			require.NoError(t, err)
		}

		operation.Done()

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToADCS)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {

				require.Len(t, results, 1)
				rel := results[0]

				start, end, err := ops.FetchRelationshipNodes(tx, rel)
				require.NoError(t, err)

				require.Equal(t, start.ID, harness.NTLMCoerceAndRelayNTLMToADCS.AuthenticatedUsersGroup.ID)
				require.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToADCS.Computer.ID)
			}
			return nil
		})
	})
}

func TestPostNTLMRelaySMB(t *testing.T) {
	//TODO: Add some negative tests here
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.NTLMCoerceAndRelayNTLMToSMB.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToSMB")

		groupExpansions, computers, domains, authenticatedUsers, err := fetchNTLMPrereqs(db)
		require.NoError(t, err)

		for _, domain := range domains {
			innerDomain := domain

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer
					domainSid, _ := innerDomain.Properties.Get(ad.DomainSID.String()).String()
					authenticatedUserID := authenticatedUsers[domainSid]

					if err = ad2.PostCoerceAndRelayNTLMToSMB(tx, outC, groupExpansions, innerComputer, authenticatedUserID); err != nil {
						t.Logf("failed post processig for %s: %v", ad.CoerceAndRelayNTLMToSMB.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)
		}

		err = operation.Done()
		require.NoError(t, err)

		// Test start node
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToSMB)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {
				require.Len(t, results, 1)
				rel := results[0]
				start, end, err := ops.FetchRelationshipNodes(tx, rel)
				require.NoError(t, err)

				require.Equal(t, start.ID, harness.NTLMCoerceAndRelayNTLMToSMB.AuthenticatedUsers.ID)
				require.Equal(t, end.ID, harness.NTLMCoerceAndRelayNTLMToSMB.Computer8.ID)
			}
			return nil
		})
	})
}

func fetchComputerCache(db graph.Database, domain *graph.Node) (map[string]cardinality.Duplex[uint64], error) {
	cache := make(map[string]cardinality.Duplex[uint64])
	if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		return cache, err
	} else {
		cache[domainSid] = cardinality.NewBitmap64()
		return cache, db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			return tx.Nodes().Filter(
				query.And(
					query.Kind(query.Node(), ad.Computer),
					query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
				),
			).FetchIDs(func(cursor graph.Cursor[graph.ID]) error {
				for id := range cursor.Chan() {
					cache[domainSid].Add(id.Uint64())
				}

				return cursor.Error()
			})
		})
	}
}

func fetchNTLMPrereqs(db graph.Database) (expansions impact.PathAggregator, computers []*graph.Node, domains []*graph.Node, authenticatedUsers map[string]graph.ID, err error) {
	cache := make(map[string]graph.ID)
	if expansions, err = ad2.ExpandAllRDPLocalGroups(context.Background(), db); err != nil {
		return nil, nil, nil, cache, err
	} else if computers, err = ad2.FetchNodesByKind(context.Background(), db, ad.Computer); err != nil {
		return nil, nil, nil, cache, err
	} else if err = db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
		if cache, err = ad2.FetchAuthUsersMappedToDomains(tx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, nil, cache, err
	} else if domains, err = ad2.FetchNodesByKind(context.Background(), db, ad.Domain); err != nil {
		return nil, nil, nil, cache, err
	} else {
		return expansions, computers, domains, cache, nil
	}
}
