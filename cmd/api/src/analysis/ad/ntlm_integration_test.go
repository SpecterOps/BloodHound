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

func TestPostNtlm(t *testing.T) {
	testContex := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	testContex.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.NtlmCoerceAndRelayNtlmToSmb.Setup(testContex)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNtlmToSmb")

		groupExpansions, computers, domains, authenticatedUsers, err := fetchNtlmPrereqs(db)
		require.NoError(t, err)

		for _, domain := range domains {
			innerDomain := domain

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer

					if err = ad2.PostCoerceAndRelayNtlmToSmb(tx, outC, groupExpansions, innerComputer, innerDomain.ID.String(), authenticatedUsers); err != nil {
						t.Logf("failed post processig for %s: %v", ad.CoerceAndRelayNTLMToSMB.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)
		}

		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToSMB)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {
				require.Equal(t, 1, len(results))

				require.True(t, results.Contains(harness.NtlmCoerceAndRelayNtlmToSmb.DomainAdminsUser))
				require.True(t, results.Contains(harness.NtlmCoerceAndRelayNtlmToSmb.AuthenticatedUsers))
				require.True(t, results.Contains(harness.NtlmCoerceAndRelayNtlmToSmb.ServerAdmins))

			}
			return nil
		})

		err = operation.Done()
		require.NoError(t, err)
	})
}

func fetchNtlmPrereqs(db graph.Database) (expansions impact.PathAggregator, computers []*graph.Node, domains []*graph.Node, authenticatedUsers map[string]graph.ID, err error) {
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
