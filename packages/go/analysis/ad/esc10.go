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
	"github.com/specterops/bloodhound/log"
)

func PostADCSESC10a(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca, domain *graph.Node, cache ADCSCache) error {
	if canAbuseUPNRels, err := FetchCanAbuseUPNCertMappingRels(tx, eca); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}
		return err
	} else if len(canAbuseUPNRels) == 0 {
		return nil
	} else if publishedCertTemplates, ok := cache.PublishedTemplateCache[eca.ID]; !ok {
		return nil
	} else if ecaControllers, ok := cache.EnterpriseCAEnrollers[eca.ID]; !ok {
		return nil
	} else {
		results := cardinality.NewBitmap32()

		for _, template := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForESC10(template, false); err != nil {
				log.Warnf("error validating cert template %d: %v", template.ID, err)
				continue
			} else if !valid {
				continue
			} else if certTemplateControllers, ok := cache.CertTemplateControllers[template.ID]; !ok {
				log.Debugf("Failed to retrieve controllers for cert template %d from cache", template.ID)
				continue
			} else {
				victimBitmap := getVictimBitmap(groupExpansions, certTemplateControllers, ecaControllers)

				if filteredVictims, err := filterUserDNSResults(tx, victimBitmap, template); err != nil {
					log.Warnf("error filtering users from victims for esc9a: %v", err)
					continue
				} else if attackers, err := FetchAttackersForEscalations9and10(tx, filteredVictims, false); err != nil {
					log.Warnf("Error getting start nodes for esc10a attacker nodes: %v", err)
					continue
				} else {
					results.Or(cardinality.NodeIDsToDuplex(attackers))
				}
			}
		}

		results.Each(func(value uint32) bool {
			channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC10a,
			})
			return true
		})
	}
	return nil
}

func PostADCSESC10b(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	if canAbuseUPNRels, err := FetchCanAbuseUPNCertMappingRels(tx, enterpriseCA); err != nil {
		if graph.IsErrNotFound(err) {
			return nil
		}
		return err
	} else if len(canAbuseUPNRels) == 0 {
		return nil
	} else if publishedCertTemplates, ok := cache.PublishedTemplateCache[enterpriseCA.ID]; !ok {
		return nil
	} else if ecaControllers, ok := cache.EnterpriseCAEnrollers[enterpriseCA.ID]; !ok {
		return nil
	} else {
		results := cardinality.NewBitmap32()

		for _, template := range publishedCertTemplates {
			if valid, err := isCertTemplateValidForESC10(template, true); err != nil {
				log.Warnf("error validating cert template %d: %v", template.ID, err)
				continue
			} else if !valid {
				continue
			} else if certTemplateControllers, ok := cache.CertTemplateControllers[template.ID]; !ok {
				log.Debugf("Failed to retrieve controllers for cert template %d from cache", template.ID)
				continue
			} else {
				victimBitmap := getVictimBitmap(groupExpansions, certTemplateControllers, ecaControllers)

				if attackers, err := FetchAttackersForEscalations9and10(tx, victimBitmap, true); err != nil {
					log.Warnf("Error getting start nodes for esc10b attacker nodes: %v", err)
					continue
				} else {
					results.Or(cardinality.NodeIDsToDuplex(attackers))
				}
			}
		}

		results.Each(func(value uint32) bool {
			channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC10b,
			})
			return true
		})
	}
	return nil
}

func isCertTemplateValidForESC10(ct *graph.Node, scenarioB bool) (bool, error) {
	if reqManagerApproval, err := ct.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		return false, err
	} else if reqManagerApproval {
		return false, nil
	} else if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
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
			return false, nil
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

func getESC10VictimCriteria(edgeKind graph.Kind) graph.Criteria {
	if edgeKind == ad.ADCSESC10a {
		return query.KindIn(query.End(), ad.Computer, ad.User)
	}
	return query.KindIn(query.End(), ad.Computer)
}

func getESC10CertTemplateCriteria(edgeKind graph.Kind) graph.Criteria {
	if edgeKind == ad.ADCSESC10a {
		return query.Or(
			query.Equals(query.EndProperty(ad.SubjectAltRequireUPN.String()), true),
			query.Equals(query.EndProperty(ad.SubjectAltRequireSPN.String()), true),
		)
	}
	return query.Equals(query.EndProperty(ad.SubjectAltRequireDNS.String()), true)
}

func adcsESC10Path1Pattern(domainID graph.ID, edgeKind graph.Kind) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(
			1, 1,
			query.And(
				query.KindIn(query.Relationship(), ad.GenericWrite, ad.GenericAll, ad.Owns, ad.WriteOwner, ad.WriteDACL),
				getESC10VictimCriteria(edgeKind),
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
				query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), false),
				getESC10CertTemplateCriteria(edgeKind),
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
			query.KindIn(query.Relationship(), ad.PublishedTo, ad.IssuedSignedBy),
			query.Kind(query.End(), ad.EnterpriseCA),
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

func adcsESC10APath3Pattern(caIDs []graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		Inbound(
			query.KindIn(query.Relationship(), ad.DCFor, ad.TrustedBy),
		).
		Inbound(query.And(
			query.Kind(query.Relationship(), ad.CanAbuseUPNCertMapping),
			query.InIDs(query.StartID(), caIDs...),
		))
}

func GetADCSESC10EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/* Scenario A
	MATCH (n {objectid:'S-1-5-21-3933516454-2894985453-2515407000-500'})-[:ADCSESC10a]->(d:Domain {objectid:'S-1-5-21-3933516454-2894985453-2515407000'})
	OPTIONAL MATCH p1 = (n)-[:GenericAll|GenericWrite|Owns|WriteOwner|WriteDacl]->(m)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
	WHERE ct.requiresmanagerapproval = false
	  AND ct.authenticationenabled = true
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
	OPTIONAL MATCH p2 = (m)-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)
	OPTIONAL MATCH p3 = (ca)-[:CanAbuseUPNCertMapping|DCFor|TrustedBy*1..]->(d)
	RETURN p1,p2,p3*/

	/* Scenario B
	MATCH (n {objectid:'S-1-5-21-3933516454-2894985453-2515407000-500'})-[:ADCSESC10b]->(d:Domain {objectid:'S-1-5-21-3933516454-2894985453-2515407000'})
	OPTIONAL MATCH p1 = (n)-[:GenericAll|GenericWrite|Owns|WriteOwner|WriteDacl]->(m:Computer)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
	WHERE ct.requiresmanagerapproval = false
	AND ct.authenticationenabled = true
	AND ct.enrolleesuppliessubject = False
	AND ct.subjectaltrequiredns = true
	AND (
		(ct.schemaversion > 1 AND ct.authorizedsignatures = 0)
		OR ct.schemaversion = 1
	)
	OPTIONAL MATCH p2 = (m)-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)
	OPTIONAL MATCH p3 = (ca)-[:CanAbuseUPNCertMapping|DCFor|TrustedBy*1..]->(d)
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
		path3CandidateSegments = map[graph.ID][]*graph.PathSegment{}
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
		Driver: adcsESC10Path1Pattern(edge.EndID, edge.Kind).Do(func(terminal *graph.PathSegment) error {
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

			caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
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

	//We can re-use p2 from ESC9a, since they're the same
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
			Driver: adcsESC10APath3Pattern(p2canodes).Do(func(terminal *graph.PathSegment) error {
				caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
				})

				lock.Lock()
				path3CandidateSegments[caNode.ID] = append(path3CandidateSegments[caNode.ID], terminal)
				lock.Unlock()
				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	for _, p1paths := range path1CandidateSegments {
		for _, p1path := range p1paths {
			caNode := p1path.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			if p2segments, ok := path2CandidateSegments[caNode.ID]; !ok {
				continue
			} else if p3segments, ok := path3CandidateSegments[caNode.ID]; !ok {
				continue
			} else {
				paths.AddPath(p1path.Path())
				for _, p2 := range p2segments {
					paths.AddPath(p2.Path())
				}

				for _, p3 := range p3segments {
					paths.AddPath(p3.Path())
				}
			}
		}
	}

	return paths, nil
}
