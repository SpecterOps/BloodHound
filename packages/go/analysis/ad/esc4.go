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
		closureErr           error
		startNode            *graph.Node
		traversalInst        = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		lock                 = &sync.Mutex{}
		paths                = graph.PathSet{}
		certTemplateSegments = map[graph.ID][]*graph.PathSegment{}
		enterpriseCASegments = map[graph.ID][]*graph.PathSegment{}
		pattern1Segments     = map[graph.ID][]*graph.PathSegment{}
		certTemplates        = cardinality.NewBitmap32()
		enterpriseCAs        = cardinality.NewBitmap32()
		path1EnterpriseCAs   = cardinality.NewBitmap32()
	)

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

	// Start by fetching all EnterpriseCA nodes that our user has Enroll rights on via group membership or directly
	if err := traversalInst.BreadthFirst(ctx,
		traversal.Plan{
			Root: startNode,
			Driver: enterpriseCAsForPrincipal().Do(
				func(terminal *graph.PathSegment) error {

					enterpriseCA := terminal.Search(
						func(nextSegment *graph.PathSegment) bool {
							return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
						})

					lock.Lock()
					path1EnterpriseCAs.Add(enterpriseCA.ID.Uint32())
					lock.Unlock()

					return nil
				}),
		}); err != nil {
		return nil, err
	}

	// use the enterpriseCA nodes from the previous step to gather the first set of cert templates for p1
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ESC4Path1Pattern(edge.EndID, path1EnterpriseCAs).Do(
			func(terminal *graph.PathSegment) error {
				certTemplate := terminal.Search(
					func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
					})

				lock.Lock()
				certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
				certTemplates.Add(certTemplate.ID.Uint32())
				lock.Unlock()

				return nil
			})}); err != nil {
		return nil, err
	}

	// use the enterpriseCA and certTemplate nodes from previous steps to find enterprise CAs that are trusted for NTAuth (p2)
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6Path2Pattern(edge.EndID, path1EnterpriseCAs, certTemplates).Do(
			func(terminal *graph.PathSegment) error {
				certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
				})

				lock.Lock()
				certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
				certTemplates.Add(certTemplate.ID.Uint32())
				lock.Unlock()

				return nil
			})}); err != nil {
		return nil, err
	}

	// find the enterpriseCAs that have an outbound CanAbuseWeakCertBinding if 6a or a CanAbuseUPNCertMapping if 6b to a computer that is a DC for the domain (p3)
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6Path3Pattern(edge.EndID, path1EnterpriseCAs, edge.Kind).Do(func(terminal *graph.PathSegment) error {
			enterpriseCA := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			paths.AddPath(terminal.Path())
			enterpriseCASegments[enterpriseCA.ID] = append(enterpriseCASegments[enterpriseCA.ID], terminal)
			enterpriseCAs.Add(enterpriseCA.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	certTemplates.Each(func(value uint32) bool {
		var certTemplate *graph.Node

		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			if node, err := ops.FetchNode(tx, graph.ID(value)); err != nil {
				return err
			} else {
				certTemplate = node
				return nil
			}
		}); err != nil {
			closureErr = fmt.Errorf("could not fetch cert template node: %w", err)
			return false
		}

		for _, segment := range certTemplateSegments[graph.ID(value)] {
			if startNode.Kinds.ContainsOneOf(ad.User) {
				if !certTemplateValidForUserVictim(certTemplate) {
					continue
				} else if checkEmailValidity(startNode, certTemplate) {
					continue
				} else {
					paths.AddPath(segment.Path())
				}
			} else if startNode.Kinds.ContainsOneOf(ad.Computer) {
				if checkEmailValidity(startNode, certTemplate) {
					continue
				} else {
					paths.AddPath(segment.Path())
				}
			} else {
				paths.AddPath(segment.Path())
			}
		}
		return true
	})

	if closureErr != nil {
		return paths, closureErr
	}

	if paths.Len() > 0 {
		enterpriseCAs.Each(func(value uint32) bool {
			for _, segment := range enterpriseCASegments[graph.ID(value)] {
				paths.AddPath(segment.Path())
			}
			return true
		})
	}

	return paths, nil
}

func ESC4Path1Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			)).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.GenericAll, ad.Owns, ad.WriteOwner, ad.WriteDACL),
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
				query.Equals(query.EndID(), domainId),
			))
}
