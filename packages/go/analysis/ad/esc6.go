// Copyright 2023 Specter Ops, Inc.
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
	"fmt"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
)

func PostCanAbuseUPNCertMapping(_ context.Context, _ graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node) error {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, eca := range enterpriseCertAuthorities {
			if ecaDomainSID, err := eca.Properties.Get(ad.DomainSID.String()).String(); err != nil {
				return err
			} else if ecaDomain, err := analysis.FetchNodeByObjectID(tx, ecaDomainSID); err != nil {
				return err
			} else if trustedByNodes, err := fetchNodesWithTrustedByParentChildRelationship(tx, ecaDomain); err != nil {
				return err
			} else {
				for _, trustedByDomain := range trustedByNodes {
					if dcForNodes, err := fetchNodesWithDCForEdge(tx, trustedByDomain); err != nil {
						return err
					} else {
						for _, dcForNode := range dcForNodes {
							if cmmrProperty, err := dcForNode.Properties.Get(ad.CertificateMappingMethodsRaw.String()).Int(); err != nil {
								return err
							} else if cmmrProperty&0x04 == 0x04 {
								if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
									FromID: eca.ID,
									ToID:   dcForNode.ID,
									Kind:   ad.CanAbuseUPNCertMapping,
								}) {
									return fmt.Errorf("context timed out while creating CanAbuseUPNCert edge")
								}
							}
						}
					}
				}
			}
		}
		return nil
	})
	return nil
}

func fetchNodesWithTrustedByParentChildRelationship(tx graph.Transaction, root *graph.Node) (graph.NodeSet, error) {
	if nodeSet, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      root,
		Direction: graph.DirectionOutbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.KindIn(query.Relationship(), ad.TrustedBy),
				query.Equals(query.RelationshipProperty(ad.TrustType.String()), "ParentChild"),
			)
		},
	}); err != nil {
		return graph.NodeSet{}, err
	} else {
		nodeSet.Add(root)
		return nodeSet, nil
	}
}

func fetchNodesWithDCForEdge(tx graph.Transaction, rootNode *graph.Node) (graph.NodeSet, error) {
	return ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
		Root:      rootNode,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.And(
				query.KindIn(query.Start(), ad.Computer),
				query.KindIn(query.Relationship(), ad.DCFor),
			)
		},
	})
}
