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
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/specterops/bloodhound/dawgs/traversal"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

// PostNTLM is the initial function used to execute our NTLM analysis
func PostNTLM(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator, adcsCache ADCSCache, ntlmEnabled bool, compositionCounter *analysis.CompositionCounter) (*analysis.AtomicPostProcessingStats, error) {
	var (
		adcsComputerCache       = make(map[string]cardinality.Duplex[uint64])
		operation               = analysis.NewPostRelationshipOperation(ctx, db, "PostNTLM")
		authenticatedUsersCache = make(map[string]graph.ID)
		// compositionChannel      = make(chan analysis.CompositionInfo)
		protectedUsersCache = make(map[string]cardinality.Duplex[uint64])
	)

	// NTLM must be enabled through the feature flag
	if !ntlmEnabled {
		operation.Done()
		return &operation.Stats, nil
	}

	// This is a POC on how to pipe composition info up through the operations
	// go func() {
	//	count := 0
	//	edgeBuffer := make([]model.EdgeCompositionEdge, 0)
	//	nodeBuffer := make([]model.EdgeCompositionNode, 0)
	//
	//	for {
	//		if elem, hasNextElem := channels.Receive(ctx, compositionChannel); hasNextElem {
	//			count++
	//			edgeBuffer = append(edgeBuffer, elem.GetCompositionEdges()...)
	//			nodeBuffer = append(nodeBuffer, elem.GetCompositionNodes()...)
	//			if count == 100 {
	//				count = 0
	//
	//				if _, _, err := pgDB.CreateCompositionInfo(ctx, nodeBuffer, edgeBuffer); err != nil {
	//					slog.ErrorContext(ctx, fmt.Sprintf("error creating composition info: %v", err))
	//				}
	//			}
	//		} else {
	//			break
	//		}
	//	}
	// }()

	// TODO: after adding all of our new NTLM edges, benchmark performance between submitting multiple readers per computer or single reader per computer
	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		// Fetch all nodes where the node is a Group and is an Authenticated User
		if innerAuthenticatedUsersCache, err := FetchAuthUsersMappedToDomains(tx); err != nil {
			return err
		} else if innerProtectedUsersCache, err := FetchProtectedUsersMappedToDomains(ctx, db, groupExpansions); err != nil {
			return err
		} else if ldapSigningCache, err := FetchLDAPSigningCache(ctx, db); err != nil {
			return err
		} else {
			authenticatedUsersCache = innerAuthenticatedUsersCache
			protectedUsersCache = innerProtectedUsersCache

			// Fetch all nodes where the type is Computer
			return tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for computer := range cursor.Chan() {
					innerComputer := computer

					domain, err := innerComputer.Properties.Get(ad.DomainSID.String()).String()

					if err != nil {
						continue
					} else if authenticatedUserGroupID, ok := authenticatedUsersCache[domain]; !ok {
						continue
					} else if protectedUsersForDomain, ok := protectedUsersCache[domain]; !ok {
						continue
					} else if ldapSigningForDomain, ok := ldapSigningCache[domain]; !ok {
						continue
					} else if protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsValidFunctionalLevel {
						continue
					} else if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						return PostCoerceAndRelayNTLMToSMB(tx, outC, groupExpansions, innerComputer, authenticatedUserGroupID, compositionCounter)
					}); err != nil {
						slog.WarnContext(ctx, fmt.Sprintf("Post processing failed for %s: %v", ad.CoerceAndRelayNTLMToSMB, err))
						// Additional analysis may occur if one of our analysis errors
						continue
					} else if err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
						return PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserGroupID, ldapSigningCache)
					}); err != nil {
						slog.WarnContext(ctx, fmt.Sprintf("Post processing failed for %s: %v", ad.CoerceAndRelayNTLMToLDAP, err))
						continue
					}

					if webclientRunning, err := innerComputer.Properties.Get(ad.WebClientRunning.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
						slog.WarnContext(ctx, fmt.Sprintf("Error getting webclientrunningproperty from computer %d", innerComputer.ID))
					} else if restrictOutboundNtlm, err := innerComputer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
						slog.WarnContext(ctx, fmt.Sprintf("Error getting restrictoutboundntlm from computer %d", innerComputer.ID))
					} else if webclientRunning && !restrictOutboundNtlm {
						if _, ok := adcsComputerCache[domain]; !ok {
							adcsComputerCache[domain] = cardinality.NewBitmap64()
						}
						adcsComputerCache[domain].Add(innerComputer.ID.Uint64())
					}
				}

				return cursor.Error()
			})
		}
	})
	if err != nil {
		operation.Done()
		return nil, err
	}

	if err := PostCoerceAndRelayNTLMToADCS(adcsCache, operation, authenticatedUsersCache, adcsComputerCache); err != nil {
		return nil, err
	}

	return &operation.Stats, operation.Done()
}

func GetCoerceAndRelayNTLMtoADCSEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		startNode  *graph.Node
		endNode    *graph.Node
		domainNode *graph.Node

		traversalInst      = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths              = graph.PathSet{}
		candidateSegments  = map[graph.ID][]*graph.PathSegment{}
		path1EnterpriseCAs = cardinality.NewBitmap64()
		path2EnterpriseCAs = cardinality.NewBitmap64()
		lock               = &sync.Mutex{}
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

	if domainsid, err := endNode.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("Error getting domain SID for domain %d: %v", endNode.ID, err))
		return nil, err
	} else if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		domainNode, err = analysis.FetchNodeByObjectID(tx, domainsid)
		return err
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: coerceAndRelayNTLMtoADCSPath1Pattern(domainNode.ID).Do(func(terminal *graph.PathSegment) error {
			var enterpriseCANode *graph.Node
			terminal.WalkReverse(func(nextSegment *graph.PathSegment) bool {
				if nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) {
					enterpriseCANode = nextSegment.Node
				}
				return true
			})

			lock.Lock()
			candidateSegments[enterpriseCANode.ID] = append(candidateSegments[enterpriseCANode.ID], terminal)
			path1EnterpriseCAs.Add(enterpriseCANode.ID.Uint64())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: coerceAndRelayNTLMtoADCSPath2Pattern(domainNode.ID, path1EnterpriseCAs).Do(func(terminal *graph.PathSegment) error {
			enterpriseCANode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			candidateSegments[enterpriseCANode.ID] = append(candidateSegments[enterpriseCANode.ID], terminal)
			path2EnterpriseCAs.Add(enterpriseCANode.ID.Uint64())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	// Intersect the CAs and take only those seen in both paths
	path1EnterpriseCAs.And(path2EnterpriseCAs)
	// Render paths from the segments
	path1EnterpriseCAs.Each(func(value uint64) bool {
		for _, segment := range candidateSegments[graph.ID(value)] {
			paths.AddPath(segment.Path())
		}

		return true
	})

	return paths, nil
}

func coerceAndRelayNTLMtoADCSPath1Pattern(domainID graph.ID) traversal.PatternContinuation {
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
				),
				query.And(
					query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
				),
			),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.Kind(query.End(), ad.EnterpriseCA),
			query.Or(
				query.Equals(query.EndProperty(ad.ADCSWebEnrollmentHTTP.String()), true),
				query.And(
					query.Equals(query.EndProperty(ad.ADCSWebEnrollmentHTTPS.String()), true),
					query.Equals(query.EndProperty(ad.ADCSWebEnrollmentHTTPSEPA.String()), false),
				),
			),
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

func coerceAndRelayNTLMtoADCSPath2Pattern(domainID graph.ID, enterpriseCAs cardinality.Duplex[uint64]) traversal.PatternContinuation {
	return traversal.NewPattern().OutboundWithDepth(0, 0, query.And(
		query.Kind(query.Relationship(), ad.MemberOf),
		query.Kind(query.End(), ad.Group),
	)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.Enroll),
			query.InIDs(query.EndID(), graph.DuplexToGraphIDs(enterpriseCAs)...),
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

func PostCoerceAndRelayNTLMToADCS(adcsCache ADCSCache, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], authUsersCache map[string]graph.ID, adcsComputerCache map[string]cardinality.Duplex[uint64]) error {
	for _, outerDomain := range adcsCache.domains {
		for _, outerEnterpriseCA := range adcsCache.GetEnterpriseCertAuthorities() {
			domain := outerDomain
			enterpriseCA := outerEnterpriseCA
			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if publishedCertTemplates := adcsCache.GetPublishedTemplateCache(enterpriseCA.ID); len(publishedCertTemplates) == 0 {
					// If this enterprise CA has no published templates, then there's no reason to check further
					return nil
				} else if !adcsCache.DoesCAChainProperlyToDomain(enterpriseCA, domain) {
					// If the CA doesn't chain up to the domain properly then its invalid
					return nil
				} else if ecaValid, err := isEnterpriseCAValidForADCS(enterpriseCA); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error validating EnterpriseCA %d for ADCS relay: %v", enterpriseCA.ID, err))
					return nil
				} else if !ecaValid {
					// Check some prereqs on the enterprise CA. If the enterprise CA is invalid, we can fast skip it
					return nil
				} else if domainsid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					slog.WarnContext(ctx, fmt.Sprintf("Error getting domainsid for domain %d: %v", domain.ID, err))
					return nil
				} else if authUsersGroup, ok := authUsersCache[domainsid]; !ok {
					// If we cant find an auth users group for this domain then we're not going to be able to make an edge regardless
					slog.WarnContext(ctx, fmt.Sprintf("Unable to find auth users group for domain %s", domainsid))
					return nil
				} else {
					// If auth users doesn't have enroll rights here than it's not valid either. Unroll enrollers into a slice and check if auth users is in it
					ecaEnrollers := adcsCache.GetEnterpriseCAEnrollers(enterpriseCA.ID)
					authUsersHasEnrollmentRights := false
					for _, l := range ecaEnrollers {
						if l.ID == authUsersGroup {
							authUsersHasEnrollmentRights = true
							break
						}
					}

					if !authUsersHasEnrollmentRights {
						return nil
					}

					for _, certTemplate := range publishedCertTemplates {
						if valid, err := isCertTemplateValidForADCSRelay(certTemplate); err != nil {
							slog.ErrorContext(ctx, fmt.Sprintf("Error validating cert template %d for NTLM ADCS relay: %v", certTemplate.ID, err))
							continue
						} else if !valid {
							continue
						} else if computers, ok := adcsComputerCache[domainsid]; !ok {
							continue
						} else {
							computers.Each(func(value uint64) bool {
								outC <- analysis.CreatePostRelationshipJob{
									FromID: authUsersGroup,
									ToID:   graph.ID(value),
									Kind:   ad.CoerceAndRelayNTLMToADCS,
								}
								return true
							})
						}
					}

					return nil
				}
			})
		}
	}

	return nil
}

func isEnterpriseCAValidForADCS(eca *graph.Node) (bool, error) {
	if httpEnrollment, err := eca.Properties.Get(ad.ADCSWebEnrollmentHTTP.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return false, err
	} else if httpEnrollment {
		return true, nil
	} else if httpsEnrollment, err := eca.Properties.Get(ad.ADCSWebEnrollmentHTTPS.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return false, err
	} else if !httpsEnrollment {
		return false, nil
	} else if httpsEnrollmentEpa, err := eca.Properties.Get(ad.ADCSWebEnrollmentHTTPSEPA.String()).Bool(); err != nil {
		if errors.Is(err, graph.ErrPropertyNotFound) {
			return false, nil
		}
		return false, err
	} else {
		return !httpsEnrollmentEpa, nil
	}
}

func isCertTemplateValidForADCSRelay(ct *graph.Node) (bool, error) {
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
	} else if schemaVersion <= 1 {
		return true, nil
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		return false, err
	} else {
		return authorizedSignatures == 0, nil
	}
}

func GetCoerceAndRelayNTLMtoSMBComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		endNode *graph.Node
		paths   = graph.PathSet{}
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if endNode, err = ops.FetchNode(tx, edge.EndID); err != nil {
			return err
		} else if paths, err = ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      endNode,
			Direction: graph.DirectionInbound,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), ad.MemberOf, ad.AdminTo)
			},
			DescentFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				if segment.Depth() > 1 && segment.Trunk.Node.Kinds.ContainsOneOf(ad.Computer, ad.User) {
					return false
				}

				return true
			},
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				return segment.Node.Kinds.ContainsOneOf(ad.Computer)
			},
			Skip:  0,
			Limit: 0,
		}); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	}

	return paths, nil
}

// PostCoerceAndRelayNTLMToSMB creates edges that allow a computer with unrolled admin access to one or more computers where SMB signing is disabled.
// Comprised solely of adminTo and memberOf edges
func PostCoerceAndRelayNTLMToSMB(tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, expandedGroups impact.PathAggregator, computer *graph.Node, authenticatedUserID graph.ID, compositionCounter *analysis.CompositionCounter) error {
	if smbSigningEnabled, err := computer.Properties.Get(ad.SMBSigning.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		return nil
	} else if err != nil {
		return err
	} else if restrictOutboundNtlm, err := computer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		return nil
	} else if err != nil {
		return err
	} else if !smbSigningEnabled && !restrictOutboundNtlm {
		// Fetch the admins with edges to the provided computer
		if firstDegreeAdmins, err := fetchFirstDegreeNodes(tx, computer, ad.AdminTo); err != nil {
			return err
		} else if firstDegreeAdmins.ContainingNodeKinds(ad.Computer).Len() > 0 {
			// compositionID := compositionCounter.Get()
			outC <- analysis.CreatePostRelationshipJob{
				FromID: authenticatedUserID,
				ToID:   computer.ID,
				Kind:   ad.CoerceAndRelayNTLMToSMB,
				// RelProperties: map[string]any{common.CompositionID.String(): compositionID},
			}
			// This is an example of how you would use the composition counter
			// compositionC <- analysis.CompositionInfo{
			//	CompositionID: compositionID,
			//	EdgeIDs:       nil,
			//	NodeIDs:       nil,
			// }
		} else {
			allAdminGroups := cardinality.NewBitmap64()
			for group := range firstDegreeAdmins.ContainingNodeKinds(ad.Group) {
				allAdminGroups.Or(expandedGroups.Cardinality(group.Uint64()))
			}

			currentComputerID, err := computer.Properties.Get(common.ObjectID.String()).String()
			if err != nil {
				return err
			}

			// Fetch nodes where the node id is in our allAdminGroups bitmap, are of type Computer, and are not the target computer
			if computerIds, err := ops.FetchNodeIDs(tx.Nodes().Filter(
				query.And(
					query.InIDs(query.Node(), graph.DuplexToGraphIDs(allAdminGroups)...),
					query.Kind(query.Node(), ad.Computer),
					query.Not(query.Equals(query.NodeProperty(common.ObjectID.String()), currentComputerID)),
				),
			)); err != nil {
				return err
			} else if len(computerIds) > 0 {
				outC <- analysis.CreatePostRelationshipJob{
					FromID: authenticatedUserID,
					ToID:   computer.ID,
					Kind:   ad.CoerceAndRelayNTLMToSMB,
					// RelProperties: map[string]any{common.CompositionID.String(): compositionCounter.Get()},
				}
			}
		}
	}

	return nil
}

func GetVulnerableDomainControllersForRelayNTLMtoLDAP(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.NodeSet, error) {
	var (
		startNode *graph.Node
		nodes     graph.NodeSet
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

	if domainsid, err := startNode.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("Error getting domain SID for domain %d: %v", startNode.ID, err))
		return nil, err
	} else if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var ierr error
		nodes, ierr = ops.FetchNodeSet(tx.Nodes().Filter(
			query.And(
				query.Kind(query.Node(), ad.Computer),
				query.Equals(query.NodeProperty(ad.IsDC.String()), true),
				query.Equals(query.NodeProperty(ad.LDAPAvailable.String()), true),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainsid),
				query.Equals(query.NodeProperty(ad.LDAPSigning.String()), false),
			),
		))

		return ierr
	}); err != nil {
		return nil, err
	} else {
		return nodes, nil
	}
}

func GetVulnerableDomainControllersForRelayNTLMtoLDAPS(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.NodeSet, error) {
	var (
		startNode *graph.Node
		nodes     graph.NodeSet
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

	if domainsid, err := startNode.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		slog.WarnContext(ctx, fmt.Sprintf("Error getting domain SID for domain %d: %v", startNode.ID, err))
		return nil, err
	} else if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var ierr error
		nodes, ierr = ops.FetchNodeSet(tx.Nodes().Filter(
			query.And(
				query.Kind(query.Node(), ad.Computer),
				query.Equals(query.NodeProperty(ad.IsDC.String()), true),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainsid),
				query.Equals(query.NodeProperty(ad.LDAPSEPA.String()), false),
				query.Equals(query.NodeProperty(ad.LDAPSAvailable.String()), true),
			),
		))

		return ierr
	}); err != nil {
		return nil, err
	} else {
		return nodes, nil
	}
}

// PostCoerceAndRelayNTLMToLDAP creates edges where an authenticated user group, for a given domain, is able to target the provided computer.
// This will create either a CoerceAndRelayNTLMToLDAP or CoerceAndRelayNTLMToLDAPS edges, depending on the ldapSigning property of the domain
func PostCoerceAndRelayNTLMToLDAP(outC chan<- analysis.CreatePostRelationshipJob, computer *graph.Node, authenticatedUserGroupID graph.ID, ldapSigningCache map[string]LDAPSigningCache) error {
	// restrictoutboundntlm must be set to false and webclientrunning must be set to true for the computer's properties
	// in order for this attack path to be viable
	// If the property is not found, we will assume false
	if restrictOutboundNtlm, err := computer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return err
	} else if restrictOutboundNtlm {
		return nil
	} else if webClientRunning, err := computer.Properties.Get(ad.WebClientRunning.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
		return err
	} else if webClientRunning {
		if domainSid, err := computer.Properties.Get(ad.DomainSID.String()).String(); err != nil {
			if errors.Is(err, graph.ErrPropertyNotFound) {
				return nil
			} else {
				return err
			}
		} else {
			if signingCache, ok := ldapSigningCache[domainSid]; !ok {
				// If no DomainSID was found in our cache, then we can simply ignore the domain as it does not match our path creation criteria
				return nil
			} else {
				// We will create relationships from the AuthenticatedUsers group to the vulnerable computer,
				// for both LDAP and LDAPS scenarios, assuming the passed in signingCache has any vulnerable paths
				// We also ignore instances where the computer is relaying to itself
				if len(signingCache.relayableToDCLDAP) == 1 && signingCache.relayableToDCLDAP[0] != computer.ID {
					outC <- analysis.CreatePostRelationshipJob{
						FromID: authenticatedUserGroupID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToLDAP,
					}
				} else if len(signingCache.relayableToDCLDAP) > 1 {
					outC <- analysis.CreatePostRelationshipJob{
						FromID: authenticatedUserGroupID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToLDAP,
					}
				}

				if len(signingCache.relayableToDCLDAPS) == 1 && signingCache.relayableToDCLDAPS[0] != computer.ID {
					outC <- analysis.CreatePostRelationshipJob{
						FromID: authenticatedUserGroupID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToLDAP,
					}
				} else if len(signingCache.relayableToDCLDAPS) > 1 {
					outC <- analysis.CreatePostRelationshipJob{
						FromID: authenticatedUserGroupID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToLDAPS,
					}
				}
			}
		}
	}

	return nil
}

// FetchAuthUsersMappedToDomains fetches all nodes where the node is a Group and is an Authenticated User
func FetchAuthUsersMappedToDomains(tx graph.Transaction) (map[string]graph.ID, error) {
	authenticatedUsers := make(map[string]graph.ID)

	err := tx.Nodes().Filter(
		query.And(
			query.Kind(query.Node(), ad.Group),
			query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), AuthenticatedUsersSuffix)),
	).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for authenticatedUser := range cursor.Chan() {
			if domain, err := authenticatedUser.Properties.Get(ad.DomainSID.String()).String(); err != nil {
				continue
			} else {
				authenticatedUsers[domain] = authenticatedUser.ID
			}
		}

		return cursor.Error()
	},
	)

	return authenticatedUsers, err
}

// FetchProtectedUsersMappedToDomains fetches all protected users groups mapped by their domain SID
func FetchProtectedUsersMappedToDomains(ctx context.Context, db graph.Database, groupExpansions impact.PathAggregator) (map[string]cardinality.Duplex[uint64], error) {
	protectedUsers := make(map[string]cardinality.Duplex[uint64])

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Filter(
			query.And(
				query.Kind(query.Node(), ad.Group),
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), ProtectedUsersSuffix)),
		).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for protectedUserGroup := range cursor.Chan() {
				if domain, err := protectedUserGroup.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					continue
				} else {
					set := cardinality.NewBitmap64()
					set.Or(groupExpansions.Cardinality(protectedUserGroup.ID.Uint64()))
					protectedUsers[domain] = set
				}
			}

			return cursor.Error()
		},
		)
	})

	return protectedUsers, err
}

// LDAPSigningCache encapsulates whether a domain had a valid functionallevel property and slices of node ids that meet the criteria
// for a CoerceAndRelayNTLMToLDAP or CoerceAndRelayNTLMToLDAPS edge
type LDAPSigningCache struct {
	IsValidFunctionalLevel bool
	relayableToDCLDAP      []graph.ID
	relayableToDCLDAPS     []graph.ID
}

// FetchLDAPSigningCache will check all Domain Controllers (DCs) for LDAP signing. If the DC has the "ldap_signing" set to true along with "ldaps_available" to true and "ldaps_epa" to false,
// we add the DC to the relayableToDCLDAPS slice. If the DC has "ldap_signing" set to false then we simply set the DC to be a relayableToDCLDAP" slice
func FetchLDAPSigningCache(ctx context.Context, db graph.Database) (map[string]LDAPSigningCache, error) {
	if domains, err := FetchAllDomains(ctx, db); err != nil {
		return nil, err
	} else {
		cache := make(map[string]LDAPSigningCache)
		// Iterate all domains to obtain the DomainSID, which we can use to query for DCs that control the domain
		for _, domain := range domains {
			if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
				if errors.Is(err, graph.ErrPropertyNotFound) {
					continue
				} else {
					return nil, err
				}
			} else if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
				if relayableToDcLdap, err := ops.FetchNodeIDs(tx.Nodes().Filter(
					query.And(
						query.Equals(
							query.NodeProperty(ad.DomainSID.String()), domainSid,
						),
						// IsDC is a property for computers that are Domain Controllers
						// This allows us to ensure the computer has a DCFor relationship to the currently iterated domain
						query.Equals(
							query.NodeProperty(ad.IsDC.String()), true,
						),
						query.Equals(
							query.NodeProperty(ad.LDAPSigning.String()), false,
						),
						query.Equals(
							query.NodeProperty(ad.LDAPAvailable.String()), true,
						),
					),
				)); err != nil {
					return err
				} else if relayableToDcLdaps, err := ops.FetchNodeIDs(tx.Nodes().Filter(
					query.And(
						query.Equals(
							query.NodeProperty(ad.DomainSID.String()), domainSid,
						),
						query.Equals(
							query.NodeProperty(ad.IsDC.String()), true,
						),
						query.Equals(
							query.NodeProperty(ad.LDAPSAvailable.String()), true,
						),
						query.Equals(
							query.NodeProperty(ad.LDAPSEPA.String()), false,
						),
					),
				)); err != nil {
					return err
				} else {
					// Domains with a functionallevel property after 2012 are protected from this attack path
					// This will ensure that the domain is vulnerable
					// If the domain does not have this property set, we will assume that the domain is protected
					isFunctionalLevelValid := false
					if functionalLevel, err := domain.Properties.Get(ad.FunctionalLevel.String()).String(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
						return err
					} else if slices.Contains(vulnerableFunctionalLevels(), functionalLevel) {
						isFunctionalLevelValid = true
					}

					cache[domainSid] = LDAPSigningCache{
						IsValidFunctionalLevel: isFunctionalLevelValid,
						relayableToDCLDAP:      relayableToDcLdap,
						relayableToDCLDAPS:     relayableToDcLdaps,
					}

					return nil
				}
			}); err != nil {
				return nil, err
			}
		}

		return cache, nil
	}
}

// vulnerableFunctionalLevels is a simple constant slice of releases that are vulnerable to a CoerceAndRelayNTLMToLDAP(S) attack path
// They can be used by checking a node's functionalllevel property
func vulnerableFunctionalLevels() []string {
	return []string{
		"2000 Mixed/Native",
		"2003 Interim",
		"2003",
		"2008",
		"2008 R2",
		"2012",
	}
}
