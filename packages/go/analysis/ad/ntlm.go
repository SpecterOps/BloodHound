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
	"log/slog"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

// PostNTLM is the initial function used to execute our NTLM analysis
func PostNTLM(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "PostNTLM")

	// TODO: after adding all of our new NTLM edges, benchmark performance between submitting multiple readers per computer or single reader per computer
	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {

		// Fetch all nodes where the node is a Group and is an Authenticated User
		if authenticatedUsersCache, err := FetchAuthUsersMappedToDomains(tx); err != nil {
			return err
		} else {
			// Fetch all nodes where the type is Computer
			return tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for computer := range cursor.Chan() {
					innerComputer := computer

					if domain, err := innerComputer.Properties.Get(ad.Domain.String()).String(); err != nil {
						continue
					} else if authenticatedUserID, ok := authenticatedUsersCache[domain]; !ok {
						continue
					} else if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						return PostCoerceAndRelayNTLMToSMB(tx, outC, groupExpansions, innerComputer, authenticatedUserID)
					}); err != nil {
						slog.WarnContext(ctx, "Post processing failed for %s: %v", ad.CoerceAndRelayNTLMToSMB, err)
						// Additional analysis may occur if one of our analysis errors
						continue
					}
				}

				return cursor.Error()
			})
		}
	})
	if err != nil {
		operation.Done()
		return nil, err
	}

	return &operation.Stats, operation.Done()
}

// PostCoerceAndRelayNTLMToSMB creates edges that allow a computer with unrolled admin access to one or more computers where SMB signing is disabled.
// Comprised solely of adminTo and memberOf edges
func PostCoerceAndRelayNTLMToSMB(tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, expandedGroups impact.PathAggregator, computer *graph.Node, authenticatedUserID graph.ID) error {
	if smbSigningEnabled, err := computer.Properties.Get(ad.SMBSigning.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		return nil
	} else if err != nil {
		return err
	} else if restrictOutboundNtlm, err := computer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		return nil
	} else if err != nil {
		return err
	} else if !smbSigningEnabled && !restrictOutboundNtlm {

		// Fetch the admins with edges to the provided computer
		if firstDegreeAdmins, err := fetchFirstDegreeNodes(tx, computer, ad.AdminTo); err != nil {
			return err
		} else if firstDegreeAdmins.ContainingNodeKinds(ad.Computer).Len() > 0 {
			outC <- analysis.CreatePostRelationshipJob{
				FromID: authenticatedUserID,
				ToID:   computer.ID,
				Kind:   ad.CoerceAndRelayNTLMToSMB,
			}
		} else {
			allAdminGroups := cardinality.NewBitmap64()
			for group := range firstDegreeAdmins.ContainingNodeKinds(ad.Group) {
				allAdminGroups.And(expandedGroups.Cardinality(group.Uint64()))
			}

			// Fetch nodes where the node id is in our allAdminGroups bitmap and are of type Computer
			if computerIds, err := ops.FetchNodeIDs(tx.Nodes().Filter(
				query.And(
					query.InIDs(query.Node(), graph.DuplexToGraphIDs(allAdminGroups)...),
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

	return nil
}

// FetchAuthUsersMappedToDomains Fetch all nodes where the node is a Group and is an Authenticated User
func FetchAuthUsersMappedToDomains(tx graph.Transaction) (map[string]graph.ID, error) {
	authenticatedUsers := make(map[string]graph.ID)

	err := tx.Nodes().Filter(
		query.And(
			query.Kind(query.Node(), ad.Group),
			query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), AuthenticatedUsersSuffix)),
	).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for authenticatedUser := range cursor.Chan() {
			if domain, err := authenticatedUser.Properties.Get(ad.Domain.String()).String(); err != nil {
				continue
			} else {
				authenticatedUsers[domain] = authenticatedUser.ID
			}
		}

		return cursor.Error()
	},
	)

	return authenticatedUsers, err
}
