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
	"errors"
	"fmt"
	"log/slog"
	"slices"
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

func PostADCSESC3(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, groupExpansions impact.PathAggregator, eca2, domain *graph.Node, cache ADCSCache) error {
	results := cardinality.NewBitmap64()
	if domainsid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		log.Warnf(fmt.Sprintf("Error getting domain SID for domain %d: %v", domain.ID, err))
		return nil
	} else if publishedCertTemplates := cache.GetPublishedTemplateCache(eca2.ID); len(publishedCertTemplates) == 0 {
		return nil
	} else if collected, err := eca2.Properties.Get(ad.EnrollmentAgentRestrictionsCollected.String()).Bool(); err != nil {
		return fmt.Errorf("error getting enrollmentagentcollected for eca2 %d: %w", eca2.ID, err)
	} else {
		// Assuming no enrollement agent restrictions if not collected
		eARestrictions := false
		if collected {
			if hasRestrictions, err := eca2.Properties.Get(ad.HasEnrollmentAgentRestrictions.String()).Bool(); err != nil {
				return fmt.Errorf("error getting hasenrollmentagentrestrictions for ca %d: %w", eca2.ID, err)
			} else {
				eARestrictions = hasRestrictions
			}
		}

		for _, certTemplateTwo := range publishedCertTemplates {
			if !isEndCertTemplateValidESC3(certTemplateTwo) {
				continue
			}

			if inboundTemplates, err := ops.FetchStartNodes(tx.Relationships().Filterf(func() graph.Criteria {
				return query.And(
					query.Equals(query.EndID(), certTemplateTwo.ID),
					query.Kind(query.Relationship(), ad.EnrollOnBehalfOf),
					query.Kind(query.Start(), ad.CertTemplate),
				)
			})); err != nil {
				if !graph.IsErrNotFound(err) {
					slog.ErrorContext(ctx, fmt.Sprintf("Error getting target nodes for esc3 for node %d: %v", certTemplateTwo.ID, err))
				}
			} else {
				for _, certTemplateOne := range inboundTemplates {
					if !isStartCertTemplateValidESC3(certTemplateOne) {
						continue
					}

					var (
						ecaEnrollersTwo          = cache.GetEnterpriseCAEnrollers(eca2.ID)
						certTemplateEnrollersOne = cache.GetCertTemplateEnrollers(certTemplateOne.ID)
						certTemplateEnrollersTwo = cache.GetCertTemplateEnrollers(certTemplateTwo.ID)
					)

					if publishedECAs, err := FetchCertTemplateCAs(tx, certTemplateOne); err != nil {
						slog.ErrorContext(ctx, fmt.Sprintf("Error getting cas for cert template %d: %v", certTemplateOne.ID, err))
					} else if publishedECAs.Len() == 0 {
						continue
					} else if eARestrictions {
						if delegatedAgents, err := fetchFirstDegreeNodes(tx, certTemplateTwo, ad.DelegatedEnrollmentAgent); err != nil {
							slog.ErrorContext(ctx, fmt.Sprintf("Error getting delegated agents for cert template %d: %v", certTemplateTwo.ID, err))
						} else {
							for _, eca1 := range publishedECAs {
								tempResults := CalculateCrossProductNodeSets(tx,
									domainsid,
									groupExpansions,
									certTemplateEnrollersOne,
									certTemplateEnrollersTwo,
									cache.GetEnterpriseCAEnrollers(eca1.ID),
									ecaEnrollersTwo,
									delegatedAgents.Slice())

								// Add principals to result set unless it's a user and DNS is required
								if filteredResults, err := filterUserDNSResults(tx, tempResults, certTemplateOne); err != nil {
									slog.ErrorContext(ctx, fmt.Sprintf("Error filtering user dns results: %v", err))
								} else {
									results.Or(filteredResults)
								}
							}
						}
					} else {
						for _, eca1 := range publishedECAs {
							tempResults := CalculateCrossProductNodeSets(tx,
								domainsid,
								groupExpansions,
								certTemplateEnrollersOne,
								certTemplateEnrollersTwo,
								cache.GetEnterpriseCAEnrollers(eca1.ID),
								ecaEnrollersTwo)

							if filteredResults, err := filterUserDNSResults(tx, tempResults, certTemplateOne); err != nil {
								slog.ErrorContext(ctx, fmt.Sprintf("Error filtering user dns results: %v", err))
							} else {
								results.Or(filteredResults)
							}
						}
					}
				}
			}
		}
	}

	results.Each(func(value uint64) bool {
		channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
			FromID: graph.ID(value),
			ToID:   domain.ID,
			Kind:   ad.ADCSESC3,
		})
		return true
	})

	return nil
}

func PostEnrollOnBehalfOf(domains, enterpriseCertAuthorities, certTemplates []*graph.Node, cache ADCSCache, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) error {
	versionOneTemplates := make([]*graph.Node, 0)
	versionTwoTemplates := make([]*graph.Node, 0)
	for _, node := range certTemplates {
		if version, err := node.Properties.Get(ad.SchemaVersion.String()).Float64(); errors.Is(err, graph.ErrPropertyNotFound) {
			log.Warnf(fmt.Sprintf("Did not get schema version for cert template %d: %v", node.ID, err))
		} else if err != nil {
			slog.Error(fmt.Sprintf("Error getting schema version for cert template %d: %v", node.ID, err))
		} else if version == 1 {
			versionOneTemplates = append(versionOneTemplates, node)
		} else if version >= 2 {
			versionTwoTemplates = append(versionTwoTemplates, node)
		} else {
			log.Warnf(fmt.Sprintf("Got cert template %d with an invalid version %d", node.ID, version))
		}
	}

	for _, domain := range domains {
		innerDomain := domain

		for _, enterpriseCA := range enterpriseCertAuthorities {
			innerEnterpriseCA := enterpriseCA

			if cache.DoesCAChainProperlyToDomain(innerEnterpriseCA, innerDomain) {
				if publishedCertTemplates := cache.GetPublishedTemplateCache(enterpriseCA.ID); len(publishedCertTemplates) == 0 {
					return nil
				} else {
					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if results, err := EnrollOnBehalfOfVersionTwo(tx, versionTwoTemplates, publishedCertTemplates, innerDomain); err != nil {
							return err
						} else {
							for _, result := range results {
								if !channels.Submit(ctx, outC, result) {
									return nil
								}
							}

							return nil
						}
					})

					operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						if results, err := EnrollOnBehalfOfVersionOne(tx, versionOneTemplates, publishedCertTemplates, innerDomain); err != nil {
							return err
						} else {
							for _, result := range results {
								if !channels.Submit(ctx, outC, result) {
									return nil
								}
							}

							return nil
						}
					})
				}
			}
		}
	}

	return nil
}

func EnrollOnBehalfOfVersionTwo(tx graph.Transaction, versionTwoCertTemplates, publishedTemplates []*graph.Node, domainNode *graph.Node) ([]analysis.CreatePostRelationshipJob, error) {
	results := make([]analysis.CreatePostRelationshipJob, 0)
	for _, certTemplateOne := range publishedTemplates {
		if hasBadEku, err := certTemplateHasEku(certTemplateOne, EkuAnyPurpose); errors.Is(err, graph.ErrPropertyNotFound) {
			log.Warnf(fmt.Sprintf("Did not get EffectiveEKUs for cert template %d: %v", certTemplateOne.ID, err))
		} else if err != nil {
			slog.Error(fmt.Sprintf("Error getting EffectiveEKUs for cert template %d: %v", certTemplateOne.ID, err))
		} else if hasBadEku {
			continue
		} else if hasEku, err := certTemplateHasEku(certTemplateOne, EkuCertRequestAgent); errors.Is(err, graph.ErrPropertyNotFound) {
			log.Warnf(fmt.Sprintf("Did not get EffectiveEKUs for cert template %d: %v", certTemplateOne.ID, err))
		} else if err != nil {
			slog.Error(fmt.Sprintf("Error getting EffectiveEKUs for cert template %d: %v", certTemplateOne.ID, err))
		} else if !hasEku {
			continue
		} else {
			for _, certTemplateTwo := range versionTwoCertTemplates {
				if certTemplateOne.ID == certTemplateTwo.ID {
					continue
				} else if authorizedSignatures, err := certTemplateTwo.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
					slog.Error(fmt.Sprintf("Error getting authorized signatures for cert template %d: %v", certTemplateTwo.ID, err))
				} else if authorizedSignatures < 1 {
					continue
				} else if applicationPolicies, err := certTemplateTwo.Properties.Get(ad.ApplicationPolicies.String()).StringSlice(); err != nil {
					slog.Error(fmt.Sprintf("Error getting application policies for cert template %d: %v", certTemplateTwo.ID, err))
				} else if !slices.Contains(applicationPolicies, EkuCertRequestAgent) {
					continue
				} else if isLinked, err := DoesCertTemplateLinkToDomain(tx, certTemplateTwo, domainNode); err != nil {
					slog.Error(fmt.Sprintf("Error fetch paths from cert template %d to domain: %v", certTemplateTwo.ID, err))
				} else if !isLinked {
					continue
				} else {
					results = append(results, analysis.CreatePostRelationshipJob{
						FromID: certTemplateOne.ID,
						ToID:   certTemplateTwo.ID,
						Kind:   ad.EnrollOnBehalfOf,
					})
				}
			}
		}
	}

	return results, nil
}

func certTemplateHasEku(certTemplate *graph.Node, targetEkus ...string) (bool, error) {
	if ekus, err := certTemplate.Properties.Get(ad.EffectiveEKUs.String()).StringSlice(); err != nil {
		return false, err
	} else {
		for _, eku := range ekus {
			for _, targetEku := range targetEkus {
				if eku == targetEku {
					return true, nil
				}
			}
		}

		return false, nil
	}
}

func EnrollOnBehalfOfVersionOne(tx graph.Transaction, versionOneCertTemplates []*graph.Node, publishedTemplates []*graph.Node, domainNode *graph.Node) ([]analysis.CreatePostRelationshipJob, error) {
	results := make([]analysis.CreatePostRelationshipJob, 0)

	for _, certTemplateOne := range publishedTemplates {
		//prefilter as much as we can first
		if hasEku, err := certTemplateHasEkuOrAll(certTemplateOne, EkuCertRequestAgent, EkuAnyPurpose); errors.Is(err, graph.ErrPropertyNotFound) {
			log.Warnf(fmt.Sprintf("Error checking ekus for certtemplate %d: %v", certTemplateOne.ID, err))
		} else if err != nil {
			slog.Error(fmt.Sprintf("Error checking ekus for certtemplate %d: %v", certTemplateOne.ID, err))
		} else if !hasEku {
			continue
		} else {
			for _, certTemplateTwo := range versionOneCertTemplates {
				if hasPath, err := DoesCertTemplateLinkToDomain(tx, certTemplateTwo, domainNode); err != nil {
					slog.Error(fmt.Sprintf("Error getting domain node for certtemplate %d: %v", certTemplateTwo.ID, err))
				} else if !hasPath {
					continue
				} else {
					results = append(results, analysis.CreatePostRelationshipJob{
						FromID: certTemplateOne.ID,
						ToID:   certTemplateTwo.ID,
						Kind:   ad.EnrollOnBehalfOf,
					})
				}
			}
		}
	}

	return results, nil
}

func isStartCertTemplateValidESC3(template *graph.Node) bool {
	if reqManagerApproval, err := template.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		slog.Error(fmt.Sprintf("Error getting reqmanagerapproval for certtemplate %d: %v", template.ID, err))
	} else if reqManagerApproval {
		return false
	} else if schemaVersion, err := template.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		slog.Error(fmt.Sprintf("Error getting schemaversion for certtemplate %d: %v", template.ID, err))
	} else if schemaVersion == 1 {
		return true
	} else if schemaVersion > 1 {
		if authorizedSignatures, err := template.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
			slog.Error(fmt.Sprintf("Error getting authorizedsignatures for certtemplate %d: %v", template.ID, err))
		} else if authorizedSignatures > 0 {
			return false
		} else {
			return true
		}
	}

	return false
}

func isEndCertTemplateValidESC3(template *graph.Node) bool {
	if authEnabled, err := template.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		log.Warnf(fmt.Sprintf("Did not getting authenabled for cert template %d: %v", template.ID, err))
		return false
	} else if err != nil {
		slog.Error(fmt.Sprintf("Error getting authenabled for cert template %d: %v", template.ID, err))
		return false
	} else if !authEnabled {
		return false
	} else if reqManagerApproval, err := template.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		log.Warnf(fmt.Sprintf("Did not getting reqManagerApproval for cert template %d: %v", template.ID, err))
		return false
	} else if err != nil {
		slog.Error(fmt.Sprintf("Error getting reqManagerApproval for cert template %d: %v", template.ID, err))
		return false
	} else if reqManagerApproval {
		return false
	} else {
		return true
	}
}

func certTemplateHasEkuOrAll(certTemplate *graph.Node, targetEkus ...string) (bool, error) {
	if ekus, err := certTemplate.Properties.Get(ad.EffectiveEKUs.String()).StringSlice(); err != nil {
		return false, err
	} else if len(ekus) == 0 {
		return true, nil
	} else {
		for _, eku := range ekus {
			for _, targetEku := range targetEkus {
				if eku == targetEku {
					return true, nil
				}
			}
		}

		return false, nil
	}
}

func GetADCSESC3EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH p1 = (x)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct1:CertTemplate)-[:PublishedTo]->(eca1:EnterpriseCA)-[:TrustedForNTAuth]->(:NTAuthStore)-[:NTAuthStoreFor]->(d)
		WHERE x.objectid = "S-1-5-21-83094068-830424655-2031507174-500"
		AND d.objectid = "S-1-5-21-83094068-830424655-2031507174"
		AND ct1.requiresmanagerapproval = false
		AND (ct1.schemaversion = 1 OR ct1.authorizedsignatures = 0)
		AND (
			x:Group
			OR x:Computer
			OR (
			x:User
			AND ct1.subjectaltrequiredns = false
			AND ct1.subjectaltrequiredomaindns = false
			)
		)

		MATCH p2 = (x)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct2:CertTemplate)-[:PublishedTo]->(eca2:EnterpriseCA)-[:TrustedForNTAuth]->(:NTAuthStore)-[:NTAuthStoreFor]->(d)
		WHERE ct2.authenticationenabled = true
		AND ct2.requiresmanagerapproval = false

		MATCH p3 = (ct1)-[:EnrollOnBehalfOf]->(ct2)

		MATCH p4 = (x)-[:MemberOf*0..]->()-[:Enroll]->(eca1)

		MATCH p5 = (x)-[:MemberOf*0..]->()-[:Enroll]->(eca2)

		MATCH p6 = (eca1)-[:IssuedSignedBy|EnterpriseCAFor*1..]->(:RootCA)-[:RootCAFor]->(d)
		MATCH p7 = (eca2)-[:IssuedSignedBy|EnterpriseCAFor*1..]->(:RootCA)-[:RootCAFor]->(d)

		OPTIONAL MATCH p8 = (x)-[:MemberOf*0..]->()-[:DelegatedEnrollmentAgent]->(ct2)

		WITH *
		WHERE (
			NOT eca2.hasenrollmentagentrestrictions = True
			OR p8 IS NOT NULL
		)

		RETURN p1,p2,p3,p4,p5,p6,p7,p8
	*/
	var (
		startNode  *graph.Node
		endNode    *graph.Node
		startNodes = graph.NodeSet{}

		traversalInst            = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths                    = graph.PathSet{}
		path1CandidateSegments   = map[graph.ID][]*graph.PathSegment{}
		path2CandidateSegments   = map[graph.ID][]*graph.PathSegment{}
		path6_7CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path8CandidateSegments   = map[graph.ID][]*graph.PathSegment{}
		lock                     = &sync.Mutex{}
		path1CertTemplates       = cardinality.NewBitmap64()
		path2CertTemplates       = cardinality.NewBitmap64()
		enterpriseCANodes        = cardinality.NewBitmap64()
		enterpriseCASegments     = map[graph.ID][]*graph.PathSegment{}
		path2CandidateTemplates  = cardinality.NewBitmap64()
		enrollOnBehalfOfPaths    graph.PathSet
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

	// Add startnode, Auth. Users, and Everyone to start nodes
	if domainsid, err := endNode.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		log.Warnf(fmt.Sprintf("Error getting domain SID for domain %d: %v", endNode.ID, err))
		return nil, err
	} else if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if nodeSet, err := FetchAuthUsersAndEveryoneGroups(tx, domainsid); err != nil {
			return err
		} else {
			startNodes.AddSet(nodeSet)
			return nil
		}
	}); err != nil {
		return nil, err
	}
	startNodes.Add(startNode)

	//Start by fetching all EnterpriseCA nodes that our user has Enroll rights on via group membership or directly (P4/P5)
	for _, n := range startNodes.Slice() {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: n,
			Driver: ADCSESC3Path3Pattern().Do(func(terminal *graph.PathSegment) error {
				enterpriseCANode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
				})

				lock.Lock()
				enterpriseCASegments[enterpriseCANode.ID] = append(enterpriseCASegments[enterpriseCANode.ID], terminal)
				enterpriseCANodes.Add(enterpriseCANode.ID.Uint64())
				lock.Unlock()

				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	//Use the enterprise CA nodes we gathered to filter the first set of paths for P1
	for _, n := range startNodes.Slice() {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: n,
			Driver: ADCSESC3Path1Pattern(edge.EndID, enterpriseCANodes).Do(func(terminal *graph.PathSegment) error {
				certTemplateNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
				})

				lock.Lock()
				path1CandidateSegments[certTemplateNode.ID] = append(path1CandidateSegments[certTemplateNode.ID], terminal)

				// Check that CT is valid for user start nodes
				userStartNode := startNode.Kinds.ContainsOneOf(ad.User)
				if !userStartNode || certTemplateValidForUserVictim(certTemplateNode) {
					path1CertTemplates.Add(certTemplateNode.ID.Uint64())
				}
				lock.Unlock()

				return nil
			})}); err != nil {
			return nil, err
		}
	}

	//Find all cert templates we have EnrollOnBehalfOf from our first group of templates to prefilter again
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if p, err := ops.FetchPathSet(tx.Relationships().Filter(
			query.And(
				query.InIDs(query.StartID(), graph.DuplexToGraphIDs(path1CertTemplates)...),
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
		path2CandidateTemplates.Add(path.Terminal().ID.Uint64())
	}

	//Use our enterprise ca + candidate templates as filters for the third query (P2)
	for _, n := range startNodes.Slice() {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: n,
			Driver: ADCSESC3Path2Pattern(edge.EndID, enterpriseCANodes, path2CandidateTemplates).Do(func(terminal *graph.PathSegment) error {
				certTemplateNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
				})

				lock.Lock()
				path2CandidateSegments[certTemplateNode.ID] = append(path2CandidateSegments[certTemplateNode.ID], terminal)
				path2CertTemplates.Add(certTemplateNode.ID.Uint64())
				lock.Unlock()

				return nil
			})}); err != nil {
			return nil, err
		}
	}

	//Manifest P6/P7 keyed to enterprise ca nodes
	for ecaID := range enterpriseCASegments {
		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			if ecaNode, err := ops.FetchNode(tx, ecaID); err != nil {
				return err
			} else {
				if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
					Root: ecaNode,
					Driver: ADCSESC3Path6_7Pattern(edge.EndID).Do(func(terminal *graph.PathSegment) error {
						eca := terminal.Path().Root()
						if eca.ID == ecaID {
							lock.Lock()
							path6_7CandidateSegments[ecaID] = append(path6_7CandidateSegments[ecaID], terminal)
							lock.Unlock()
						}
						return nil
					}),
				}); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return nil, err
		}
	}

	//Manifest p8 keyed to certificate template nodes
	for _, n := range startNodes.Slice() {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: n,
			Driver: ADCSESC3Path8Pattern(path2CandidateTemplates).Do(func(terminal *graph.PathSegment) error {
				certTemplateNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
				})

				lock.Lock()
				path8CandidateSegments[certTemplateNode.ID] = append(path8CandidateSegments[certTemplateNode.ID], terminal)
				lock.Unlock()
				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	//EnrollOnBehalfOf is used to join P1 and P2, so we'll use it as the key
	for _, p3 := range enrollOnBehalfOfPaths {

		ct1 := p3.Root()
		ct2 := p3.Terminal()

		if !path1CertTemplates.Contains(ct1.ID.Uint64()) {
			continue
		}

		if !path2CertTemplates.Contains(ct2.ID.Uint64()) {
			continue
		}

		p1paths := path1CandidateSegments[ct1.ID]
		p2paths := path2CandidateSegments[ct2.ID]

		for _, p1 := range p1paths {
			eca1 := p1.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) && enterpriseCANodes.Contains(nextSegment.Node.ID.Uint64())
			})

			for _, p2 := range p2paths {
				eca2 := p2.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) && enterpriseCANodes.Contains(nextSegment.Node.ID.Uint64())
				})

				// Verify P6 and P7 paths exists
				p6segments, ok := path6_7CandidateSegments[eca1.ID]
				if !ok {
					continue
				}
				p7segments, ok := path6_7CandidateSegments[eca2.ID]
				if !ok {
					continue
				}

				if collected, err := eca2.Properties.Get(ad.EnrollmentAgentRestrictionsCollected.String()).Bool(); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error getting enrollmentagentcollected for eca2 %d: %v", eca2.ID, err))
				} else if collected {
					if hasRestrictions, err := eca2.Properties.Get(ad.HasEnrollmentAgentRestrictions.String()).Bool(); err != nil {
						slog.ErrorContext(ctx, fmt.Sprintf("Error getting hasenrollmentagentrestrictions for ca %d: %v", eca2.ID, err))
					} else if hasRestrictions {

						// Verify p8 path exist
						p8segments, ok := path8CandidateSegments[ct2.ID]
						if !ok {
							continue
						}

						for _, p8 := range p8segments {
							paths.AddPath(p8.Path())
						}
					}
				}

				for _, p4 := range enterpriseCASegments[eca1.ID] {
					paths.AddPath(p4.Path())
				}

				for _, p5 := range enterpriseCASegments[eca2.ID] {
					paths.AddPath(p5.Path())
				}

				for _, p6 := range p6segments {
					paths.AddPath(p6.Path())
				}

				for _, p7 := range p7segments {
					paths.AddPath(p7.Path())
				}

				paths.AddPath(p3)
				paths.AddPath(p1.Path())
				paths.AddPath(p2.Path())
			}
		}
	}

	return paths, nil
}

func ADCSESC3Path1Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint64]) traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
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
			query.InIDs(query.End(), graph.DuplexToGraphIDs(enterpriseCAs)...),
			query.Kind(query.End(), ad.EnterpriseCA),
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

func ADCSESC3Path2Pattern(domainId graph.ID, enterpriseCAs, candidateTemplates cardinality.Duplex[uint64]) traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.KindIn(query.End(), ad.CertTemplate),
			query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
			query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
			query.InIDs(query.EndID(), graph.DuplexToGraphIDs(candidateTemplates)...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.InIDs(query.End(), graph.DuplexToGraphIDs(enterpriseCAs)...))).
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
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.KindIn(query.Relationship(), ad.Enroll),
		))
}

func ADCSESC3Path6_7Pattern(domainId graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
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
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC3Path8Pattern(candidateTemplates cardinality.Duplex[uint64]) traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.DelegatedEnrollmentAgent),
			query.InIDs(query.EndID(), graph.DuplexToGraphIDs(candidateTemplates)...),
		))
}
