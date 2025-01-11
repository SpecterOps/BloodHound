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
	"fmt"
	"log/slog"
	"sync"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/traversal"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
)

func PostADCSESC9a(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca, domain *graph.Node, cache ADCSCache) error {
	results := cardinality.NewBitmap64()

	if ok := cache.HasWeakCertBindingInForest(domain.ID.Uint64()); !ok {
		return nil
	} else if publishedCertTemplates := cache.GetPublishedTemplateCache(eca.ID); len(publishedCertTemplates) == 0 {
		return nil
	} else if ecaEnrollers := cache.GetEnterpriseCAEnrollers(eca.ID); len(ecaEnrollers) == 0 {
		return nil
	} else {
		for _, template := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForESC9(template, false); err != nil {
				slog.WarnContext(ctx, fmt.Sprintf("Error validating cert template %d: %v", template.ID, err))
				continue
			} else if !valid {
				continue
			} else if certTemplateEnrollers := cache.GetCertTemplateEnrollers(template.ID); len(certTemplateEnrollers) == 0 {
				slog.DebugContext(ctx, fmt.Sprintf("Failed to retrieve enrollers for cert template %d from cache", template.ID))
				continue
			} else {
				victimBitmap := getVictimBitmap(groupExpansions, certTemplateEnrollers, ecaEnrollers, cache.GetCertTemplateHasSpecialEnrollers(template.ID), cache.GetEnterpriseCAHasSpecialEnrollers(eca.ID))

				if filteredVictims, err := filterUserDNSResults(tx, victimBitmap, template); err != nil {
					slog.WarnContext(ctx, fmt.Sprintf("Error filtering users from victims for esc9a: %v", err))
					continue
				} else if attackers, err := FetchAttackersForEscalations9and10(tx, filteredVictims, false); err != nil {
					slog.WarnContext(ctx, fmt.Sprintf("Error getting start nodes for esc9a attacker nodes: %v", err))
					continue
				} else {
					results.Or(graph.NodeIDsToDuplex(attackers))
				}
			}
		}

		results.Each(func(value uint64) bool {
			return channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC9a,
			})
		})

		return nil
	}
}

func PostADCSESC9b(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca, domain *graph.Node, cache ADCSCache) error {
	results := cardinality.NewBitmap64()

	if ok := cache.HasWeakCertBindingInForest(domain.ID.Uint64()); !ok {
		return nil
	} else if publishedCertTemplates := cache.GetPublishedTemplateCache(eca.ID); len(publishedCertTemplates) == 0 {
		return nil
	} else if ecaEnrollers := cache.GetEnterpriseCAEnrollers(eca.ID); len(ecaEnrollers) == 0 {
		return nil
	} else {
		for _, template := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForESC9(template, true); err != nil {
				slog.WarnContext(ctx, fmt.Sprintf("Error validating cert template %d: %v", template.ID, err))
				continue
			} else if !valid {
				continue
			} else if certTemplateEnrollers := cache.GetCertTemplateEnrollers(template.ID); len(certTemplateEnrollers) == 0 {
				slog.DebugContext(ctx, fmt.Sprintf("Failed to retrieve enrollers for cert template %d from cache", template.ID))
				continue
			} else {
				victimBitmap := getVictimBitmap(groupExpansions, certTemplateEnrollers, ecaEnrollers, cache.GetCertTemplateHasSpecialEnrollers(template.ID), cache.GetEnterpriseCAHasSpecialEnrollers(eca.ID))

				if attackers, err := FetchAttackersForEscalations9and10(tx, victimBitmap, true); err != nil {
					slog.WarnContext(ctx, fmt.Sprintf("Error getting start nodes for esc9a attacker nodes: %v", err))
					continue
				} else {
					results.Or(graph.NodeIDsToDuplex(attackers))
				}
			}
		}

		results.Each(func(value uint64) bool {
			return channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC9b,
			})
		})
		return nil
	}
}

func isCertTemplateValidForESC9(ct *graph.Node, scenarioB bool) (bool, error) {
	if reqManagerApproval, err := ct.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		return false, err
	} else if reqManagerApproval {
		return false, nil
	} else if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
		return false, nil
	} else if noSecurityExtension, err := ct.Properties.Get(ad.NoSecurityExtension.String()).Bool(); err != nil {
		return false, err
	} else if !noSecurityExtension {
		return false, nil
	} else if enrolleeSuppliesSubject, err := ct.Properties.Get(ad.EnrolleeSuppliesSubject.String()).Bool(); err != nil {
		return false, err
	} else if enrolleeSuppliesSubject {
		return false, nil
	} else if schemaVersion, err := ct.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		return false, err
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return false, err
	} else if schemaVersion > 1 && authorizedSignatures > 0 {
		return false, nil
	} else if !scenarioB {
		if subjectAltRequireUPN, err := ct.Properties.Get(ad.SubjectAltRequireUPN.String()).Bool(); err != nil {
			return false, err
		} else if subjectAltRequireSPN, err := ct.Properties.Get(ad.SubjectAltRequireSPN.String()).Bool(); err != nil {
			return false, err
		} else if subjectAltRequireSPN || subjectAltRequireUPN {
			return true, nil
		} else {
			return false, err
		}
	} else {
		if subjectAltRequireDNS, err := ct.Properties.Get(ad.SubjectAltRequireDNS.String()).Bool(); err != nil {
			return false, err
		} else if subjectAltRequireDNS {
			return true, nil
		} else {
			return false, nil
		}
	}
}

func GetADCSESC9aEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH (n {objectid:'S-1-5-21-3933516454-2894985453-2515407000-500'})-[:ADCSESC9a]->(d:Domain {objectid:'S-1-5-21-3933516454-2894985453-2515407000'})
		MATCH p1 = (n)-[:GenericAll|GenericWrite|Owns|WriteOwner|WriteDacl]->(m)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		WHERE ct.requiresmanagerapproval = false
		AND ct.authenticationenabled = true
		AND ct.nosecurityextension = true
		AND ct.enrolleesuppliessubject = false
		AND (ct.subjectaltrequireupn = true OR ct.subjectaltrequirespn = true)
		AND (
		(ct.schemaversion > 1 AND ct.authorizedsignatures = 0)
		OR ct.schemaversion = 1
		)
		AND (
		m:Computer
		OR (m:User AND ct.subjectaltrequiredns = false AND ct.subjectaltrequiredomaindns = false)
		)
		MATCH p2 = (m)-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)
		MATCH p3 = (d)<-[r:TrustedBy*0..]-()<-[:DCFor]-(dc:Computer)
		WITH *, relationships(p3) AS r
		WHERE ALL(rel IN r WHERE type(rel) = "DCFor" OR rel.trusttype = "ParentChild")
		AND (
			dc.strongcertificatebindingenforcementraw = 0
			OR dc.strongcertificatebindingenforcementraw = 1
		)
		RETURN p1,p2,p3
	*/

	var (
		startNode *graph.Node
		endNode   *graph.Node

		traversalInst          = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths                  = graph.PathSet{}
		path1CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		victimCANodes          = map[graph.ID][]graph.ID{}
		path2CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path3CandidateSegments = []*graph.PathSegment{}
		p2canodes              = make([]graph.ID, 0)
		nodeMap                = map[graph.ID]*graph.Node{}
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

	//Fully manifest p1
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: adcsESC9aPath1Pattern(edge.EndID).Do(func(terminal *graph.PathSegment) error {
			victimNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Depth() == 1
			})

			if victimNode.Kinds.ContainsOneOf(ad.User) {
				certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
				})

				if !certTemplateValidForUserVictim(certTemplate) {
					return nil
				}
			}

			// First ECA in the path
			var caNode *graph.Node
			terminal.Path().Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
				if end.Kinds.ContainsOneOf(ad.EnterpriseCA) {
					caNode = end
					return false
				}
				return true
			})

			lock.Lock()
			path1CandidateSegments[victimNode.ID] = append(path1CandidateSegments[victimNode.ID], terminal)
			nodeMap[victimNode.ID] = victimNode
			victimCANodes[victimNode.ID] = append(victimCANodes[victimNode.ID], caNode.ID)
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	for victim, p1CANodes := range victimCANodes {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: nodeMap[victim],
			Driver: adcsESC9APath2Pattern(p1CANodes, edge.EndID).Do(func(terminal *graph.PathSegment) error {
				caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
				})

				lock.Lock()
				path2CandidateSegments[caNode.ID] = append(path2CandidateSegments[caNode.ID], terminal)
				p2canodes = append(p2canodes, caNode.ID)
				lock.Unlock()

				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	if len(p2canodes) > 0 {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: endNode,
			Driver: adcsESC9APath3Pattern().Do(func(terminal *graph.PathSegment) error {
				terminalNode := terminal.Node
				if terminalNode.Kinds.ContainsOneOf(ad.Computer) {
					strongBinding, err := terminalNode.Properties.Get(ad.StrongCertificateBindingEnforcementRaw.String()).Float64()
					if err == nil && (strongBinding == 1 || strongBinding == 0) {
						lock.Lock()
						path3CandidateSegments = append(path3CandidateSegments, terminal)
						lock.Unlock()
					}
				}
				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	for _, p1paths := range path1CandidateSegments {
		for _, p1path := range p1paths {
			// First ECA in the path
			var caNode *graph.Node
			p1path.Path().Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
				if end.Kinds.ContainsOneOf(ad.EnterpriseCA) {
					caNode = end
					return false
				}
				return true
			})

			if p2segments, ok := path2CandidateSegments[caNode.ID]; !ok {
				continue
			} else {
				paths.AddPath(p1path.Path())
				for _, p2 := range p2segments {
					paths.AddPath(p2.Path())
				}
			}
		}
	}

	if len(paths) > 0 {
		for _, p3 := range path3CandidateSegments {
			paths.AddPath(p3.Path())
		}
	}

	return paths, nil
}

func adcsESC9aPath1Pattern(domainID graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(
			1, 1,
			query.And(
				query.KindIn(query.Relationship(), ad.GenericWrite, ad.GenericAll, ad.Owns, ad.WriteOwner, ad.WriteDACL),
				query.KindIn(query.End(), ad.Computer, ad.User),
			),
		).
		OutboundWithDepth(
			0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			),
		).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
				query.Kind(query.End(), ad.CertTemplate),
				query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
				query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
				query.Equals(query.EndProperty(ad.NoSecurityExtension.String()), true),
				query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), false),
				query.Or(
					query.Equals(query.EndProperty(ad.SubjectAltRequireUPN.String()), true),
					query.Equals(query.EndProperty(ad.SubjectAltRequireSPN.String()), true),
				),
				query.Or(
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.And(
						query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
						query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					),
				),
			),
		).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		OutboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.KindIn(query.End(), ad.EnterpriseCA, ad.AIACA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainID),
		))
}

func adcsESC9APath2Pattern(caNodes []graph.ID, domainId graph.ID) traversal.PatternContinuation {
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
			query.Equals(query.EndID(), domainId),
		))
}

func adcsESC9APath3Pattern() traversal.PatternContinuation {
	return traversal.NewPattern().
		InboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.TrustedBy),
				query.Equals(query.RelationshipProperty(ad.TrustType.String()), "ParentChild"),
				query.Kind(query.Start(), ad.Domain),
			)).
		Inbound(query.And(
			query.Kind(query.Relationship(), ad.DCFor),
			query.Kind(query.Start(), ad.Computer),
		))
}

func adcsESC9bPath1Pattern(domainID graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(
			1, 1,
			query.And(
				query.KindIn(query.Relationship(), ad.GenericWrite, ad.GenericAll, ad.Owns, ad.WriteOwner, ad.WriteDACL),
				query.KindIn(query.End(), ad.Computer),
			),
		).
		OutboundWithDepth(
			0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			),
		).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
				query.Kind(query.End(), ad.CertTemplate),
				query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
				query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
				query.Equals(query.EndProperty(ad.NoSecurityExtension.String()), true),
				query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), false),
				query.Equals(query.EndProperty(ad.SubjectAltRequireDNS.String()), true),
				query.Or(
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.And(
						query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
						query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					),
				),
			),
		).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		OutboundWithDepth(0, 0, query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.KindIn(query.End(), ad.EnterpriseCA, ad.AIACA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainID),
		))
}

func adcsESC9bPath2Pattern(caNodes []graph.ID, domainId graph.ID) traversal.PatternContinuation {
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
			query.Equals(query.EndID(), domainId),
		))
}

func adcsESC9bPath3Pattern() traversal.PatternContinuation {
	return traversal.NewPattern().
		InboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.TrustedBy),
				query.Equals(query.RelationshipProperty(ad.TrustType.String()), "ParentChild"),
				query.Kind(query.Start(), ad.Domain),
			)).
		Inbound(query.And(
			query.Kind(query.Relationship(), ad.DCFor),
			query.Kind(query.Start(), ad.Computer),
		))
}

func GetADCSESC9bEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH (n {objectid:'S-1-5-21-3933516454-2894985453-2515407000-500'})-[:ADCSESC9b]->(d:Domain {objectid:'S-1-5-21-3933516454-2894985453-2515407000'})
		MATCH p1 = (n)-[:GenericAll|GenericWrite|Owns|WriteOwner|WriteDacl]->(m:Computer)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		WHERE ct.requiresmanagerapproval = false
		AND ct.authenticationenabled = true
		AND ct.nosecurityextension = true
		AND ct.enrolleesuppliessubject = False
		AND ct.subjectaltrequiredns = true
		AND (
			(ct.schemaversion > 1 AND ct.authorizedsignatures = 0)
			OR ct.schemaversion = 1
		)
		MATCH p2 = (m)-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)
		MATCH p3 = (d)<-[r:TrustedBy*0..]-()<-[:DCFor]-(dc:Computer)
		WITH *, relationships(p3) AS r
		WHERE ALL(rel IN r WHERE type(rel) = "DCFor" OR rel.trusttype = "ParentChild")
		AND (
			dc.strongcertificatebindingenforcementraw = 0
			OR dc.strongcertificatebindingenforcementraw = 1
		)

		RETURN p1,p2,p3
	*/

	var (
		startNode *graph.Node
		endNode   *graph.Node

		traversalInst          = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths                  = graph.PathSet{}
		path1CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		victimCANodes          = map[graph.ID][]graph.ID{}
		path2CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path3CandidateSegments = []*graph.PathSegment{}
		p2canodes              = make([]graph.ID, 0)
		nodeMap                = map[graph.ID]*graph.Node{}
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

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: adcsESC9bPath1Pattern(edge.EndID).Do(func(terminal *graph.PathSegment) error {
			victimNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Depth() == 1
			})

			// First ECA in the path
			var caNode *graph.Node
			terminal.Path().Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
				if end.Kinds.ContainsOneOf(ad.EnterpriseCA) {
					caNode = end
					return false
				}
				return true
			})

			lock.Lock()
			path1CandidateSegments[victimNode.ID] = append(path1CandidateSegments[victimNode.ID], terminal)
			nodeMap[victimNode.ID] = victimNode
			victimCANodes[victimNode.ID] = append(victimCANodes[victimNode.ID], caNode.ID)
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	for victim, p1CANodes := range victimCANodes {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: nodeMap[victim],
			Driver: adcsESC9bPath2Pattern(p1CANodes, edge.EndID).Do(func(terminal *graph.PathSegment) error {
				caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
				})

				lock.Lock()
				path2CandidateSegments[caNode.ID] = append(path2CandidateSegments[caNode.ID], terminal)
				p2canodes = append(p2canodes, caNode.ID)
				lock.Unlock()

				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	if len(p2canodes) > 0 {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: endNode,
			Driver: adcsESC9bPath3Pattern().Do(func(terminal *graph.PathSegment) error {
				terminalNode := terminal.Node
				if terminalNode.Kinds.ContainsOneOf(ad.Computer) {
					strongBinding, err := terminalNode.Properties.Get(ad.StrongCertificateBindingEnforcementRaw.String()).Float64()
					if err == nil && (strongBinding == 1 || strongBinding == 0) {
						lock.Lock()
						path3CandidateSegments = append(path3CandidateSegments, terminal)
						lock.Unlock()
					}
				}
				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	for _, p1paths := range path1CandidateSegments {
		for _, p1path := range p1paths {
			// First ECA in the path
			var caNode *graph.Node
			p1path.Path().Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
				if end.Kinds.ContainsOneOf(ad.EnterpriseCA) {
					caNode = end
					return false
				}
				return true
			})

			if p2segments, ok := path2CandidateSegments[caNode.ID]; !ok {
				continue
			} else {
				paths.AddPath(p1path.Path())
				for _, p2 := range p2segments {
					paths.AddPath(p2.Path())
				}
			}
		}
	}

	if len(paths) > 0 {
		for _, p3 := range path3CandidateSegments {
			paths.AddPath(p3.Path())
		}
	}

	return paths, nil
}
