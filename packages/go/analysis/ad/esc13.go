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
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/traversal"
	"sync"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

func PostADCSESC13(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca, domain *graph.Node, cache ADCSCache) error {
	if publishedCertTemplates, ok := cache.PublishedTemplateCache[eca.ID]; !ok {
		return nil
	} else {
		for _, template := range publishedCertTemplates {
			if isValid, err := isCertTemplateValidForESC13(template); err != nil {
				log.Errorf("Error checking esc13 cert template: %v", err)
			} else if !isValid {
				continue
			} else if groupNodes, err := getCertTemplateGroupLinks(template, tx); err != nil {
				log.Errorf("Error getting cert template group links: %v", err)
			} else if len(groupNodes) == 0 {
				continue
			} else {
				controlBitmap := CalculateCrossProductNodeSets(groupExpansions, cache.CertTemplateEnrollers[template.ID], cache.EnterpriseCAEnrollers[eca.ID])
				if filtered, err := filterUserDNSResults(tx, controlBitmap, template); err != nil {
					log.Warnf("Error filtering users from victims for esc13: %v", err)
					continue
				} else {
					for _, group := range groupNodes.Slice() {
						if groupIsContainedOrTrusted(tx, group, domain) {
							filtered.Each(func(value uint32) bool {
								channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
									FromID: graph.ID(value),
									ToID:   group.ID,
									Kind:   ad.ADCSESC13,
								})
								return true
							})
						}
					}
				}
			}
		}

		return nil
	}
}

func groupIsContainedOrTrusted(tx graph.Transaction, group, domain *graph.Node) bool {
	var matchFound bool
	if err := ops.Traversal(tx, ops.TraversalPlan{
		Root:      group,
		Direction: graph.DirectionInbound,
		BranchQuery: func() graph.Criteria {
			return query.KindIn(query.Relationship(), ad.Contains, ad.TrustedBy)
		},
		PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
			return segment.Node.Kinds.ContainsOneOf(ad.Domain)
		},
	}, func(ctx *ops.TraversalContext, segment *graph.PathSegment) error {
		//Check to make sure that this segment contains our target domain id
		segment.WalkReverse(func(nextSegment *graph.PathSegment) bool {
			if nextSegment.Node.ID == domain.ID {
				matchFound = true
				return false
			}

			if !nextSegment.Node.Kinds.ContainsOneOf(ad.Domain) {
				return false
			}

			return true
		})

		return nil
	}); err != nil {
		return false
	} else {
		return matchFound
	}
}

func isCertTemplateValidForESC13(ct *graph.Node) (bool, error) {
	if reqManagerApproval, err := ct.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		return false, err
	} else if reqManagerApproval {
		return false, nil
	} else if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
		return false, nil
	} else if schemaVersion, err := ct.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		return false, err
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return false, err
	} else if schemaVersion > 1 && authorizedSignatures > 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func getCertTemplateGroupLinks(ct *graph.Node, tx graph.Transaction) (graph.NodeSet, error) {
	if policyNodes, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.InIDs(query.Start(), ct.ID),
			query.KindIn(query.Relationship(), ad.ExtendedByPolicy),
			query.KindIn(query.End(), ad.IssuancePolicy),
		)
	})); err != nil {
		return graph.NodeSet{}, err
	} else if len(policyNodes) == 0 {
		return graph.NodeSet{}, nil
	} else {
		policyNodeIDs := policyNodes.IDs()
		if groupNodes, err := ops.FetchEndNodes(tx.Relationships().Filterf(func() graph.Criteria {
			return query.And(
				query.InIDs(query.Start(), policyNodeIDs...),
				query.KindIn(query.Relationship(), ad.OIDGroupLink),
				query.KindIn(query.End(), ad.Group),
			)
		})); err != nil {
			return graph.NodeSet{}, err
		} else {
			return groupNodes, nil
		}
	}
}
func GetADCSESC13EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH (n {objectid:'<principal sid>'})-[:ADCSESC13]->(g:Group {objectid:'<group sid>'})
		MATCH p1 = (n)-[:MemberOf*0..]->()-[:Enroll|GenericAll|AllExtendedRights]->(ct:CertTemplate)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		WHERE ct.authenticationenabled = true
		  AND (ct.authorizedsignatures = 0 OR ct.schemaversion = 1)
		  AND ct.requiresmanagerapproval = False
		  AND (
		    n:Group
		    OR n:Computer
		    OR (
		      n:User
		      AND ct.subjectaltrequiredns = false
		      AND ct.subjectaltrequiredomaindns = false
		    )
		  )
		MATCH p2 = (n)-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)
		MATCH p3 = (ct)-[:ExtendedByPolicy]->(:IssuancePolicy)-[:OIDGroupLink]->(g)
		MATCH p4 = (d)-[:Contains|TrustedBy*..]->(g)
		RETURN p1,p2,p3,p4
	*/

	var (
		startNode *graph.Node
		endNode   *graph.Node

		traversalInst          = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths                  = graph.PathSet{}
		path1CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path1EnterpriseCAs     = cardinality.NewBitmap32()
		path1DomainNodes       = cardinality.NewBitmap32()
		path1CertTemplates     = cardinality.NewBitmap32()
		path2CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path3CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path4CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path2EnterpriseCAs     = cardinality.NewBitmap32()
		lock                   = &sync.Mutex{}
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else if endNode, err = ops.FetchNode(tx, edge.EndID); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	}

	//Manifest P1
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: adcsESC13Path1Pattern().Do(func(terminal *graph.PathSegment) error {
			domainNode := terminal.Node
			var enterpriseCANode *graph.Node
			terminal.WalkReverse(func(nextSegment *graph.PathSegment) bool {
				if nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) {
					enterpriseCANode = nextSegment.Node
				}
				return true
			})

			certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			path1CandidateSegments[enterpriseCANode.ID] = append(path1CandidateSegments[enterpriseCANode.ID], terminal)
			path1EnterpriseCAs.Add(enterpriseCANode.ID.Uint32())
			path1DomainNodes.Add(domainNode.ID.Uint32())
			path1CertTemplates.Add(certTemplate.ID.Uint32())
			lock.Unlock()
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	//Manifest P2 and key it to the enterprise CA nodes to cross product
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: adcsESC13Path2Pattern(cardinality.DuplexToGraphIDs(path1EnterpriseCAs), cardinality.DuplexToGraphIDs(path1DomainNodes)).Do(func(terminal *graph.PathSegment) error {
			var enterpriseCANode *graph.Node
			terminal.WalkReverse(func(nextSegment *graph.PathSegment) bool {
				if nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) {
					enterpriseCANode = nextSegment.Node
				}
				return true
			})
			lock.Lock()
			path2CandidateSegments[enterpriseCANode.ID] = append(path2CandidateSegments[enterpriseCANode.ID], terminal)
			path2EnterpriseCAs.Add(enterpriseCANode.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	//Manifest P3 keyed to cert template nodes
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: endNode,
		Driver: adcsESC13Path3Pattern(cardinality.DuplexToGraphIDs(path1CertTemplates)).Do(func(terminal *graph.PathSegment) error {
			certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			path3CandidateSegments[certTemplate.ID] = append(path3CandidateSegments[certTemplate.ID], terminal)
			lock.Unlock()
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	//Manifest P4, keyed to the domain nodes
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: endNode,
		Driver: adcsESC13Path4Pattern().Do(func(terminal *graph.PathSegment) error {
			if terminal.Node.Kinds.ContainsOneOf(ad.Domain) {
				lock.Lock()
				path4CandidateSegments[terminal.Node.ID] = append(path4CandidateSegments[terminal.Node.ID], terminal)
				lock.Unlock()
			}

			return nil

		}),
	}); err != nil {
		return nil, err
	}

	//And the 2 bitmaps together to ensure that only enterprise CAs thats satisfy both paths are valid
	path1EnterpriseCAs.And(path2EnterpriseCAs)

	for path1NodeIndex, path1Segments := range path1CandidateSegments {
		path2segments := path2CandidateSegments[path1NodeIndex]
		for _, p1 := range path1Segments {
			//If we don't have a domain contain relationship, then this is not a valid path and we can kick out early
			p4segments, ok := path4CandidateSegments[p1.Node.ID]
			if !ok {
				continue
			}
			for _, p2 := range path2segments {
				//Check if our terminal nodes match (should be the same domain node)
				if p1.Node.ID != p2.Node.ID {
					continue
				}

				//Find the cert template in path 1 and use that to find the correct p3 segment
				certTemplate := p1.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
				})

				if p3segments, ok := path3CandidateSegments[certTemplate.ID]; !ok {
					continue
				} else {
					//Merge all our paths together
					paths.AddPath(p1.Path())
					paths.AddPath(p2.Path())
					for _, p3 := range p3segments {
						paths.AddPath(p3.Path())
					}
					for _, p4 := range p4segments {
						paths.AddPath(p4.Path())
					}
				}
			}
		}
	}

	return paths, nil
}

func adcsESC13Path1Pattern() traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(
			0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			),
		).
		OutboundWithDepth(
			1, 1,
			query.And(
				query.KindIn(query.Relationship(), ad.Enroll, ad.GenericAll, ad.AllExtendedRights),
				query.KindIn(query.End(), ad.CertTemplate),
				query.And(
					query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
					query.Or(
						query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
						query.And(
							query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
							query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
						),
					),
					query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
				),
			),
		).OutboundWithDepth(
		1, 1,
		query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.KindIn(query.End(), ad.Domain),
		))
}

func adcsESC13Path2Pattern(caNodes, domains []graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.Kind(query.Relationship(), ad.Enroll),
			query.InIDs(query.End(), caNodes...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.Kind(query.End(), ad.NTAuthStore),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.InIDs(query.End(), domains...),
		))
}

func adcsESC13Path3Pattern(certTemplates []graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		Inbound(
			query.And(
				query.Kind(query.Relationship(), ad.OIDGroupLink),
				query.Kind(query.Start(), ad.IssuancePolicy),
			),
		).
		Inbound(
			query.And(
				query.Kind(query.Relationship(), ad.ExtendedByPolicy),
				query.InIDs(query.Start(), certTemplates...),
			),
		)
}

func adcsESC13Path4Pattern() traversal.PatternContinuation {
	return traversal.NewPattern().
		Inbound(query.Kind(query.Relationship(), ad.Contains)).
		InboundWithDepth(0, 0, query.Kind(query.Relationship(), ad.TrustedBy))
}
