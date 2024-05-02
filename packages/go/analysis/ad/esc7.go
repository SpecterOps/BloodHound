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

func PostADCSESC7(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca, domain *graph.Node, cache ADCSCache) error {
	if certTemplates, ok := cache.DomainCertTemplates[domain.ID]; !ok {
		return nil
	} else if firstDegreeCAManagers, err := fetchFirstDegreeNodes(tx, eca, ad.ManageCA); err != nil {
		log.Errorf("Error fetching CA managers for enterprise ca %d: %v", eca.ID, err)
		return nil
	} else {
		var (
			results            = cardinality.NewBitmap32()
			validCertTemplates []*graph.Node
		)

		for _, certTemplate := range certTemplates {
			if valid, err := isCertTemplateValidForESC7a(certTemplate, domain); err != nil {
				log.Warnf("Error validating cert template %d: %v", certTemplate.ID, err)
				continue
			} else if valid {
				validCertTemplates = append(validCertTemplates, certTemplate)
			}
		}

		roleSeparationEnabled, err := eca.Properties.Get(ad.RoleSeparationEnabled.String()).Bool()
		if err != nil || !roleSeparationEnabled {
			if len(validCertTemplates) > 0 {
				for _, principal := range firstDegreeCAManagers {
					results.Add(principal.ID.Uint32())
				}
			}
		} else {
			for _, validCertTemplate := range validCertTemplates {
				for _, enroller := range cache.CertTemplateEnrollers[validCertTemplate.ID] {
					results.Or(CalculateCrossProductNodeSets(groupExpansions, graph.NewNodeSet(enroller).Slice(), firstDegreeCAManagers.Slice()))
				}
			}
		}

		results.Each(func(value uint32) bool {
			channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
				FromID: graph.ID(value),
				ToID:   domain.ID,
				Kind:   ad.ADCSESC7,
			})
			return true
		})
		return nil
	}
	return nil
}

func isCertTemplateValidForESC7a(ct *graph.Node, d *graph.Node) (bool, error) {
	if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		return false, err
	} else if !authenticationEnabled {
		return false, nil
	} else if enrolleeSuppliesSubject, err := ct.Properties.Get(ad.EnrolleeSuppliesSubject.String()).Bool(); err != nil {
		return false, err
	} else if !enrolleeSuppliesSubject {
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

func GetADCSESC7EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH p0 = (n {objectid:'<principal SID>'})-[:ADCSESC7a]->(d {objectid:'<domain SID>'})
		MATCH p1 = (n)-[:MemberOf*0..]->()-[:ManageCA]->(eca:EnterpriseCA)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)-[:Contains*1..]->(ct:CertTemplate)
		WHERE ct.authenticationenabled = true
			AND (ct.authorizedsignatures = 0 OR ct.schemaversion = 1)
			AND ct.enrolleesuppliessubject = true
		MATCH p2 = (:Computer)-[:HostsCAService]->(eca)-[:TrustedForNTAuth|NTAuthStoreFor*1..]->(d)

		OPTIONAL MATCH p3 = (n)-[:MemberOf*0..]->()-[:Enroll|GenericAll|AllExtendedRights]->(ct)

		WITH *
		WHERE 
			p3 IS NOT NULL
			OR NOT eca.roleseparationenabled = true

		RETURN p0,p1,p2,p3,p4
	*/

	var (
		startNode            *graph.Node
		endNode              *graph.Node
		traversalInst          = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths                  = graph.PathSet{}
		path1EnterpriseCAs     = cardinality.NewBitmap32()
		path1CertTemplates     = cardinality.NewBitmap32()
		path2EnterpriseCAs     = cardinality.NewBitmap32()
		path3CertTemplates     = cardinality.NewBitmap32()
		path1CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path2CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path3CandidateSegments = map[graph.ID][]*graph.PathSegment{}
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

	// P1
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: adcsESC7Path1Pattern(endNode.ID).Do(func(terminal *graph.PathSegment) error {
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
			path1CertTemplates.Add(certTemplate.ID.Uint32())
			lock.Unlock()
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	// P2
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: endNode,
		Driver: adcsESC7Path2Pattern(cardinality.DuplexToGraphIDs(path1EnterpriseCAs)).Do(func(terminal *graph.PathSegment) error {
			enterpriseCANode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
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

	// P3
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: adcsESC7Path3Pattern(cardinality.DuplexToGraphIDs(path1CertTemplates)).Do(func(terminal *graph.PathSegment) error {
			certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			path3CandidateSegments[certTemplate.ID] = append(path3CandidateSegments[certTemplate.ID], terminal)
			path3CertTemplates.Add(certTemplate.ID.Uint32())
			lock.Unlock()
			return nil
		}),
	}); err != nil {
		return nil, err
	}

	//And the 2 bitmaps together to ensure that only enterprise CAs thats satisfy both paths are valid
	path1EnterpriseCAs.And(path2EnterpriseCAs)

	for _, path1Segments := range path1CandidateSegments {
		for _, p1 := range path1Segments {
			//Find the cert template in path 1
			certTemplate := p1.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			//Find the enterprise ca in path 1
			enterpriseCA := p1.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			p3segments, p3Exist := path3CandidateSegments[certTemplate.ID];
			roleSeparationEnabled, err := enterpriseCA.Properties.Get(ad.RoleSeparationEnabled.String()).Bool()
			
			if !p3Exist && (err != nil || roleSeparationEnabled) {
				continue
			} else {
				//Merge all our paths together
				paths.AddPath(p1.Path())

				p2segments, p2Exist := path2CandidateSegments[enterpriseCA.ID];
				if p2Exist {
					for _, p2 := range p2segments {
						paths.AddPath(p2.Path())
					}
				}

				if p3Exist {
					for _, p3 := range p3segments {
						paths.AddPath(p3.Path())
					}	
				}
			}
		}
	}

	return paths, nil
}

func adcsESC7Path1Pattern(domainId graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
		)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.ManageCA),
				query.Kind(query.End(), ad.EnterpriseCA),
		)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
				query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.RootCAFor),
				query.Equals(query.EndID(), domainId),
		)).
		OutboundWithDepth(4, 4,
			query.And(
				query.KindIn(query.Relationship(), ad.Contains),
				query.Kind(query.End(), ad.Container),
		)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.Contains),
				query.Kind(query.End(), ad.CertTemplate),
				query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
				query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), true),
				query.Or(
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
				),
			),
		)
}

func adcsESC7Path2Pattern(enterpriseCAs []graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		Inbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.Kind(query.Start(), ad.NTAuthStore),
		)).
		Inbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.InIDs(query.Start(), enterpriseCAs...),
		)).
		Inbound(query.And(
			query.KindIn(query.Relationship(), ad.HostsCAService),
			query.Kind(query.Start(), ad.Computer),
		))
}

func adcsESC7Path3Pattern(certTemplates []graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(
			0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
		)).
		OutboundWithDepth(
			1, 1,
			query.And(
				query.KindIn(query.Relationship(), ad.Enroll, ad.GenericAll, ad.AllExtendedRights),
				query.InIDs(query.End(), certTemplates...),
		))
}
