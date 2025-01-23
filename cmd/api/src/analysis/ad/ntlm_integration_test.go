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
	"strings"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostNTLM(t *testing.T) {
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
					domainSid, _ := innerDomain.Properties.Get(ad.Domain.String()).String()
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
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToSMB)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {
				require.Len(t, results, 1)
				resultIds := results.IDs()

				objectId := results.Get(resultIds[0]).Properties.Get("objectid")
				require.False(t, objectId.IsNil())

				objectIdStr, err := objectId.String()
				require.NoError(t, err)
				assert.True(t, strings.HasSuffix(objectIdStr, ad2.AuthenticatedUsersSuffix))
			}
			return nil
		})

		// Test end node
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToSMB)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {
				require.Len(t, results, 1)
				resultIds := results.IDs()

				objectId := results.Get(resultIds[0]).Properties.Get("objectid")
				require.False(t, objectId.IsNil())

				smbSigning, err := results.Get(resultIds[0]).Properties.Get(ad.SMBSigning.String()).Bool()
				require.NoError(t, err)

				restrictOutbountNtlm, err := results.Get(resultIds[0]).Properties.Get(ad.RestrictOutboundNTLM.String()).Bool()
				require.NoError(t, err)

				assert.False(t, smbSigning)
				assert.False(t, restrictOutbountNtlm)
			}
			return nil
		})
	})
}

func TestPostCoerceAndRelayNTLMToLDAP(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())

	// Test NTLMCoerceAndRelayNTLMToLDAP
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.NTLMCoerceAndRelayNTLMToLDAP.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToLDAP")

		_, computers, domains, authenticatedUsers, err := fetchNTLMPrereqs(db)
		require.NoError(t, err)

		for _, domain := range domains {
			innerDomain := domain

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer
					domainSid, _ := innerDomain.Properties.Get(ad.Domain.String()).String()
					authenticatedUserID := authenticatedUsers[domainSid]

					if err = ad2.PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, ad.RelayableToDCLDAP, authenticatedUserID); err != nil {
						t.Logf("failed post processig for %s: %v", ad.CoerceAndRelayNTLMToLDAP.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)
		}

		err = operation.Done()
		require.NoError(t, err)

		// Test start nodes
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToLDAP)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {
				require.Len(t, results, 2)
				require.True(t, results.Contains(harness.NTLMCoerceAndRelayNTLMToLDAP.Group1))
				require.True(t, results.Contains(harness.NTLMCoerceAndRelayNTLMToLDAP.Group5))
			}
			return nil
		})

		// Test end nodes
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToLDAP)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {
				require.Len(t, results, 2)
				require.True(t, results.Contains(harness.NTLMCoerceAndRelayNTLMToLDAP.Computer2))
				require.True(t, results.Contains(harness.NTLMCoerceAndRelayNTLMToLDAP.Computer7))
			}
			return nil
		})
	})

	// Test NTLMCoerceAndRelayNTLMToLDAPs
	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.NTLMCoerceAndRelayNTLMToLDAPs.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, db graph.Database) {
		operation := analysis.NewPostRelationshipOperation(context.Background(), db, "NTLM Post Process Test - CoerceAndRelayNTLMToLDAP")

		_, computers, domains, authenticatedUsers, err := fetchNTLMPrereqs(db)
		require.NoError(t, err)

		for _, domain := range domains {
			innerDomain := domain

			err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				for _, computer := range computers {
					innerComputer := computer
					domainSid, _ := innerDomain.Properties.Get(ad.Domain.String()).String()
					authenticatedUserID := authenticatedUsers[domainSid]

					if err = ad2.PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, ad.RelayableToDCLDAPs, authenticatedUserID); err != nil {
						t.Logf("failed post processig for %s: %v", ad.CoerceAndRelayNTLMToLDAPs.String(), err)
					}
				}
				return nil
			})
			require.NoError(t, err)
		}

		err = operation.Done()
		require.NoError(t, err)

		// Test start nodes
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToLDAPs)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {
				require.Len(t, results, 2)
				require.True(t, results.Contains(harness.NTLMCoerceAndRelayNTLMToLDAPs.Group1))
				require.True(t, results.Contains(harness.NTLMCoerceAndRelayNTLMToLDAPs.Group5))
			}
			return nil
		})

		// Test end nodes
		db.ReadTransaction(context.Background(), func(tx graph.Transaction) error {
			if results, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.Kind(query.Relationship(), ad.CoerceAndRelayNTLMToLDAPs)
			})); err != nil {
				t.Fatalf("error fetching ntlm to smb edges in integration test; %v", err)
			} else {
				require.Len(t, results, 2)
				require.True(t, results.Contains(harness.NTLMCoerceAndRelayNTLMToLDAPs.Computer2))
				require.True(t, results.Contains(harness.NTLMCoerceAndRelayNTLMToLDAPs.Computer7))
			}
			return nil
		})
	})
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
