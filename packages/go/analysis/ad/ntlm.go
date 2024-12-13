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

package ad

import (
	"context"
	"errors"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

// PostNTLM is the initial function used to execute our NTLM analysis
func PostNTLM(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "PostNTLM")

	// TODO: after adding all of our new NTLM edges, benchmark performance between submitting multiple readers per computer or single reader per computer
	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {

		authenticatedUsersCache := make(map[string]graph.ID)

		// Fetch all nodes where the node is a Group and is an Authenticated User
		if err := tx.Nodes().Filter(
			query.And(
				query.Kind(query.Node(), ad.Group),
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), AuthenticatedUsersSuffix)),
		).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for authenticatedUser := range cursor.Chan() {
				if domain, err := authenticatedUser.Properties.Get(ad.Domain.String()).String(); err != nil {
					continue
				} else {
					authenticatedUsersCache[domain] = authenticatedUser.ID
				}
			}

			return cursor.Error()
		},
		); err != nil {
			return err
		} else {
			// Fetch all nodes where the type is Computer
			return tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for computer := range cursor.Chan() {
					innerComputer := computer

					if domain, err := innerComputer.Properties.Get(ad.Domain.String()).String(); err != nil {
						continue
					} else {
						if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
							return PostCoerceAndRelayNtlmToSmb(tx, outC, groupExpansions, innerComputer, domain, authenticatedUsersCache)
						}); err != nil {
							log.Warnf("Post processing failed for %s: %v", ad.CoerceAndRelayNTLMToSMB, err)
							// Additional analysis may occur if one of our analysis errors
							continue
						}
					}
				}

				return cursor.Error()
			})
		}
	})
	if err != nil {
		return nil, err
	}

	return &operation.Stats, operation.Done()
}

// PostCoerceAndRelayNtlmToSmb creates edges that allow a computer with unrolled admin access to one or more computers where SMB signing is disabled.
// Comprised solely oof adminTo and memberOf edges
func PostCoerceAndRelayNtlmToSmb(tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, expandedGroups impact.PathAggregator, computer *graph.Node, domain string, authenticaedUserNodes map[string]graph.ID) error {
	if authenticatedUserID, ok := authenticaedUserNodes[domain]; !ok {
		return nil
	} else if smbSigningEnabled, err := computer.Properties.Get(ad.SmbSigning.String()).Bool(); err != nil {
		if errors.Is(err, graph.ErrPropertyNotFound) {
			return nil
		} else {
			return err
		}
	} else if !smbSigningEnabled {

		// Fetch the admins with edges to the provided computer
		if firstDegreeAdmins, err := fetchFirstDegreeNodes(tx, computer, ad.AdminTo); err != nil {
			return err
		} else {
			if firstDegreeAdmins.ContainingNodeKinds(ad.Computer).Len() > 0 {
				outC <- analysis.CreatePostRelationshipJob{
					FromID: authenticatedUserID,
					ToID:   computer.ID,
					Kind:   ad.CoerceAndRelayNTLMToSMB,
				}
			} else {
				allAdmins := cardinality.NewBitmap64()
				for group := range firstDegreeAdmins.ContainingNodeKinds(ad.Group) {
					allAdmins.And(expandedGroups.Cardinality(group.Uint64()))
				}

				// Fetch nodes where the node id is in our allAdmins bitmap and are of type Computer
				if computerIds, err := ops.FetchNodeIDs(tx.Nodes().Filter(
					query.And(
						query.InIDs(query.Node(), graph.DuplexToGraphIDs(allAdmins)...),
						query.Kind(query.Node(), ad.Computer),
					),
				)); err != nil {
					return err
				} else if len(computerIds) > 0 {
					outC <- analysis.CreatePostRelationshipJob{
						FromID: authenticatedUserID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToSMB,
					}
				}
			}
		}
	}

	return nil
}
