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
	"log/slog"
	"slices"
	"sync"

	"github.com/specterops/bloodhound/packages/go/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/traversal"
)

type NTLMCache struct {
	AuthenticatedUsersCache   map[string]graph.ID
	ProtectedUsersCache       map[string]cardinality.Duplex[uint64]
	LdapCache                 map[string]LDAPSigningCache
	UnprotectedComputersCache cardinality.Duplex[uint64]
	LocalGroupData            *LocalGroupData
}

func (s NTLMCache) GetAuthenticatedUserGroupForDomain(domainSid string) (graph.ID, bool) {
	id, ok := s.AuthenticatedUsersCache[domainSid]
	return id, ok
}

func (s NTLMCache) GetProtectedUsersForDomain(domainSid string) (cardinality.Duplex[uint64], bool) {
	protectedUsers, ok := s.ProtectedUsersCache[domainSid]
	return protectedUsers, ok
}

func (s NTLMCache) GetLdapCacheForDomain(domainSid string) (LDAPSigningCache, bool) {
	cache, ok := s.LdapCache[domainSid]
	return cache, ok
}

func NewNTLMCache(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (NTLMCache, error) {
	var (
		ntlmCache                   = NTLMCache{}
		unprotectedComputerCache    = make(map[string]cardinality.Duplex[uint64])
		allUnprotectedComputerCache = cardinality.NewBitmap64()
	)

	return ntlmCache, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		// Fetch all nodes where the node is a Group and is an Authenticated User
		if innerAuthenticatedUsersCache, err := FetchAuthUsersMappedToDomains(tx); err != nil {
			return err
		} else if innerProtectedUsersCache, err := FetchProtectedUsersMappedToDomains(ctx, db, localGroupData); err != nil {
			return err
		} else if ldapSigningCache, err := FetchLDAPSigningCache(ctx, db); err != nil {
			return err
		} else {
			ntlmCache.AuthenticatedUsersCache = innerAuthenticatedUsersCache
			ntlmCache.LdapCache = ldapSigningCache
			ntlmCache.ProtectedUsersCache = innerProtectedUsersCache
			ntlmCache.LocalGroupData = localGroupData

			// Fetch all nodes where the type is Computer and build out a cache of computers that are acceptable target/victims for coercion
			return tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for computer := range cursor.Chan() {
					innerComputer := computer

					if domainSid, err := innerComputer.Properties.Get(ad.DomainSID.String()).String(); err != nil {
						continue
					} else if _, ok := ntlmCache.GetAuthenticatedUserGroupForDomain(domainSid); !ok {
						continue
					} else if ldapSigningForDomain, ok := ntlmCache.GetLdapCacheForDomain(domainSid); !ok {
						continue
					} else if restrictOutboundNtlm, err := innerComputer.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); err != nil {
						// If we've failed to retrieve the property because it doesn't exist we'll fail closed here. We will treat it as if it is protected to prevent false positives
						if !errors.Is(err, graph.ErrPropertyNotFound) {
							slog.WarnContext(
								ctx,
								"Error getting restrictoutboundntlm from computer",
								slog.Uint64("computer_id", uint64(innerComputer.ID)),
								attr.Error(err),
							)
						}
						continue
					} else if restrictOutboundNtlm {
						continue
					} else {
						// Check if the computer is in protected users. If it is and the functional level isn't vulnerable, this computer isn't vulnerable.
						// If protected users doesn't exist, we intentionally fail open here as it is valid for older domains to not have this group
						if protectedUsersForDomain, ok := ntlmCache.GetProtectedUsersForDomain(domainSid); ok && protectedUsersForDomain.Contains(innerComputer.ID.Uint64()) && !ldapSigningForDomain.IsVulnerableFunctionalLevel {
							continue
						}

						if _, ok := unprotectedComputerCache[domainSid]; !ok {
							unprotectedComputerCache[domainSid] = cardinality.NewBitmap64()
						}
						unprotectedComputerCache[domainSid].Add(innerComputer.ID.Uint64())
						allUnprotectedComputerCache.Add(innerComputer.ID.Uint64())
					}
				}

				ntlmCache.UnprotectedComputersCache = allUnprotectedComputerCache

				return cursor.Error()
			})
		}
	})
}

// PostNTLM is the initial function used to execute our NTLM analysis
func PostNTLM(ctx context.Context, db graph.Database, localGroupData *LocalGroupData, adcsCache *ADCSCache, ntlmEnabled bool) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing NTLM",
		attr.Namespace("analysis"),
		attr.Function("PostNTLM"),
		attr.Scope("process"),
	)()

	var (
		operation = post.NewPostRelationshipOperation(ctx, db, "PostNTLM")
	)

	// NTLM must be enabled through the feature flag
	if !ntlmEnabled {
		operation.Done()
		return &operation.Stats, nil
	}

	// TODO: after adding all of our new NTLM edges, benchmark performance between submitting multiple readers per computer or single reader per computer
	// First fetch pre-reqs + find all vulnerable computers that are not protected
	if ntlmCache, err := NewNTLMCache(ctx, db, localGroupData); err != nil {
		operation.Done()
		return nil, err
	} else if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Filter(query.Kind(query.Node(), ad.Computer)).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for computer := range cursor.Chan() {
				innerComputer := computer

				if domainSid, err := innerComputer.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					continue
				} else if authenticatedUserGroupID, ok := ntlmCache.GetAuthenticatedUserGroupForDomain(domainSid); !ok {
					continue
				} else {
					if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
						return PostCoerceAndRelayNTLMToSMB(tx, outC, ntlmCache, innerComputer, authenticatedUserGroupID)
					}); err != nil {
						slog.WarnContext(
							ctx,
							"Post processing failed for CoerceAndRelayNTLMToSMB",
							attr.Error(err),
						)
						// Additional analysis may occur if one of our analysis errors
						continue
					}

					// Any computers that are restricted/protected are not valid targets for the next relays
					if !ntlmCache.UnprotectedComputersCache.Contains(innerComputer.ID.Uint64()) {
						continue
					}

					if err = operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
						return PostCoerceAndRelayNTLMToLDAP(outC, innerComputer, authenticatedUserGroupID, ntlmCache.LdapCache)
					}); err != nil {
						slog.WarnContext(
							ctx,
							"Post processing failed for CoerceAndRelayNTLMToLDAP",
							attr.Error(err),
						)
						continue
					}
				}
			}

			return cursor.Error()
		})
	}); err != nil {
		operation.Done()
		return nil, err
	} else {
		if err := PostCoerceAndRelayNTLMToADCS(ctx, operation, adcsCache, ntlmCache); err != nil {
			operation.Done()
			return nil, err
		}

		return &operation.Stats, operation.Done()
	}
}

func GetCoerceAndRelayNTLMtoADCSEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		startNode  *graph.Node
		endNode    *graph.Node
		domainNode *graph.Node
		startNodes = graph.NodeSet{}

		traversalInst      = traversal.New(db, post.MaximumDatabaseParallelWorkers)
		paths              = graph.PathSet{}
		candidateSegments  = map[graph.ID][]*graph.PathSegment{}
		path1EnterpriseCAs = cardinality.NewBitmap64()
		path2EnterpriseCAs = cardinality.NewBitmap64()
		lock               = &sync.Mutex{}
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if nodeSet, err := FetchAuthUsersAndEveryoneGroups(tx); err != nil {
			return err
		} else if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else if endNode, err = ops.FetchNode(tx, edge.EndID); err != nil {
			return err
		} else {
			startNodes.AddSet(nodeSet)
			startNodes.Add(endNode)
			return nil
		}
	}); err != nil {
		return nil, err
	}

	if domainsid, err := startNode.Properties.Get(ad.DomainSID.String()).String(); err != nil {
		slog.WarnContext(
			ctx,
			"Error getting domain SID for start node",
			slog.Uint64("node_id", uint64(startNode.ID)),
			attr.Error(err),
		)
		return nil, err
	} else if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		domainNode, err = tx.Nodes().Filter(query.And(
			query.Equals(query.NodeProperty(common.ObjectID.String()), domainsid),
			query.Kind(query.Node(), ad.Entity),
		)).First()
		return err
	}); err != nil {
		return nil, err
	}

	for _, startNode := range startNodes.Slice() {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: startNode,
			Driver: coerceAndRelayNTLMtoADCSPath1Pattern(domainNode.ID).Do(func(terminal *graph.PathSegment) error {
				var (
					certTemplateNode *graph.Node
					enterpriseCANode *graph.Node
				)

				terminal.WalkReverse(func(nextSegment *graph.PathSegment) bool {
					if nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) {
						enterpriseCANode = nextSegment.Node
					} else if nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate) {
						certTemplateNode = nextSegment.Node
					}
					return true
				})

				if certTemplateNode == nil || enterpriseCANode == nil {
					return nil
				} else if valid := isCertTemplateValidForADCSRelay(ctx, certTemplateNode); !valid {
					return nil
				}

				lock.Lock()
				candidateSegments[enterpriseCANode.ID] = append(candidateSegments[enterpriseCANode.ID], terminal)
				path1EnterpriseCAs.Add(enterpriseCANode.ID.Uint64())
				lock.Unlock()

				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	for _, n := range startNodes.Slice() {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: n,
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
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.Kind(query.End(), ad.EnterpriseCA),
			query.Equals(query.EndProperty(ad.HasVulnerableEndpoint.String()), true),
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

func PostCoerceAndRelayNTLMToADCS(ctx context.Context, operation post.StatTrackedOperation[post.EnsureRelationshipJob], adcsCache *ADCSCache, ntlmCache NTLMCache) error {
	for eca, chains := range adcsCache.GetECAHostedChainedDomains() {
		ecaID := graph.ID(eca)

		if vulnerable := hasVulnerableEndpoint(chains.EnterpriseCA); !vulnerable {
			continue
		} else if publishedCertTemplates := adcsCache.GetPublishedTemplateCache(ecaID); len(publishedCertTemplates) == 0 {
			// If this enterprise CA has no published templates, then there's no reason to check further
			continue
		} else {
			ecaEnrollers := adcsCache.GetEnterpriseCAEnrollers(ecaID)

			for _, certTemplate := range publishedCertTemplates {
				// Verify cert template enables authentication and get cert template enrollers
				if valid := isCertTemplateValidForADCSRelay(ctx, certTemplate); !valid {
					continue
				} else if certTemplateEnrollers := adcsCache.GetCertTemplateEnrollers(certTemplate.ID); len(certTemplateEnrollers) == 0 {
					continue
				} else {
					victims := getVictimBitmap(
						ntlmCache.LocalGroupData,
						certTemplateEnrollers,
						ecaEnrollers,
						adcsCache.GetCertTemplateHasSpecialEnrollers(certTemplate.ID),
						adcsCache.GetEnterpriseCAHasSpecialEnrollers(ecaID),
					)

					victims.And(ntlmCache.UnprotectedComputersCache)

					var submitErr error
					chains.Domains.Each(func(domain uint64) bool {
						if authUsersGroup, ok := adcsCache.GetAuthUserForDomain(graph.ID(domain)); ok {
							if err := operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
								victims.Each(func(target uint64) bool {
									outC <- post.EnsureRelationshipJob{
										FromID: authUsersGroup,
										ToID:   graph.ID(target),
										Kind:   ad.CoerceAndRelayNTLMToADCS,
									}
									return true
								})
								return nil
							}); err != nil {
								submitErr = err
								return false
							}
						}
						return true
					})
					if submitErr != nil {
						return submitErr
					}
				}
			}
		}
	}

	return nil
}

func hasVulnerableEndpoint(eca *graph.Node) bool {
	if vulnerable, err := eca.Properties.Get(ad.HasVulnerableEndpoint.String()).Bool(); err != nil {
		return false
	} else {
		return vulnerable
	}
}

func isCertTemplateValidForADCSRelay(ctx context.Context, ct *graph.Node) bool {
	if reqManagerApproval, err := ct.Properties.Get(ad.RequiresManagerApproval.String()).Bool(); err != nil {
		logPropertyLookupFailure(ctx, ct, err)
		return false
	} else if reqManagerApproval {
		return false
	} else if authenticationEnabled, err := ct.Properties.Get(ad.AuthenticationEnabled.String()).Bool(); err != nil {
		logPropertyLookupFailure(ctx, ct, err)
		return false
	} else if !authenticationEnabled {
		return false
	} else if schemaVersion, err := ct.Properties.Get(ad.SchemaVersion.String()).Float64(); err != nil {
		logPropertyLookupFailure(ctx, ct, err)
		return false
	} else if schemaVersion <= 1 {
		return true
	} else if authorizedSignatures, err := ct.Properties.Get(ad.AuthorizedSignatures.String()).Float64(); err != nil {
		logPropertyLookupFailure(ctx, ct, err)
		return false
	} else {
		return authorizedSignatures == 0
	}
}

func GetCoerceAndRelayNTLMtoSMBEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		startNode *graph.Node
		endNode   *graph.Node
		pathSet   graph.PathSet
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else if endNode, err = ops.FetchNode(tx, edge.EndID); err != nil {
			return err
		} else if domainsid, err := startNode.Properties.Get(ad.DomainSID.String()).String(); err != nil {
			slog.WarnContext(
				ctx,
				"Error getting domain SID for domain",
				slog.Uint64("node_id", uint64(startNode.ID)),
				attr.Error(err),
			)
			return err
		} else if innerPathSet, err := ops.TraversePaths(tx, ops.TraversalPlan{
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
				node := segment.Node
				if node.Kinds.ContainsOneOf(ad.User) {
					return false
				} else if nodeDomainSid, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					return false
				} else if nodeDomainSid != domainsid {
					return false
				} else if restrictNtlm, err := node.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); err != nil || restrictNtlm {
					return false
				} else {
					return true
				}
			},
		}); err != nil {
			return err
		} else {
			pathSet = innerPathSet
			return nil
		}
	}); err != nil {
		return nil, err
	} else {
		return pathSet, nil
	}
}

// PostCoerceAndRelayNTLMToSMB creates edges that allow a computer with unrolled admin access to one or more computers where SMB signing is disabled.
// Comprised solely of adminTo and memberOf edges
func PostCoerceAndRelayNTLMToSMB(tx graph.Transaction, outC chan<- post.EnsureRelationshipJob, ntlmCache NTLMCache, computer *graph.Node, authenticatedUserID graph.ID) error {
	if smbSigningEnabled, err := computer.Properties.Get(ad.SMBSigning.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
		return nil
	} else if err != nil {
		return err
	} else if !smbSigningEnabled {
		// Fetch the admins with edges to the provided computer
		if firstDegreeAdmins, err := fetchFirstDegreeNodes(tx, computer, ad.AdminTo); err != nil {
			return err
		} else {
			allAdminPrincipals := cardinality.NewBitmap64()

			for _, principal := range firstDegreeAdmins.Slice() {
				if principal.Kinds.ContainsOneOf(ad.Group) {
					ntlmCache.LocalGroupData.GroupMembershipCache.OrReach(principal.ID.Uint64(), graph.DirectionInbound, allAdminPrincipals)
				} else {
					allAdminPrincipals.Add(principal.ID.Uint64())
				}
			}

			// Get the cross between the admin group ids and all unprotected computers. This auto filters to computers only and filters out restricted outbound stuff/protected users
			allAdminPrincipals.And(ntlmCache.UnprotectedComputersCache)

			// Remove the target computer if it exists as self-relay is not possible
			allAdminPrincipals.Remove(computer.ID.Uint64())

			if allAdminPrincipals.Cardinality() > 0 {
				outC <- post.EnsureRelationshipJob{
					FromID: authenticatedUserID,
					ToID:   computer.ID,
					Kind:   ad.CoerceAndRelayNTLMToSMB,
				}
			}
		}
	}

	return nil
}

func GetVulnerableEnterpriseCAsForRelayNTLMtoADCS(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.NodeSet, error) {
	var (
		nodes = graph.NodeSet{}
	)

	if composition, err := GetCoerceAndRelayNTLMtoADCSEdgeComposition(ctx, db, edge); err != nil {
		return graph.NodeSet{}, err
	} else {
		for _, node := range composition.AllNodes().ContainingNodeKinds(ad.EnterpriseCA) {
			if vuln, err := node.Properties.Get(ad.HasVulnerableEndpoint.String()).Bool(); errors.Is(err, graph.ErrPropertyNotFound) {
				continue
			} else if err != nil {
				slog.ErrorContext(
					ctx,
					"Error getting hasvulnerableendpoint from node",
					slog.Uint64("node_id", uint64(node.ID)),
					attr.Error(err),
				)
			} else if vuln {
				nodes.Add(node)
			}
		}

		return nodes, nil
	}

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
		slog.WarnContext(
			ctx,
			"Error getting domain SID for domain",
			slog.Uint64("node_id", uint64(startNode.ID)),
			attr.Error(err),
		)
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
		slog.WarnContext(
			ctx,
			"Error getting domain SID for domain",
			slog.Uint64("node_id", uint64(startNode.ID)),
			attr.Error(err),
		)
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

func GetCoercionTargetsForCoerceAndRelayNTLMtoSMB(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.NodeSet, error) {
	var (
		startNode *graph.Node
		endNode   *graph.Node
		nodes     graph.NodeSet
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else if endNode, err = ops.FetchNode(tx, edge.EndID); err != nil {
			return err
		} else if domainsid, err := startNode.Properties.Get(ad.DomainSID.String()).String(); err != nil {
			slog.WarnContext(
				ctx,
				"Error getting domain SID for domain",
				slog.Uint64("node_id", uint64(startNode.ID)),
				attr.Error(err),
			)
			return err
		} else if innerNodes, err := ops.AcyclicTraverseTerminals(tx, ops.TraversalPlan{
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
				node := segment.Node
				if node.Kinds.ContainsOneOf(ad.User) {
					return false
				} else if nodeDomainSid, err := node.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					return false
				} else if nodeDomainSid != domainsid {
					return false
				} else if restrictNtlm, err := node.Properties.Get(ad.RestrictOutboundNTLM.String()).Bool(); err != nil || restrictNtlm {
					return false
				} else {
					return true
				}
			},
		}); err != nil {
			return err
		} else {
			innerNodes.Remove(endNode.ID)
			nodes = innerNodes
			return nil
		}
	}); err != nil {
		return nil, err
	} else {
		return nodes, nil
	}
}

// PostCoerceAndRelayNTLMToLDAP creates edges where an authenticated user group, for a given domain, is able to target the provided computer.
// This will create either a CoerceAndRelayNTLMToLDAP or CoerceAndRelayNTLMToLDAPS edges, depending on the ldapSigning property of the domain
func PostCoerceAndRelayNTLMToLDAP(outC chan<- post.EnsureRelationshipJob, computer *graph.Node, authenticatedUserGroupID graph.ID, ldapSigningCache map[string]LDAPSigningCache) error {
	// webclientrunning must be set to true for the computer's properties in order for this attack path to be viable
	// If the property is not found, we will assume false
	if webClientRunning, err := computer.Properties.Get(ad.WebClientRunning.String()).Bool(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
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
					outC <- post.EnsureRelationshipJob{
						FromID: authenticatedUserGroupID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToLDAP,
					}
				} else if len(signingCache.relayableToDCLDAP) > 1 {
					outC <- post.EnsureRelationshipJob{
						FromID: authenticatedUserGroupID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToLDAP,
					}
				}

				if len(signingCache.relayableToDCLDAPS) == 1 && signingCache.relayableToDCLDAPS[0] != computer.ID {
					outC <- post.EnsureRelationshipJob{
						FromID: authenticatedUserGroupID,
						ToID:   computer.ID,
						Kind:   ad.CoerceAndRelayNTLMToLDAPS,
					}
				} else if len(signingCache.relayableToDCLDAPS) > 1 {
					outC <- post.EnsureRelationshipJob{
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
			query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), wellknown.AuthenticatedUsersSIDSuffix.String())),
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
func FetchProtectedUsersMappedToDomains(ctx context.Context, db graph.Database, localGroupData *LocalGroupData) (map[string]cardinality.Duplex[uint64], error) {
	protectedUsers := make(map[string]cardinality.Duplex[uint64])

	err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Filter(
			query.And(
				query.Kind(query.Node(), ad.Group),
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), wellknown.ProtectedUsersSIDSuffix.String())),
		).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for protectedUserGroup := range cursor.Chan() {
				if domain, err := protectedUserGroup.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					continue
				} else {
					set := cardinality.NewBitmap64()
					localGroupData.GroupMembershipCache.OrReach(protectedUserGroup.ID.Uint64(), graph.DirectionInbound, set)
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
	IsVulnerableFunctionalLevel bool
	relayableToDCLDAP           []graph.ID
	relayableToDCLDAPS          []graph.ID
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
					isFunctionalLevelVulnerable := false
					if functionalLevel, err := domain.Properties.Get(ad.FunctionalLevel.String()).String(); err != nil && !errors.Is(err, graph.ErrPropertyNotFound) {
						return err
					} else if slices.Contains(vulnerableFunctionalLevels(), functionalLevel) {
						isFunctionalLevelVulnerable = true
					}

					cache[domainSid] = LDAPSigningCache{
						IsVulnerableFunctionalLevel: isFunctionalLevelVulnerable,
						relayableToDCLDAP:           relayableToDcLdap,
						relayableToDCLDAPS:          relayableToDcLdaps,
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
