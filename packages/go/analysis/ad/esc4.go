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

func PostADCSESC4(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, enterpriseCA, domain *graph.Node, cache ADCSCache) error {
	// 1.
	principals := cardinality.NewBitmap32()

	// 2. iterate certtemplates that have an outbound `PublishedTo` edge to eca
	for _, certTemplate := range cache.PublishedTemplateCache[enterpriseCA.ID] {
		if principalsWithGenericWrite, err := FetchPrincipalsWithGenericWriteOnCertTemplate(tx, certTemplate); err != nil {
			log.Warnf("error fetching principals with %s on cert template: %v", ad.GenericWrite, err)
		} else if principalsWithEnrollOrAllExtendedRights, err := FetchPrincipalsWithEnrollOrAllExtendedRightsOnCertTemplate(tx, certTemplate); err != nil {
			log.Warnf("error fetching principals with %s or %s on cert template: %v", ad.Enroll, ad.AllExtendedRights, err)
		} else if principalsWithPKINameFlag, err := FetchPrincipalsWithWritePKINameFlagOnCertTemplate(tx, certTemplate); err != nil {
			log.Warnf("error fetching principals with %s on cert template: %v", ad.WritePKINameFlag, err)
		} else if principalsWithPKIEnrollmentFlag, err := FetchPrincipalsWithWritePKIEnrollmentFlagOnCertTemplate(tx, certTemplate); err != nil {
			log.Warnf("error fetching principals with %s on cert template: %v", ad.WritePKIEnrollmentFlag, err)
		} else if enrolleeSuppliesSubject, err := certTemplate.Properties.Get(string(ad.EnrolleeSuppliesSubject)).Bool(); err != nil {
			log.Warnf("error fetching %s property on cert template: %v", ad.EnrolleeSuppliesSubject, err)
		} else if requiresManagerApproval, err := certTemplate.Properties.Get(string(ad.RequiresManagerApproval)).Bool(); err != nil {
			log.Warnf("error fetching %s property on cert template: %v", ad.RequiresManagerApproval, err)
		} else {

			// 2a. principals that control the cert template
			principals.Or(
				CalculateCrossProductNodeSets(
					groupExpansions,
					cache.EnterpriseCAEnrollers[enterpriseCA.ID],
					cache.CertTemplateControllers[certTemplate.ID],
				))

			// 2b. principals with `Enroll/AllExtendedRights` + `Generic Write` combination on the cert template
			principals.Or(
				CalculateCrossProductNodeSets(
					groupExpansions,
					cache.EnterpriseCAEnrollers[enterpriseCA.ID],
					principalsWithGenericWrite.Slice(),
					principalsWithEnrollOrAllExtendedRights.Slice(),
				),
			)

			// 2c. kick out early if cert template does meet conditions for ESC4
			if valid, err := isCertTemplateValidForESC4(certTemplate); err != nil {
				log.Warnf("error validating cert template %d: %v", certTemplate.ID, err)
				continue
			} else if !valid {
				continue
			}

			// 2d. principals with `Enroll/AllExtendedRights` + `WritePKINameFlag` + `WritePKIEnrollmentFlag` on the cert template
			principals.Or(
				CalculateCrossProductNodeSets(
					groupExpansions,
					cache.EnterpriseCAEnrollers[enterpriseCA.ID],
					principalsWithEnrollOrAllExtendedRights.Slice(),
					principalsWithPKINameFlag.Slice(),
					principalsWithPKIEnrollmentFlag.Slice(),
				),
			)

			// 2e.
			if enrolleeSuppliesSubject {
				principals.Or(
					CalculateCrossProductNodeSets(
						groupExpansions,
						cache.EnterpriseCAEnrollers[enterpriseCA.ID],
						principalsWithEnrollOrAllExtendedRights.Slice(),
						principalsWithPKIEnrollmentFlag.Slice(),
					),
				)
			}

			// 2f.
			if !requiresManagerApproval {
				principals.Or(
					CalculateCrossProductNodeSets(
						groupExpansions,
						cache.EnterpriseCAEnrollers[enterpriseCA.ID],
						principalsWithEnrollOrAllExtendedRights.Slice(),
						principalsWithPKINameFlag.Slice(),
					),
				)
			}
		}
	}

	principals.Each(func(value uint32) bool {
		channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
			FromID: graph.ID(value),
			ToID:   domain.ID,
			Kind:   ad.ADCSESC4,
		})
		return true
	})

	return nil
}

func isCertTemplateValidForESC4(ct *graph.Node) (bool, error) {
	if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
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

func FetchPrincipalsWithGenericWriteOnCertTemplate(tx graph.Transaction, certTemplate *graph.Node) (graph.NodeSet, error) {
	if nodes, err := ops.FetchStartNodes(tx.Relationships().Filterf(
		func() graph.Criteria {
			return query.And(
				query.Equals(query.EndID(), certTemplate.ID),
				query.Kind(query.Relationship(), ad.GenericWrite),
			)
		},
	)); err != nil {
		return nil, err
	} else {
		return nodes, nil
	}
}

func FetchPrincipalsWithEnrollOrAllExtendedRightsOnCertTemplate(tx graph.Transaction, certTemplate *graph.Node) (graph.NodeSet, error) {
	if nodes, err := ops.FetchStartNodes(
		tx.Relationships().Filterf(
			func() graph.Criteria {
				return query.And(
					query.Equals(query.EndID(), certTemplate.ID),
					query.Or(
						query.Kind(query.Relationship(), ad.Enroll),
						query.Kind(query.Relationship(), ad.AllExtendedRights),
					),
				)
			},
		)); err != nil {
		return nil, err
	} else {
		return nodes, nil
	}
}

func FetchPrincipalsWithWritePKINameFlagOnCertTemplate(tx graph.Transaction, certTemplate *graph.Node) (graph.NodeSet, error) {
	if nodes, err := ops.FetchStartNodes(
		tx.Relationships().Filterf(
			func() graph.Criteria {
				return query.And(
					query.Equals(query.EndID(), certTemplate.ID),
					query.Kind(query.Relationship(), ad.WritePKINameFlag),
				)
			},
		)); err != nil {
		return nil, err
	} else {
		return nodes, nil
	}
}

func FetchPrincipalsWithWritePKIEnrollmentFlagOnCertTemplate(tx graph.Transaction, certTemplate *graph.Node) (graph.NodeSet, error) {
	if nodes, err := ops.FetchStartNodes(
		tx.Relationships().Filterf(
			func() graph.Criteria {
				return query.And(
					query.Equals(query.EndID(), certTemplate.ID),
					query.Kind(query.Relationship(), ad.WritePKIEnrollmentFlag),
				)
			},
		)); err != nil {
		return nil, err
	} else {
		return nodes, nil
	}
}

// composition: p1
func findPathsToDomainThroughCertTemplateWithGenericAll(
	ctx context.Context,
	db graph.Database,
	startNode *graph.Node,
	domainID graph.ID,
	enterpriseCAs cardinality.Duplex[uint32],
) (map[graph.ID][]*graph.PathSegment, cardinality.Duplex[uint32], error) {

	var (
		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		lock          = &sync.Mutex{}

		certTemplateSegments = map[graph.ID][]*graph.PathSegment{}
		certTemplates        = cardinality.NewBitmap32()
	)

	// p1: use the enterpriseCA nodes to gather the set of cert templates with an inbound `GenericAll`
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: certTemplateWithPrivelegesToDomainTraversal(
			[]graph.Kind{ad.GenericAll, ad.Owns, ad.WriteOwner, ad.WriteDACL},
			domainID,
			enterpriseCAs,
		).Do(
			func(terminal *graph.PathSegment) error {

				certTemplate := terminal.Search(
					func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
					},
				)

				lock.Lock()

				certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
				certTemplates.Add(certTemplate.ID.Uint32())

				lock.Unlock()

				return nil
			}),
	}); err != nil {
		return certTemplateSegments, certTemplates, err
	} else {
		return certTemplateSegments, certTemplates, nil
	}
}

// composition: p3 + p4
func findPathsToDomainThroughCertTemplateWithGenericWrite(
	ctx context.Context,
	db graph.Database,
	startNode *graph.Node,
	domainID graph.ID,
	enterpriseCAs cardinality.Duplex[uint32],
) (map[graph.ID][]*graph.PathSegment, cardinality.Duplex[uint32], error) {

	var (
		inboundEdgeToCertTemplate = ad.GenericWrite
		criteriaForCertTemplate   graph.Criteria // there aren't any!
	)

	return traversalToDomainThroughCertTemplate(ctx, db, startNode, domainID, enterpriseCAs, inboundEdgeToCertTemplate, criteriaForCertTemplate)
}

// composition: p6 + p7
func findPathsToDomainThroughCertTemplateWithWritePKINameFlag(
	ctx context.Context,
	db graph.Database,
	startNode *graph.Node,
	domainID graph.ID,
	enterpriseCAs cardinality.Duplex[uint32],
) (map[graph.ID][]*graph.PathSegment, cardinality.Duplex[uint32], error) {

	var (
		inboundEdgeToCertTemplate = ad.WritePKINameFlag
		criteriaForCertTemplate   = query.And(
			query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
			query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
			query.Or(
				query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
				query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
			),
		)
	)

	return traversalToDomainThroughCertTemplate(ctx, db, startNode, domainID, enterpriseCAs, inboundEdgeToCertTemplate, criteriaForCertTemplate)

}

// composition: p9 + p10
func findPathsToDomainThroughCertTemplateWithWritePKIEnrollmentFlag(
	ctx context.Context,
	db graph.Database,
	startNode *graph.Node,
	domainID graph.ID,
	enterpriseCAs cardinality.Duplex[uint32],
) (map[graph.ID][]*graph.PathSegment, cardinality.Duplex[uint32], error) {

	var (
		inboundEdgeToCertTemplate = ad.WritePKIEnrollmentFlag
		criteriaForCertTemplate   = query.And(
			query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), true),
			query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
			query.Or(
				query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
				query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
			),
		)
	)

	return traversalToDomainThroughCertTemplate(ctx, db, startNode, domainID, enterpriseCAs, inboundEdgeToCertTemplate, criteriaForCertTemplate)

}

func traversalToDomainThroughCertTemplate(
	ctx context.Context,
	db graph.Database,
	startNode *graph.Node,
	domainID graph.ID,
	enterpriseCAs cardinality.Duplex[uint32],
	inboundEdgeToCertTemplate graph.Kind,
	criteriaForCertTemplate graph.Criteria,
) (map[graph.ID][]*graph.PathSegment, cardinality.Duplex[uint32], error) {

	var (
		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		lock          = &sync.Mutex{}

		certTemplateSegments = map[graph.ID][]*graph.PathSegment{}
		certTemplates        = cardinality.NewBitmap32()
	)

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: certTemplateWithPrivelegesToDomainTraversal([]graph.Kind{inboundEdgeToCertTemplate}, domainID, enterpriseCAs).Do(
			func(terminal *graph.PathSegment) error {

				certTemplate := terminal.Search(
					func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
					},
				)

				lock.Lock()
				certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
				certTemplates.Add(certTemplate.ID.Uint32())
				lock.Unlock()

				return nil
			}),
	}); err != nil {
		return certTemplateSegments, certTemplates, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: certTemplateWithEnrollmentRightsTraversal(certTemplates, criteriaForCertTemplate).Do(
			func(terminal *graph.PathSegment) error {

				certTemplate := terminal.Search(
					func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
					},
				)

				lock.Lock()
				certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
				lock.Unlock()

				return nil
			}),
	}); err != nil {
		return certTemplateSegments, certTemplates, err
	}

	return certTemplateSegments, certTemplates, nil
}

// composition: p12, p13, p14
func findPathsToDomainThroughCertTemplateWithBothPKIFlags(
	ctx context.Context,
	db graph.Database,
	startNode *graph.Node,
	domainID graph.ID,
	enterpriseCAs cardinality.Duplex[uint32],
) (map[graph.ID][]*graph.PathSegment, cardinality.Duplex[uint32], error) {

	var (
		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		lock          = &sync.Mutex{}

		certTemplateSegments = map[graph.ID][]*graph.PathSegment{}
		certTemplates        = cardinality.NewBitmap32()
	)

	// p9: use the enterpriseCA nodes to gather the set of cert templates with an inbound `WritePKIEnrollmentFlag`
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: certTemplateWithPrivelegesToDomainTraversal([]graph.Kind{ad.WritePKIEnrollmentFlag}, domainID, enterpriseCAs).Do(
			func(terminal *graph.PathSegment) error {

				certTemplate := terminal.Search(
					func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
					},
				)

				lock.Lock()
				certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
				certTemplates.Add(certTemplate.ID.Uint32())
				lock.Unlock()

				return nil
			}),
	}); err != nil {
		return certTemplateSegments, certTemplates, err
	}

	// p10: (reuse p4 logic): find cert templates that have an inbound `Enroll` OR `AllExtendedRights` edge
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: certTemplateWithEnrollmentRightsTraversal(certTemplates, nil).Do(
			func(terminal *graph.PathSegment) error {

				certTemplate := terminal.Search(
					func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
					},
				)

				lock.Lock()
				certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
				lock.Unlock()

				return nil
			}),
	}); err != nil {
		return certTemplateSegments, certTemplates, err
	}

	// p14: find cert templates with valid combination of properties that has an inbound `WritePKIName` edge
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: certTemplateWithPKINameFlagTraversal(certTemplates).Do(
			func(terminal *graph.PathSegment) error {

				certTemplate := terminal.Search(
					func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
					},
				)

				lock.Lock()
				certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
				lock.Unlock()

				return nil
			}),
	}); err != nil {
		return certTemplateSegments, certTemplates, err
	}

	return certTemplateSegments, certTemplates, nil
}

// composition: p2, p5, p8, p11, p15
func findPathToDomainThroughEnterpriseCAsTrustedForNTAuth(
	ctx context.Context,
	db graph.Database,
	startNode *graph.Node,
	domainID graph.ID,
) (
	map[graph.ID][]*graph.PathSegment,
	cardinality.Duplex[uint32],
	error,
) {

	var (
		traversalInst = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		lock          = &sync.Mutex{}

		enterpriseCASegments = map[graph.ID][]*graph.PathSegment{}
		enterpriseCAs        = cardinality.NewBitmap32()
	)

	// Start by fetching all EnterpriseCA nodes that our user has enrollment rights on via group membership or directly
	if err := traversalInst.BreadthFirst(ctx,
		traversal.Plan{
			Root: startNode,
			Driver: enterpriseCATraversal().Do(
				func(terminal *graph.PathSegment) error {

					enterpriseCA := terminal.Search(
						func(nextSegment *graph.PathSegment) bool {
							return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
						},
					)

					lock.Lock()
					enterpriseCAs.Add(enterpriseCA.ID.Uint32())
					lock.Unlock()

					return nil
				}),
		}); err != nil {

		return enterpriseCASegments, enterpriseCAs, err
	}

	// every scenario must contain a path from the enterpriseCA nodes found in the previous step to find enterprise CAs that are trusted for NTAuth
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ntAuthStoreToDomainTraversal(domainID, enterpriseCAs).Do(
			func(terminal *graph.PathSegment) error {

				enterpriseCA := terminal.Search(
					func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
					},
				)

				lock.Lock()
				enterpriseCASegments[enterpriseCA.ID] = append(enterpriseCASegments[enterpriseCA.ID], terminal)
				lock.Unlock()

				return nil
			}),
	}); err != nil {
		return enterpriseCASegments, enterpriseCAs, err
	}

	return enterpriseCASegments, enterpriseCAs, nil

}

func GetADCSESC4EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH p1 = (n1)-[:MemberOf*0..]->()-[:GenericAll|Owns|WriteOwner|WriteDacl]->(ct)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		MATCH p2 = (n1)-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)

		MATCH p3 = (n2)-[:MemberOf*0..]->()-[:GenericWrite]->(ct2)-[:PublishedTo]->(ca2)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		MATCH p4 = (n2)-[:MemberOf*0..]->()-[:Enroll|AllExtendedRights]->(ct2)
		MATCH p5 = (n2)-[:MemberOf*0..]->()-[:Enroll]->(ca2)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)

		MATCH p6 = (n3)-[:MemberOf*0..]->()-[:WritePKINameFlag]->(ct3)-[:PublishedTo]->(ca3)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		MATCH p7 = (n3)-[:MemberOf*0..]->()-[:Enroll|AllExtendedRights]->(ct3)
		WHERE ct3.requiresmanagerapproval = false
		  AND ct3.authenticationenabled = true
		  AND (
		    ct3.authorizedsignatures = 0 OR ct3.schemaversion = 1
		  )
		MATCH p8 = (n3)-[:MemberOf*0..]->()-[:Enroll]->(ca3)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)


		MATCH p9 = (n4)-[:MemberOf*0..]->()-[:WritePKIEnrollmentFlag]->(ct4)-[:PublishedTo]->(ca4)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		MATCH p10 = (n4)-[:MemberOf*0..]->()-[:Enroll|AllExtendedRights]->(ct4)
		WHERE ct4.enrolleesuppliessubject = true
		  AND ct4.authenticationenabled = true
		  AND (
		    ct4.authorizedsignatures = 0 OR ct4.schemaversion = 1
		  )
		MATCH p11 = (n4)-[:MemberOf*0..]->()-[:Enroll]->(ca4)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)

		MATCH p12 = (n5)-[:MemberOf*0..]->()-[:WritePKIEnrollmentFlag]->(ct5)-[:PublishedTo]->(ca5)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		MATCH p13 = (n5)-[:MemberOf*0..]->()-[:Enroll|AllExtendedRights]->(ct5)
		MATCH p14 = (n5)-[:MemberOf*0..]->()-[:WritePKINameFlag]->(ct5)
		WHERE ct5.authenticationenabled = true
		  AND (
		    ct5.authorizedsignatures = 0 OR ct5.schemaversion = 1
		  )
		MATCH p15 = (n5)-[:MemberOf*0..]->()-[:Enroll]->(ca5)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)

		RETURN p1,p2,p3,p4,p5,p6,p7,p8,p9,p10,p11,p12,p13,p14,p15
	*/

	var (
		startNode *graph.Node
		domainID  = edge.EndID
		paths     = graph.PathSet{}

		enterpriseCASegments = map[graph.ID][]*graph.PathSegment{}
		enterpriseCAs        = cardinality.NewBitmap32()
	)

	// hydrate the start node
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else {
			startNode = node
			return nil
		}
	}); err != nil {
		return nil, err
	}

	// p2, p5, p8, p11, p15: these pattern parts use same logic. find the paths to domain like: (eca) -> (nt auth store)-> (domain)
	if paths, ecaIDs, err := findPathToDomainThroughEnterpriseCAsTrustedForNTAuth(ctx, db, startNode, domainID); err != nil {
		return nil, err
	} else {
		enterpriseCASegments = paths
		enterpriseCAs = ecaIDs
	}

	// p1, p2
	if pathsToDomain, certTemplateIDs, err := findPathsToDomainThroughCertTemplateWithGenericAll(ctx, db, startNode, domainID, enterpriseCAs); err != nil {
		return nil, err
	} else {
		certTemplateIDs.Each(
			func(value uint32) bool {
				// add the paths which satisfy p1-p2 requirements
				for _, segment := range pathsToDomain[graph.ID(value)] {
					paths.AddPath(segment.Path())
				}

				return true
			},
		)
	}

	// p3, p4, p5
	if pathsToDomain, certTemplateIDs, err := findPathsToDomainThroughCertTemplateWithGenericWrite(ctx, db, startNode, domainID, enterpriseCAs); err != nil {
		return nil, err
	} else {
		certTemplateIDs.Each(
			func(value uint32) bool {
				// add the paths which satisfy p3, p4, and p5 requirements
				for _, segment := range pathsToDomain[graph.ID(value)] {
					paths.AddPath(segment.Path())
				}

				return true
			},
		)
	}

	// p6, p7, p8
	if pathsToDomain, certTemplateIDs, err := findPathsToDomainThroughCertTemplateWithWritePKINameFlag(ctx, db, startNode, domainID, enterpriseCAs); err != nil {
		return nil, err
	} else {
		certTemplateIDs.Each(
			func(value uint32) bool {
				// add the paths which satisfy p6, p7, and p8 requirements
				for _, segment := range pathsToDomain[graph.ID(value)] {
					paths.AddPath(segment.Path())
				}

				return true
			},
		)
	}

	// p9, p10, p11
	if pathsToDomain, certTemplateIDs, err := findPathsToDomainThroughCertTemplateWithWritePKIEnrollmentFlag(ctx, db, startNode, domainID, enterpriseCAs); err != nil {
		return nil, err
	} else {
		certTemplateIDs.Each(
			func(value uint32) bool {
				// add the paths which satisfy p9, p10, and p11 requirements
				for _, segment := range pathsToDomain[graph.ID(value)] {
					paths.AddPath(segment.Path())
				}

				return true
			},
		)
	}

	// p12, p13, p14, p15
	if pathsToDomain, certTemplateIDs, err := findPathsToDomainThroughCertTemplateWithBothPKIFlags(ctx, db, startNode, domainID, enterpriseCAs); err != nil {
		return nil, err
	} else {
		certTemplateIDs.Each(
			func(value uint32) bool {
				// add the paths which satisfy p12, p13, p14, p15 requirements
				for _, segment := range pathsToDomain[graph.ID(value)] {
					paths.AddPath(segment.Path())
				}

				return true
			},
		)
	}

	// if we have found paths already, materialize the paths we found from the enterprise CAs -> NTAuthStore -> Domain
	if paths.Len() > 0 {
		enterpriseCAs.Each(
			func(value uint32) bool {
				for _, segment := range enterpriseCASegments[graph.ID(value)] {
					paths.AddPath(segment.Path())
				}
				return true
			})
	}

	return paths, nil
}

func enterpriseCATraversal() traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.Enroll),
				query.KindIn(query.End(), ad.EnterpriseCA),
			))
}

func ntAuthStoreToDomainTraversal(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.Enroll),
				query.KindIn(query.End(), ad.EnterpriseCA),
				query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
			)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
				query.Kind(query.End(), ad.NTAuthStore),
			)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
				query.Equals(query.EndID(), domainId),
			))
}

// This traversal goes from principal -> domain via a cert template that has an inbound edge(s) corresponding to whatever `priveleges` are provided
func certTemplateWithPrivelegesToDomainTraversal(priveleges graph.Kinds, domainID graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), priveleges...),
				query.Kind(query.End(), ad.CertTemplate),
			)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.PublishedTo),
				query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
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
				query.Equals(query.EndID(), domainID),
			))
}

func certTemplateWithEnrollmentRightsTraversal(certTemplates cardinality.Duplex[uint32], criteria graph.Criteria) traversal.PatternContinuation {
	// start with outbound group membership of any depth
	initialTraversal := traversal.NewPattern().
		OutboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			))

	if criteria == nil {
		return initialTraversal.Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.Enroll, ad.AllExtendedRights),
				query.InIDs(query.End(), cardinality.DuplexToGraphIDs(certTemplates)...),
			))
	} else {
		return initialTraversal.Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.Enroll, ad.AllExtendedRights),
				query.InIDs(query.End(), cardinality.DuplexToGraphIDs(certTemplates)...),
				criteria,
			))
	}
}

func certTemplateWithPKINameFlagTraversal(certTemplates cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.WritePKINameFlag),
				query.InIDs(query.End(), cardinality.DuplexToGraphIDs(certTemplates)...),
				// ct.authenticationenabled == true
				query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
				query.Or(
					// ct.authorizedsignatures == 0
					query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					// ct.schemaversion == 1
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
				),
			))
}
