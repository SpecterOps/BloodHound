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
	"github.com/specterops/bloodhound/dawgs/traversal"

	"github.com/specterops/bloodhound/dawgs/cardinality"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

func GetEdgeCompositionPath(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	pathSet := graph.NewPathSet()
	return pathSet, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if edge.Kind == ad.GoldenCert {
			if results, err := getGoldenCertEdgeComposition(tx, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC1 {
			if results, err := GetADCSESC1EdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC3 {
			if results, err := GetADCSESC3EdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC6a {
			if results, err := GetADCSESC6aEdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC6b {
			if results, err := GetADCSESC6bEdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC9a {
			if results, err := GetADCSESC9aEdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC10a || edge.Kind == ad.ADCSESC10b {
			if results, err := GetADCSESC10EdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		}
		return nil
	})
}

func ADCSESC3Path1Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return optionalMemberOfOutboundPattern.
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.Kind(query.End(), ad.CertTemplate),
			query.And(
				query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
				query.Or(
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.And(
						query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
						query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					),
				),
			),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC3Path2Pattern(domainId graph.ID, enterpriseCAs, candidateTemplates cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return optionalMemberOfOutboundPattern.
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.KindIn(query.End(), ad.CertTemplate),
			query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
			query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
			query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(candidateTemplates)...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...))).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.Kind(query.End(), ad.NTAuthStore),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC3Path3Pattern() traversal.PatternContinuation {
	return optionalMemberOfOutboundPattern.
		Outbound(query.And(
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.KindIn(query.Relationship(), ad.Enroll),
		))
}

var optionalMemberOfOutboundPattern = traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
	query.Kind(query.Relationship(), ad.MemberOf),
	query.Kind(query.End(), ad.Group),
))

func ADCSESC6aPath1Pattern() traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.Equals(query.EndProperty(ad.IsUserSpecifiesSanEnabled.String()), true),
			query.KindIn(query.Relationship(), ad.Enroll),
		))
}

func ADCSESC6aPath2Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.Kind(query.End(), ad.CertTemplate),
			query.And(
				query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
				query.Equals(query.EndProperty(ad.NoSecurityExtension.String()), true),
				query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
				query.Or(
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.And(
						query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
						query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					),
				),
			),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC6aPath3Pattern(domainId graph.ID, enterpriseCAs, candidateTemplates cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.KindIn(query.End(), ad.CertTemplate),
			query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(candidateTemplates)...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...))).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.Kind(query.End(), ad.NTAuthStore),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC6aPath4Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.Enroll),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.CanAbuseWeakCertBinding),
			query.KindIn(query.End(), ad.Computer),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.DCFor, ad.TrustedBy),
			query.Equals(query.EndID(), domainId),
		))
}

func GetADCSESC6aEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		closureErr           error
		startNode            *graph.Node
		traversalInst        = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		lock                 = &sync.Mutex{}
		paths                = graph.PathSet{}
		certTemplateSegments = map[graph.ID][]*graph.PathSegment{}
		enterpriseCASegments = map[graph.ID][]*graph.PathSegment{}
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

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6aPath1Pattern().Do(func(terminal *graph.PathSegment) error {
			enterpriseCA := terminal.Search(func(nextSegment *graph.PathSegment) bool {
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

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6aPath2Pattern(edge.EndID, path1EnterpriseCAs).Do(func(terminal *graph.PathSegment) error {
			certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
			log.Infof("certtemplate %v", certTemplate)
			certTemplates.Add(certTemplate.ID.Uint32())
			lock.Unlock()

			return nil
		})}); err != nil {
		return nil, err
	}
	log.Infof("certtemplates %v", certTemplates.Slice())

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6aPath3Pattern(edge.EndID, path1EnterpriseCAs, certTemplates).Do(func(terminal *graph.PathSegment) error {
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

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6aPath4Pattern(edge.EndID, path1EnterpriseCAs).Do(func(terminal *graph.PathSegment) error {
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

	email, err := startNode.Properties.Get(common.Email.String()).String()
	if err != nil {
		log.Warnf("unable to access property %s for node with id %d: %v", common.Email.String(), startNode.ID, err)
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

		schemaVersion, err := certTemplate.Properties.Get(ad.SchemaVersion.String()).Float64()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SchemaVersion.String(), certTemplate.ID, err)
		}
		subjectAltRequireEmail, err := certTemplate.Properties.Get(ad.SubjectAltRequireEmail.String()).Bool()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SubjectAltRequireEmail.String(), certTemplate.ID, err)
		}
		subjectRequireEmail, err := certTemplate.Properties.Get(ad.SubjectRequireEmail.String()).Bool()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SubjectRequireEmail.String(), certTemplate.ID, err)
		}
		subjectAltRequireDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDNS.String()).Bool()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SubjectAltRequireDNS.String(), certTemplate.ID, err)
		}
		subjectAltRequireDomainDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDomainDNS.String()).Bool()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SubjectAltRequireDomainDNS.String(), certTemplate.ID, err)
		}

		for _, segment := range certTemplateSegments[graph.ID(value)] {
			if startNode.Kinds.ContainsOneOf(ad.User) {
				if subjectAltRequireDNS || subjectAltRequireDomainDNS {
					continue
				} else if email == "" && !((!subjectAltRequireEmail && !subjectRequireEmail) || schemaVersion == 1) {
					continue
				} else {
					log.Infof("Found ESC6a Path: %s", graph.FormatPathSegment(segment))
					paths.AddPath(segment.Path())
				}
			} else if startNode.Kinds.ContainsOneOf(ad.Computer) {
				if email == "" && !((!subjectAltRequireEmail && !subjectRequireEmail) || schemaVersion == 1) {
					continue
				} else {
					log.Infof("Found ESC6a Path: %s", graph.FormatPathSegment(segment))
					paths.AddPath(segment.Path())
				}
			} else {
				log.Infof("Found ESC6a Path: %s", graph.FormatPathSegment(segment))
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

func GetADCSESC3EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		startNode *graph.Node

		traversalInst           = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths                   = graph.PathSet{}
		path1CandidateSegments  = map[graph.ID][]*graph.PathSegment{}
		path2CandidateSegments  = map[graph.ID][]*graph.PathSegment{}
		lock                    = &sync.Mutex{}
		path1CertTemplates      = cardinality.NewBitmap32()
		path2CertTemplates      = cardinality.NewBitmap32()
		enterpriseCANodes       = cardinality.NewBitmap32()
		enterpriseCASegments    = map[graph.ID][]*graph.PathSegment{}
		path2CandidateTemplates = cardinality.NewBitmap32()
		enrollOnBehalfOfPaths   graph.PathSet
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

	//Start by fetching all EnterpriseCA nodes that our user has Enroll rights on via group membership or directly (P4/P5)
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC3Path3Pattern().Do(func(terminal *graph.PathSegment) error {
			enterpriseCANode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			enterpriseCASegments[enterpriseCANode.ID] = append(enterpriseCASegments[enterpriseCANode.ID], terminal)
			enterpriseCANodes.Add(enterpriseCANode.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	//Use the enterprise CA nodes we gathered to filter the first set of paths for P1
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC3Path1Pattern(edge.EndID, enterpriseCANodes).Do(func(terminal *graph.PathSegment) error {
			certTemplateNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			path1CandidateSegments[certTemplateNode.ID] = append(path1CandidateSegments[certTemplateNode.ID], terminal)
			path1CertTemplates.Add(certTemplateNode.ID.Uint32())
			lock.Unlock()

			return nil
		})}); err != nil {
		return nil, err
	}

	//Find all cert templates we have EnrollOnBehalfOf from our first group of templates to prefilter again
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if p, err := ops.FetchPathSet(tx.Relationships().Filter(
			query.And(
				query.InIDs(query.StartID(), cardinality.DuplexToGraphIDs(path1CertTemplates)...),
				query.KindIn(query.Relationship(), ad.EnrollOnBehalfOf),
				query.KindIn(query.End(), ad.CertTemplate)),
		)); err != nil {
			return err
		} else {
			enrollOnBehalfOfPaths = p
			return nil
		}
	}); err != nil {
		return nil, err
	}

	for _, path := range enrollOnBehalfOfPaths {
		path2CandidateTemplates.Add(path.Terminal().ID.Uint32())
	}

	//Use our enterprise ca + candidate templates as filters for the third query (P2)
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC3Path2Pattern(edge.EndID, enterpriseCANodes, path2CandidateTemplates).Do(func(terminal *graph.PathSegment) error {
			certTemplateNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			path2CandidateSegments[certTemplateNode.ID] = append(path2CandidateSegments[certTemplateNode.ID], terminal)
			path2CertTemplates.Add(certTemplateNode.ID.Uint32())
			lock.Unlock()

			return nil
		})}); err != nil {
		return nil, err
	}

	//EnrollOnBehalfOf is used to join P1 and P2, so we'll use it as the key
	for _, p3 := range enrollOnBehalfOfPaths {
		ct1 := p3.Root()
		ct2 := p3.Terminal()

		if !path1CertTemplates.Contains(ct1.ID.Uint32()) {
			continue
		}

		if !path2CertTemplates.Contains(ct2.ID.Uint32()) {
			continue
		}

		p1paths := path1CandidateSegments[ct1.ID]
		p2paths := path2CandidateSegments[ct2.ID]

		for _, p1 := range p1paths {
			eca1 := p1.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) && enterpriseCANodes.Contains(nextSegment.Node.ID.Uint32())
			})

			for _, p2 := range p2paths {
				eca2 := p2.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) && enterpriseCANodes.Contains(nextSegment.Node.ID.Uint32())
				})

				for _, p4 := range enterpriseCASegments[eca1.ID] {
					paths.AddPath(p4.Path())
				}

				for _, p5 := range enterpriseCASegments[eca2.ID] {
					paths.AddPath(p5.Path())
				}

				paths.AddPath(p3)
				paths.AddPath(p1.Path())
				paths.AddPath(p2.Path())

				if collected, err := eca2.Properties.Get(ad.EnrollmentAgentRestrictionsCollected.String()).Bool(); err != nil {
					log.Errorf("error getting enrollmentagentcollected for eca2 %d: %v", eca2.ID, err)
				} else if hasRestrictions, err := eca2.Properties.Get(ad.HasEnrollmentAgentRestrictions.String()).Bool(); err != nil {
					log.Errorf("error getting hasenrollmentagentrestrictions for ca %d: %v", eca2.ID, err)
				} else if collected && hasRestrictions {
					if p6, err := getDelegatedEnrollmentAgentPath(ctx, startNode, ct2, db); err != nil {
						log.Infof("Error getting p6 for composition: %v", err)
					} else {
						paths.AddPathSet(p6)
					}
				}
			}
		}
	}

	return paths, nil
}

func getDelegatedEnrollmentAgentPath(ctx context.Context, startNode, certTemplate2 *graph.Node, db graph.Database) (graph.PathSet, error) {
	var pathSet graph.PathSet

	return pathSet, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if paths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.InIDs(query.StartID(), startNode.ID),
			query.InIDs(query.EndID(), certTemplate2.ID),
			query.KindIn(query.Relationship(), ad.DelegatedEnrollmentAgent),
		))); err != nil {
			return err
		} else {
			pathSet = paths
			return nil
		}
	})
}

func ADCSESC1Path1Pattern(domainID graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.Kind(query.End(), ad.CertTemplate),
			query.Or(
				query.And(
					query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
					query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
					query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), true),
				),
				query.And(
					query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
					query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), true),
				),
			),
		)).
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

func ADCSESC1Path2Pattern(domainID graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.Enroll),
			query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.Kind(query.End(), ad.NTAuthStore),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.Equals(query.EndID(), domainID),
		))
}

func GetADCSESC1EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		startNode *graph.Node

		traversalInst      = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths              = graph.PathSet{}
		candidateSegments  = map[graph.ID][]*graph.PathSegment{}
		path1EnterpriseCAs = cardinality.NewBitmap32()
		path2EnterpriseCAs = cardinality.NewBitmap32()
		lock               = &sync.Mutex{}
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

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC1Path1Pattern(edge.EndID).Do(func(terminal *graph.PathSegment) error {
			// Find the CA and track it before stuffing this path into the candidates
			enterpriseCANode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			candidateSegments[enterpriseCANode.ID] = append(candidateSegments[enterpriseCANode.ID], terminal)
			path1EnterpriseCAs.Add(enterpriseCANode.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC1Path2Pattern(edge.EndID, path1EnterpriseCAs).Do(func(terminal *graph.PathSegment) error {
			// Find the CA and track it before stuffing this path into the candidates
			enterpriseCANode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			candidateSegments[enterpriseCANode.ID] = append(candidateSegments[enterpriseCANode.ID], terminal)
			path2EnterpriseCAs.Add(enterpriseCANode.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	// Intersect the CAs and take only those seen in both paths
	path1EnterpriseCAs.And(path2EnterpriseCAs)

	// Render paths from the segments
	path1EnterpriseCAs.Each(func(value uint32) bool {
		for _, segment := range candidateSegments[graph.ID(value)] {
			paths.AddPath(segment.Path())
		}

		return true
	})

	return paths, nil
}

func getGoldenCertEdgeComposition(tx graph.Transaction, edge *graph.Relationship) (graph.PathSet, error) {
	finalPaths := graph.NewPathSet()
	//Grab the start node (computer) as well as the target domain node
	if startNode, targetDomainNode, err := ops.FetchRelationshipNodes(tx, edge); err != nil {
		return finalPaths, err
	} else {
		//Find hosted enterprise CA
		if ecaPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.Equals(query.StartID(), startNode.ID),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.KindIn(query.Relationship(), ad.HostsCAService),
		))); err != nil {
			log.Errorf("error getting hostscaservice edge to enterprise ca for computer %d : %v", startNode.ID, err)
		} else {
			for _, ecaPath := range ecaPaths {
				eca := ecaPath.Terminal()
				if chainToRootCAPaths, err := FetchEnterpriseCAsCertChainPathToDomain(tx, eca, targetDomainNode); err != nil {
					log.Errorf("error getting eca %d path to domain %d: %v", eca.ID, targetDomainNode.ID, err)
				} else if chainToRootCAPaths.Len() == 0 {
					continue
				} else if trustedForAuthPaths, err := FetchEnterpriseCAsTrustedForAuthPathToDomain(tx, eca, targetDomainNode); err != nil {
					log.Errorf("error getting eca %d path to domain %d via trusted for auth: %v", eca.ID, targetDomainNode.ID, err)
				} else if trustedForAuthPaths.Len() == 0 {
					continue
				} else {
					finalPaths.AddPath(ecaPath)
					finalPaths.AddPathSet(chainToRootCAPaths)
					finalPaths.AddPathSet(trustedForAuthPaths)
				}
			}
		}

		return finalPaths, nil
	}
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

func adcsESC9APath3Pattern(caIDs []graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		Inbound(
			query.KindIn(query.Relationship(), ad.DCFor, ad.TrustedBy),
		).
		Inbound(query.And(
			query.Kind(query.Relationship(), ad.CanAbuseWeakCertBinding),
			query.InIDs(query.StartID(), caIDs...),
		))
}

func GetADCSESC9aEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH (n {objectid:'S-1-5-21-3933516454-2894985453-2515407000-500'})-[:ADCSESC9a]->(d:Domain {objectid:'S-1-5-21-3933516454-2894985453-2515407000'})
		OPTIONAL MATCH p1 = (n)-[:GenericAll|GenericWrite|Owns|WriteOwner|WriteDacl]->(m)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
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
		OPTIONAL MATCH p2 = (m)-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)
		OPTIONAL MATCH p3 = (ca)-[:CanAbuseWeakCertBinding|DCFor|TrustedBy*1..]->(d)
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
			Driver: adcsESC9APath3Pattern(p2canodes).Do(func(terminal *graph.PathSegment) error {
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

func certTemplateValidForUserVictim(certTemplate *graph.Node) bool {
	if subjectAltRequireDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDNS.String()).Bool(); err != nil {
		return false
	} else if subjectAltRequireDNS {
		return false
	} else if subjectAltRequireDomainDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDomainDNS.String()).Bool(); err != nil {
		return false
	} else if subjectAltRequireDomainDNS {
		return false
	} else {
		return true
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
