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
	"log/slog"
	"slices"
	"sync"

	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/traversal"
	"github.com/specterops/dawgs/util/channels"
)

func PostADCSESC3(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob, localGroupData *LocalGroupData, certChains *EnterpriseCAChainedDomains, cache *ADCSCache) error {
	var (
		resultsByDomain = map[graph.ID]cardinality.Duplex[uint64]{}
		eca2ID          = certChains.EnterpriseCA.ID
	)
	for _, domain := range certChains.Domains.Slice() {
		resultsByDomain[graph.ID(domain)] = cardinality.NewBitmap64()
	}

	if publishedCertTemplates := cache.GetPublishedTemplateCache(eca2ID); len(publishedCertTemplates) == 0 {
		return nil
	} else {
		hasEnrollmentAgentRestrictions := enterpriseCAHasEnrollmentAgentRestrictions(certChains.EnterpriseCA)

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
					slog.ErrorContext(
						ctx,
						"Error getting target nodes for esc3 for node",
						slog.Uint64("cert_template_two_id", uint64(certTemplateTwo.ID)),
						attr.Error(err),
					)
				}
			} else {
				for _, certTemplateOne := range inboundTemplates {
					if !isStartCertTemplateValidESC3(certTemplateOne) {
						continue
					}

					var (
						ecaEnrollersTwo          = cache.GetEnterpriseCAEnrollers(eca2ID)
						certTemplateEnrollersOne = cache.GetCertTemplateEnrollers(certTemplateOne.ID)
						certTemplateEnrollersTwo = cache.GetCertTemplateEnrollers(certTemplateTwo.ID)
					)

					if publishedECAs, err := FetchCertTemplateCAs(tx, certTemplateOne); err != nil {
						slog.ErrorContext(
							ctx,
							"Error getting cas for cert template",
							slog.Uint64("cert_template_one_id", uint64(certTemplateOne.ID)),
							attr.Error(err),
						)
					} else if publishedECAs.Len() == 0 {
						continue
					} else {
						var delegatedAgentSet CachedPrincipalSet
						if hasEnrollmentAgentRestrictions {
							if delegatedAgents, err := fetchFirstDegreeNodes(tx, certTemplateTwo, ad.DelegatedEnrollmentAgent); err != nil {
								slog.ErrorContext(
									ctx,
									"Error getting delegated agents for cert template",
									slog.Uint64("cert_template_two_id", uint64(certTemplateTwo.ID)),
									attr.Error(err),
								)
								continue
							} else if delegatedAgents.Len() == 0 {
								continue
							} else {
								delegatedAgentSet = NewCachedPrincipalSet(delegatedAgents.Slice())
							}
						}

						for _, eca1 := range publishedECAs {
							if !cache.enterpriseCAHasQualifyingHost(eca1.ID) {
								continue
							}

							principalSets := []CachedPrincipalSet{
								certTemplateEnrollersOne,
								certTemplateEnrollersTwo,
								cache.GetEnterpriseCAEnrollers(eca1.ID),
								ecaEnrollersTwo,
							}
							if hasEnrollmentAgentRestrictions {
								principalSets = append(principalSets, delegatedAgentSet)
							}

							tempResults := CalculateCrossProductNodeSets(localGroupData, principalSets...)

							if filteredResults, err := filterUserDNSResults(tx, tempResults, certTemplateOne); err != nil {
								slog.ErrorContext(
									ctx,
									"Error filtering user dns results",
									attr.Error(err),
								)
							} else {
								for domainID, results := range resultsByDomain {
									if cache.enterpriseCAHasChainedDomain(eca1.ID, domainID) {
										results.Or(filteredResults)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	for domainID, results := range resultsByDomain {
		results.Each(func(source uint64) bool {
			channels.Submit(ctx, outC, post.EnsureRelationshipJob{
				FromID: graph.ID(source),
				ToID:   domainID,
				Kind:   ad.ADCSESC3,
			})
			return true
		})
	}

	return nil
}

func PostEnrollOnBehalfOf(cache *ADCSCache, operation post.StatTrackedOperation[post.EnsureRelationshipJob]) error {
	hostedChainedDomains := cache.GetECAHostedChainedDomains()

	operation.Operation.SubmitReader(func(ctx context.Context, _ graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
		submittedTargetsBySource := make(map[graph.ID]map[graph.ID]struct{})
		type schemaSplit struct {
			versionOneTemplates []*graph.Node
			versionTwoTemplates []*graph.Node
		}
		splitsByTargetCA := make(map[graph.ID]schemaSplit)
		splitForTargetCA := func(enterpriseCAID graph.ID) schemaSplit {
			if split, ok := splitsByTargetCA[enterpriseCAID]; ok {
				return split
			}
			versionOneTemplates, versionTwoTemplates := splitCertTemplatesBySchemaVersion(cache.GetPublishedTemplateCache(enterpriseCAID))
			split := schemaSplit{versionOneTemplates: versionOneTemplates, versionTwoTemplates: versionTwoTemplates}
			splitsByTargetCA[enterpriseCAID] = split
			return split
		}

		submitRelationship := func(result post.EnsureRelationshipJob) bool {
			if targets, ok := submittedTargetsBySource[result.FromID]; ok {
				if _, ok := targets[result.ToID]; ok {
					return true
				}
				targets[result.ToID] = struct{}{}
			} else {
				submittedTargetsBySource[result.FromID] = map[graph.ID]struct{}{result.ToID: {}}
			}

			return channels.Submit(ctx, outC, result)
		}

		for _, enrollmentAgentChains := range hostedChainedDomains {
			enrollmentAgentTemplates := cache.GetPublishedTemplateCache(enrollmentAgentChains.EnterpriseCA.ID)
			if len(enrollmentAgentTemplates) == 0 {
				continue
			}

			for _, targetChains := range hostedChainedDomains {
				if !enterpriseCAChainsShareDomain(enrollmentAgentChains, targetChains) {
					continue
				}

				split := splitForTargetCA(targetChains.EnterpriseCA.ID)

				for _, result := range EnrollOnBehalfOfVersionTwo(split.versionTwoTemplates, enrollmentAgentTemplates) {
					if !submitRelationship(result) {
						return nil
					}
				}

				for _, result := range EnrollOnBehalfOfVersionOne(split.versionOneTemplates, enrollmentAgentTemplates) {
					if !submitRelationship(result) {
						return nil
					}
				}
			}
		}

		return nil
	})

	return nil
}

func enterpriseCAChainsShareDomain(first, second *EnterpriseCAChainedDomains) bool {
	for _, domainID := range first.Domains.Slice() {
		if second.Domains.Contains(domainID) {
			return true
		}
	}

	return false
}

func splitCertTemplatesBySchemaVersion(certTemplates []*graph.Node) ([]*graph.Node, []*graph.Node) {
	var (
		versionOneTemplates = make([]*graph.Node, 0)
		versionTwoTemplates = make([]*graph.Node, 0)
	)

	for _, certTemplate := range certTemplates {
		if version, err := certTemplate.Properties.Get(ad.SchemaVersion.String()).Float64(); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.Warn(
				"Did not get schema version for cert template",
				slog.Uint64("cert_template_id", uint64(certTemplate.ID)),
				attr.Error(err),
			)
		} else if err != nil {
			slog.Error(
				"Error getting schema version for cert template",
				slog.Uint64("cert_template_id", uint64(certTemplate.ID)),
				attr.Error(err),
			)
		} else if version == 1 {
			versionOneTemplates = append(versionOneTemplates, certTemplate)
		} else if version >= 2 {
			versionTwoTemplates = append(versionTwoTemplates, certTemplate)
		} else {
			slog.Warn(
				"Got cert template with an invalid version",
				slog.Uint64("cert_template_id", uint64(certTemplate.ID)),
				slog.Float64("version", version),
			)
		}
	}

	return versionOneTemplates, versionTwoTemplates
}

// EnrollOnBehalfOfVersionOne creates the schema-level compatibility relationships
// between enrollment agent and schema version 1 certificate templates. Callers
// must restrict both templates to publishers that share a qualifying domain.
func EnrollOnBehalfOfVersionOne(versionOneCertTemplates, enrollmentAgentCertTemplates []*graph.Node) []post.EnsureRelationshipJob {
	results := make([]post.EnsureRelationshipJob, 0)

	for _, certTemplateOne := range enrollmentAgentCertTemplates {
		//prefilter as much as we can first
		if hasEku, err := certTemplateHasEkuOrAll(certTemplateOne, EkuCertRequestAgent, EkuAnyPurpose); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.Warn(
				"Error checking ekus for certtemplate",
				slog.Uint64("cert_template_id", uint64(certTemplateOne.ID)),
				attr.Error(err),
			)
		} else if err != nil {
			slog.Error(
				"Error checking ekus for certtemplate",
				slog.Uint64("cert_template_id", uint64(certTemplateOne.ID)),
				attr.Error(err),
			)
		} else if !hasEku {
			continue
		} else {
			for _, certTemplateTwo := range versionOneCertTemplates {
				results = append(results, post.EnsureRelationshipJob{
					FromID: certTemplateOne.ID,
					ToID:   certTemplateTwo.ID,
					Kind:   ad.EnrollOnBehalfOf,
				})
			}
		}
	}

	return results
}

// EnrollOnBehalfOfVersionTwo creates the schema-level compatibility relationships
// between enrollment agent and schema version 2+ certificate templates. Callers
// must restrict both templates to publishers that share a qualifying domain.
func EnrollOnBehalfOfVersionTwo(versionTwoCertTemplates, enrollmentAgentCertTemplates []*graph.Node) []post.EnsureRelationshipJob {
	results := make([]post.EnsureRelationshipJob, 0)
	for _, certTemplateOne := range enrollmentAgentCertTemplates {
		if hasBadEku, err := certTemplateHasEku(certTemplateOne, EkuAnyPurpose); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.Warn(
				"Did not get EffectiveEKUs for cert template",
				slog.Uint64("cert_template_id", uint64(certTemplateOne.ID)),
				attr.Error(err),
			)
		} else if err != nil {
			slog.Error(
				"Error getting EffectiveEKUs for cert template",
				slog.Uint64("cert_template_id", uint64(certTemplateOne.ID)),
				attr.Error(err),
			)
		} else if hasBadEku {
			continue
		} else if hasEku, err := certTemplateHasEku(certTemplateOne, EkuCertRequestAgent); errors.Is(err, graph.ErrPropertyNotFound) {
			slog.Warn(
				"Did not get EffectiveEKUs for cert template",
				slog.Uint64("cert_template_id", uint64(certTemplateOne.ID)),
				attr.Error(err),
			)
		} else if err != nil {
			slog.Error(
				"Error getting EffectiveEKUs for cert template",
				slog.Uint64("cert_template_id", uint64(certTemplateOne.ID)),
				attr.Error(err),
			)
		} else if !hasEku {
			continue
		} else {
			for _, certTemplateTwo := range versionTwoCertTemplates {
				if certTemplateOne.ID == certTemplateTwo.ID {
					continue
				} else if authorizedSignatures, err := certTemplateTwo.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
					slog.Error(
						"Error getting authorized signatures for cert template",
						slog.Uint64("cert_template_id", uint64(certTemplateTwo.ID)),
						attr.Error(err),
					)
				} else if authorizedSignatures != 1 {
					continue
				} else if applicationPolicies, err := certTemplateTwo.Properties.Get(ad.ApplicationPolicies.String()).StringSlice(); err != nil {
					slog.Error(
						"Error getting application policies for cert template",
						slog.Uint64("cert_template_id", uint64(certTemplateTwo.ID)),
						attr.Error(err),
					)
				} else if !slices.Contains(applicationPolicies, EkuCertRequestAgent) {
					continue
				} else {
					results = append(results, post.EnsureRelationshipJob{
						FromID: certTemplateOne.ID,
						ToID:   certTemplateTwo.ID,
						Kind:   ad.EnrollOnBehalfOf,
					})
				}
			}
		}
	}

	return results
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

func isStartCertTemplateValidESC3(template *graph.Node) bool {
	if reqManagerApproval, err := template.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		if errors.Is(err, graph.ErrPropertyNotFound) {
			slog.Warn(
				"Node missing reqmanagerapproval for certtemplate",
				slog.Int("node_id", int(template.ID)),
				attr.Error(err),
			)
		} else {
			slog.Error(
				"Error getting reqmanagerapproval for certtemplate",
				slog.Int("node_id", int(template.ID)),
				attr.Error(err),
			)
		}
		return false
	} else if reqManagerApproval {
		return false
	}

	schemaVersion, err := template.Properties.Get(ad.SchemaVersion.String()).Float64()
	if err != nil {
		if errors.Is(err, graph.ErrPropertyNotFound) {
			slog.Warn(
				"Node missing schemaversion for certtemplate",
				slog.Int("node_id", int(template.ID)),
				attr.Error(err),
			)
		} else {
			slog.Error(
				"Error getting schemaversion for certtemplate",
				slog.Int("node_id", int(template.ID)),
				attr.Error(err),
			)
		}
		return false
	} else if schemaVersion == 1 {
		return true
	} else if schemaVersion <= 1 {
		slog.Warn(
			"Got cert template with an invalid schema version",
			slog.Int("node_id", int(template.ID)),
			slog.Float64("schema_version", schemaVersion),
		)
		return false
	}

	if authorizedSignatures, err := template.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		if errors.Is(err, graph.ErrPropertyNotFound) {
			slog.Warn(
				"Node missing authorizedsignatures for certtemplate",
				slog.Int("node_id", int(template.ID)),
				attr.Error(err),
			)
		} else {
			slog.Error(
				"Error getting authorizedsignatures for certtemplate",
				slog.Int("node_id", int(template.ID)),
				attr.Error(err),
			)
		}
		return false
	} else {
		return authorizedSignatures == 0
	}
}

func isEndCertTemplateValidESC3(template *graph.Node) bool {
	if authEnabled, err := template.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		slog.Warn(
			"Could not get property authenabled for cert template",
			slog.Uint64("cert_template_id", uint64(template.ID)),
			attr.Error(err),
		)
		return false
	} else if err != nil {
		slog.Error(
			"Error getting authenabled for cert template",
			slog.Uint64("cert_template_id", uint64(template.ID)),
			attr.Error(err),
		)
		return false
	} else if !authEnabled {
		return false
	} else if reqManagerApproval, err := template.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		slog.Warn(
			"Could not get property reqmanagerapproval for cert template",
			slog.Uint64("cert_template_id", uint64(template.ID)),
			attr.Error(err),
		)
		return false
	} else if err != nil {
		slog.Error(
			"Error getting reqManagerApproval for cert template",
			slog.Uint64("cert_template_id", uint64(template.ID)),
			attr.Error(err),
		)
		return false
	} else if reqManagerApproval {
		return false
	}

	return true
}

func GetADCSESC3EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	// The query represents the composed graph. The shared host-eligibility helper
	// additionally validates the host forest when the Enterprise CA forest resolves.
	/*
		MATCH (n {objectid: '<principal SID>'})-[:ADCSESC3]->(d:Domain {objectid: '<domain SID>'})

		MATCH p1 = (p1Principal)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct1:CertTemplate)-[:PublishedTo]->(eca1:EnterpriseCA)-[:TrustedForNTAuth]->(:NTAuthStore)-[:NTAuthStoreFor]->(d)
		WHERE (
			p1Principal.objectid = n.objectid
			OR p1Principal.objectid ENDS WITH '-S-1-5-11'
			OR p1Principal.objectid ENDS WITH '-S-1-1-0'
		)
		AND ct1.requiresmanagerapproval = false
		AND (
			ct1.schemaversion = 1
			OR (ct1.schemaversion > 1 AND ct1.authorizedsignatures = 0)
		)
		AND (
			n:Group
			OR n:Computer
			OR (
				n:User
				AND (
					n.gmsa = true
					OR n.msa = true
					OR (ct1.subjectaltrequiredns = false AND ct1.subjectaltrequiredomaindns = false)
				)
			)
		)

		MATCH p2 = (p2Principal)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct2:CertTemplate)-[:PublishedTo]->(eca2:EnterpriseCA)-[:TrustedForNTAuth]->(:NTAuthStore)-[:NTAuthStoreFor]->(d)
		WHERE (
			p2Principal.objectid = n.objectid
			OR p2Principal.objectid ENDS WITH '-S-1-5-11'
			OR p2Principal.objectid ENDS WITH '-S-1-1-0'
		)
		AND ct2.authenticationenabled = true
		AND ct2.requiresmanagerapproval = false

		MATCH p3 = (ct1)-[:EnrollOnBehalfOf]->(ct2)

		MATCH p4 = (p4Principal)-[:MemberOf*0..]->()-[:Enroll]->(eca1)
		WHERE (
			p4Principal.objectid = n.objectid
			OR p4Principal.objectid ENDS WITH '-S-1-5-11'
			OR p4Principal.objectid ENDS WITH '-S-1-1-0'
		)

		MATCH p5 = (p5Principal)-[:MemberOf*0..]->()-[:Enroll]->(eca2)
		WHERE (
			p5Principal.objectid = n.objectid
			OR p5Principal.objectid ENDS WITH '-S-1-5-11'
			OR p5Principal.objectid ENDS WITH '-S-1-1-0'
		)

		MATCH p6 = (eca1)-[:IssuedSignedBy|EnterpriseCAFor*1..]->(:RootCA)-[:RootCAFor]->(d)
		MATCH p7 = (eca2)-[:IssuedSignedBy|EnterpriseCAFor*1..]->(:RootCA)-[:RootCAFor]->(d)
		MATCH p9 = (host1:Computer)-[:HostsCAService]->(eca1)
		MATCH p10 = (host2:Computer)-[:HostsCAService]->(eca2)
		WHERE host1.enabled = true AND host2.enabled = true

		OPTIONAL MATCH p8 = (p8Principal)-[:MemberOf*0..]->()-[:DelegatedEnrollmentAgent]->(ct2)
		WHERE (
			p8Principal.objectid = n.objectid
			OR p8Principal.objectid ENDS WITH '-S-1-5-11'
			OR p8Principal.objectid ENDS WITH '-S-1-1-0'
		)

		WITH *
		WHERE (
			coalesce(eca2.enrollmentagentrestrictionscollected, false) = false
			OR coalesce(eca2.hasenrollmentagentrestrictions, false) = false
			OR p8 IS NOT NULL
		)

		RETURN p1, p2, p3, p4, p5, p6, p7, p8, p9, p10
	*/
	var (
		startNode  *graph.Node
		startNodes = graph.NodeSet{}

		traversalInst            = traversal.New(db, post.MaximumDatabaseParallelWorkers)
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
		finalEnterpriseCAs       = cardinality.NewBitmap64()
		hostPathsByEnterpriseCA  map[graph.ID]graph.PathSet
		path2CandidateTemplates  = cardinality.NewBitmap64()
		enrollOnBehalfOfPaths    graph.PathSet
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	}

	// Add startnode, Auth. Users, and Everyone to start nodes
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if nodeSet, err := FetchAuthUsersAndEveryoneGroups(tx); err != nil {
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

	if qualifyingHostPaths, err := fetchQualifyingEnterpriseCAHostPaths(ctx, db, enterpriseCANodes); err != nil {
		return nil, err
	} else {
		qualifyingEnterpriseCAs := cardinality.NewBitmap64()
		for enterpriseCAID := range qualifyingHostPaths {
			qualifyingEnterpriseCAs.Add(enterpriseCAID.Uint64())
		}

		enterpriseCANodes.And(qualifyingEnterpriseCAs)
		hostPathsByEnterpriseCA = qualifyingHostPaths
	}

	if enterpriseCANodes.Cardinality() == 0 {
		return paths, nil
	}

	//Use the enterprise CA nodes we gathered to filter the first set of paths for P1
	for _, n := range startNodes.Slice() {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: n,
			Driver: ADCSESC3Path1Pattern(edge.EndID, enterpriseCANodes).Do(func(terminal *graph.PathSegment) error {
				var (
					certTemplateNode = terminal.Search(func(nextSegment *graph.PathSegment) bool {
						return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
					})
					userStartNode = startNode.Kinds.ContainsOneOf(ad.User)
				)

				managedServiceAccount, err := isManagedServiceAccount(startNode)
				if err != nil {
					return err
				}

				lock.Lock()
				path1CandidateSegments[certTemplateNode.ID] = append(path1CandidateSegments[certTemplateNode.ID], terminal)

				// gMSAs and sMSAs are User nodes with DNS names, so DNS requirements are valid for them.
				if !userStartNode || managedServiceAccount || certTemplateValidForUserVictim(certTemplateNode) {
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
		if !enterpriseCANodes.Contains(ecaID.Uint64()) {
			continue
		}

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
				_, eca1HasQualifyingHost := hostPathsByEnterpriseCA[eca1.ID]
				_, eca2HasQualifyingHost := hostPathsByEnterpriseCA[eca2.ID]
				if !eca1HasQualifyingHost || !eca2HasQualifyingHost {
					continue
				}

				// Verify P6 and P7 paths exists
				p6segments, ok := path6_7CandidateSegments[eca1.ID]
				if !ok {
					continue
				}
				p7segments, ok := path6_7CandidateSegments[eca2.ID]
				if !ok {
					continue
				}

				if enterpriseCAHasEnrollmentAgentRestrictions(eca2) {
					// Verify p8 path exists.
					p8segments, ok := path8CandidateSegments[ct2.ID]
					if !ok {
						continue
					}

					for _, p8 := range p8segments {
						paths.AddPath(p8.Path())
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
				finalEnterpriseCAs.Add(eca1.ID.Uint64())
				finalEnterpriseCAs.Add(eca2.ID.Uint64())
			}
		}
	}

	finalEnterpriseCAs.Each(func(enterpriseCAID uint64) bool {
		paths.AddPathSet(hostPathsByEnterpriseCA[graph.ID(enterpriseCAID)])
		return true
	})

	return paths, nil
}

func enterpriseCAHasEnrollmentAgentRestrictions(enterpriseCA *graph.Node) bool {
	if collected, err := enterpriseCA.Properties.Get(ad.EnrollmentAgentRestrictionsCollected.String()).Bool(); err != nil || !collected {
		return false
	} else if hasRestrictions, err := enterpriseCA.Properties.Get(ad.HasEnrollmentAgentRestrictions.String()).Bool(); err != nil {
		return false
	} else {
		return hasRestrictions
	}
}

func ADCSESC3Path1Pattern(domainID graph.ID, enterpriseCAs cardinality.Duplex[uint64]) traversal.PatternContinuation {
	return enterpriseCATrustedForNTAuthToDomainPattern(traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.Kind(query.End(), ad.CertTemplate),
			query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
			query.Or(
				query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
				query.And(
					query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
				),
			),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.InIDs(query.End(), graph.DuplexToGraphIDs(enterpriseCAs)...),
			query.Kind(query.End(), ad.EnterpriseCA),
		)), domainID)
}

func ADCSESC3Path2Pattern(domainId graph.ID, enterpriseCAs, candidateTemplates cardinality.Duplex[uint64]) traversal.PatternContinuation {
	return enterpriseCATrustedForNTAuthToDomainPattern(traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
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
			query.InIDs(query.End(), graph.DuplexToGraphIDs(enterpriseCAs)...))), domainId)
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
	return enterpriseCAChainToDomainPattern(traversal.NewPattern(), domainId)
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
