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
	"strings"

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/impact"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/slicesext"
	"github.com/specterops/dawgs/cardinality"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

func PostTrustedForNTAuth(ctx context.Context, db graph.Database, operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob]) error {
	if ntAuthStoreNodes, err := FetchNodesByKind(ctx, db, ad.NTAuthStore); err != nil {
		return err
	} else {
		for _, node := range ntAuthStoreNodes {
			innerNode := node

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
				if thumbprints, err := innerNode.Properties.Get(ad.CertThumbprints.String()).StringSlice(); err != nil {
					if strings.Contains(err.Error(), graph.ErrPropertyNotFound.Error()) {
						slog.WarnContext(ctx, fmt.Sprintf("Unable to post-process TrustedForNTAuth edge for NTAuthStore node %d due to missing adcs data: %v", innerNode.ID, err))
						return nil
					}
					return err
				} else {
					for _, thumbprint := range thumbprints {
						if thumbprint != "" {
							if sourceNodeIDs, err := findNodesByCertThumbprint(thumbprint, tx, ad.EnterpriseCA); err != nil {
								return err
							} else {
								for _, sourceNodeID := range sourceNodeIDs {
									if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
										FromID: sourceNodeID,
										ToID:   innerNode.ID,
										Kind:   ad.TrustedForNTAuth,
									}) {
										return nil
									}
								}
							}
						}
					}
				}
				return nil
			})
		}
	}

	return nil
}

func PostIssuedSignedBy(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node, rootCertAuthorities []*graph.Node, aiaCertAuthorities []*graph.Node) error {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, node := range enterpriseCertAuthorities {
			if postRels, err := processCertChainParent(node, tx); err != nil && !errors.Is(err, ErrNoCertParent) {
				return err
			} else if errors.Is(err, ErrNoCertParent) {
				continue
			} else {
				for _, rel := range postRels {
					if !channels.Submit(ctx, outC, rel) {
						return nil
					}
				}
			}
		}

		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, node := range rootCertAuthorities {
			if postRels, err := processCertChainParent(node, tx); err != nil && !errors.Is(err, ErrNoCertParent) {
				return err
			} else if errors.Is(err, ErrNoCertParent) {
				continue
			} else {
				for _, rel := range postRels {
					if !channels.Submit(ctx, outC, rel) {
						return nil
					}
				}
			}
		}

		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, node := range aiaCertAuthorities {
			if postRels, err := processCertChainParent(node, tx); err != nil && !errors.Is(err, ErrNoCertParent) {
				return err
			} else if errors.Is(err, ErrNoCertParent) {
				continue
			} else {
				for _, rel := range postRels {
					if !channels.Submit(ctx, outC, rel) {
						return nil
					}
				}
			}
		}

		return nil
	})

	return nil
}

func PostEnterpriseCAFor(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node) error {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, ecaNode := range enterpriseCertAuthorities {
			if thumbprint, err := ecaNode.Properties.Get(ad.CertThumbprint.String()).String(); err != nil {
				if graph.IsErrPropertyNotFound(err) {
					continue
				}
				return err
			} else if thumbprint != "" {
				if rootCAIDs, err := findNodesByCertThumbprint(thumbprint, tx, ad.RootCA); err != nil {
					return err
				} else {
					for _, rootCANodeID := range rootCAIDs {
						if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
							FromID: ecaNode.ID,
							ToID:   rootCANodeID,
							Kind:   ad.EnterpriseCAFor,
						}) {
							return fmt.Errorf("context timed out while creating EnterpriseCAFor edge")
						}
					}
				}
				if aiaCAIDs, err := findNodesByCertThumbprint(thumbprint, tx, ad.AIACA); err != nil {
					return err
				} else {
					for _, aiaCANodeID := range aiaCAIDs {
						if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
							FromID: ecaNode.ID,
							ToID:   aiaCANodeID,
							Kind:   ad.EnterpriseCAFor,
						}) {
							return fmt.Errorf("context timed out while creating EnterpriseCAFor edge")
						}
					}
				}
			}
		}
		return nil
	})
	return nil
}

func PostGoldenCert(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob, enterpriseCA *graph.Node, targetDomains *graph.NodeSet) error {
	if hostCAServiceComputers, err := FetchHostsCAServiceComputers(tx, enterpriseCA); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error fetching host ca computer for enterprise ca %d: %v", enterpriseCA.ID, err))
	} else {
		for _, computer := range hostCAServiceComputers {
			for _, domain := range targetDomains.Slice() {
				channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
					FromID: computer.ID,
					ToID:   domain.ID,
					Kind:   ad.GoldenCert,
				})
			}
		}
	}
	return nil
}

func PostExtendedByPolicyBinding(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], certTemplates []*graph.Node) error {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		if allIssuancePolicies, err := fetchAllIssuancePolicies(tx); err != nil {
			return err
		} else {
			// Get an O(1) lookup of Issuance Policies Required keyed by CertificatePolicyOID
			certTemplateOIDToIssuancePolicyMap := getIssuancePolicyCertOIDMap(allIssuancePolicies)

			// For each certTemplate, find all issuance policies within its CertificatePolicy property array
			// such that IssuancePolicy.CertificatePolicyOID is in CertificateTemplate.CertificatePolicy
			// and shares its domain
			for _, certTemplate := range certTemplates {
				if certPolicies, err := certTemplate.Properties.Get(ad.CertificatePolicy.String()).StringSlice(); err != nil {
					continue
				} else {
					for _, policy := range certPolicies {
						for _, issuancePolicy := range certTemplateOIDToIssuancePolicyMap[policy] {
							if certTemplateDomain, err := certTemplate.Properties.Get(ad.DomainSID.String()).String(); err != nil {
								continue
							} else if issuancePolicyDomain, err := issuancePolicy.Properties.Get(ad.DomainSID.String()).String(); err != nil {
								continue
							} else if certTemplateDomain != "" && certTemplateDomain == issuancePolicyDomain {
								// Create ExtendedByPolicy edge
								if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
									FromID: certTemplate.ID,
									ToID:   issuancePolicy.ID,
									Kind:   ad.ExtendedByPolicy,
								}) {
									return fmt.Errorf("context timed out while creating ExtendedByPolicy edge")
								}
							}
						}
					}
				}
			}
			return nil
		}
	})
	return nil
}

func fetchAllIssuancePolicies(tx graph.Transaction) (graph.NodeSet, error) {
	if nodes, err := ops.FetchNodes(tx.Nodes().Filterf(
		func() graph.Criteria {
			return query.And(
				query.Kind(query.Node(), ad.IssuancePolicy),
			)
		},
	)); err != nil {
		return nil, err
	} else {
		set := make(graph.NodeSet)
		set.Add(nodes...)
		return set, nil
	}
}

func getIssuancePolicyCertOIDMap(issuancePolicies graph.NodeSet) map[string][]graph.Node {
	oidMap := make(map[string][]graph.Node)

	for _, policy := range issuancePolicies {
		if certTemplateOID, err := policy.Properties.Get(ad.CertTemplateOID.String()).String(); err != nil {
			continue
		} else {
			oidMap[certTemplateOID] = append(oidMap[certTemplateOID], *policy)
		}
	}

	return oidMap
}

func processCertChainParent(node *graph.Node, tx graph.Transaction) ([]analysis.CreatePostRelationshipJob, error) {
	if certChain, err := node.Properties.Get(ad.CertChain.String()).StringSlice(); err != nil {
		if errors.Is(err, graph.ErrPropertyNotFound) {
			return []analysis.CreatePostRelationshipJob{}, nil
		}
		return []analysis.CreatePostRelationshipJob{}, err
	} else if len(certChain) > 1 {
		parentCert := certChain[1]
		if targetNodes, err := findNodesByCertThumbprint(parentCert, tx, ad.EnterpriseCA, ad.RootCA, ad.AIACA); err != nil {
			return []analysis.CreatePostRelationshipJob{}, err
		} else {
			return slicesext.Map(targetNodes, func(nodeId graph.ID) analysis.CreatePostRelationshipJob {
				return analysis.CreatePostRelationshipJob{
					FromID: node.ID,
					ToID:   nodeId,
					Kind:   ad.IssuedSignedBy,
				}
			}), nil
		}
	} else {
		return []analysis.CreatePostRelationshipJob{}, ErrNoCertParent
	}
}

func findNodesByCertThumbprint(certThumbprint string, tx graph.Transaction, kinds ...graph.Kind) ([]graph.ID, error) {
	return ops.FetchNodeIDs(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), kinds...),
			query.Equals(
				query.NodeProperty(ad.CertThumbprint.String()),
				certThumbprint,
			),
		)
	}))
}

func expandNodeSliceToBitmapWithoutGroups(nodes []*graph.Node, groupExpansions impact.PathAggregator) cardinality.Duplex[uint64] {
	var bitmap = cardinality.NewBitmap64()

	for _, controller := range nodes {
		if controller.Kinds.ContainsOneOf(ad.Group) {
			groupExpansions.Cardinality(controller.ID.Uint64()).(cardinality.Duplex[uint64]).Each(func(id uint64) bool {
				//Check group expansions against each id, if cardinality is 0 than its not a group
				if groupExpansions.Cardinality(id).Cardinality() == 0 {
					bitmap.Add(id)
				}
				return true
			})
		} else {
			bitmap.Add(controller.ID.Uint64())
		}
	}

	return bitmap
}

func containsAuthUsersOrEveryone(tx graph.Transaction, nodes []*graph.Node) (bool, error) {
	if specialGroups, err := FetchAuthUsersAndEveryoneGroups(tx); err != nil {
		return false, err
	} else {
		for _, node := range nodes {
			if specialGroups.Contains(node) {
				return true, nil
			} else if node.Kinds.ContainsOneOf(ad.Group) {
				for _, group := range specialGroups {
					if path, err := ops.FetchRelationships(tx.Relationships().Filterf(func() graph.Criteria {
						return query.And(
							query.Equals(query.StartID(), group.ID),
							query.KindIn(query.Relationship(), ad.MemberOf),
							query.Equals(query.EndID(), node.ID),
						)
					})); err != nil {
						return false, err
					} else if len(path) > 0 {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
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

func filterUserDNSResults(tx graph.Transaction, bitmap cardinality.Duplex[uint64], certTemplate *graph.Node) (cardinality.Duplex[uint64], error) {
	if userNodes, err := ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), ad.User),
			query.InIDs(query.NodeID(), graph.DuplexToGraphIDs(bitmap)...),
		)
	})); err != nil {
		if !graph.IsErrNotFound(err) {
			return nil, err
		}
	} else if len(userNodes) > 0 && !certTemplateValidForUserVictim(certTemplate) {
		bitmap.Xor(graph.NodeSetToDuplex(userNodes))
	}

	return bitmap, nil
}

func getVictimBitmap(groupExpansions impact.PathAggregator, certTemplateControllers, ecaControllers []*graph.Node, specialGroupHasTemplateEnroll, specialGroupHasECAEnroll bool) cardinality.Duplex[uint64] {
	// Expand controllers for the eca + template completely because we don't do group shortcutting here
	var (
		templateBitmap = expandNodeSliceToBitmapWithoutGroups(certTemplateControllers, groupExpansions)
		ecaBitmap      = expandNodeSliceToBitmapWithoutGroups(ecaControllers, groupExpansions)
		victimBitmap   = cardinality.NewBitmap64()
	)

	// If no special group has enroll neither the template or eca then return the common nodes among the enrollers
	if !specialGroupHasTemplateEnroll && !specialGroupHasECAEnroll {
		templateBitmap.And(ecaBitmap)
		return templateBitmap
	}

	// If a special group has enroll on the template then all enrollers of the eca can be a victim
	if specialGroupHasTemplateEnroll {
		victimBitmap.Or(ecaBitmap)
	}

	// If a special group has enroll on the eca then all enrollers of the template can be a victim
	if specialGroupHasECAEnroll {
		victimBitmap.Or(templateBitmap)
	}

	return victimBitmap
}

func schannelAuthenticationEnabled(certTemplate *graph.Node) (bool, error) {
	schannelAuthenticationEnabledExist := certTemplate.Properties.Exists(ad.SchannelAuthenticationEnabled.String())

	if schannelAuthenticationEnabledExist {
		return certTemplate.Properties.Get(ad.SchannelAuthenticationEnabled.String()).Bool()
	} else {
		// Fallback to EffectiveEKUs property
		if effectiveekus, err2 := certTemplate.Properties.Get(ad.EffectiveEKUs.String()).StringSlice(); err2 != nil {
			return false, err2
		} else {
			return slices.Contains(effectiveekus, "1.3.6.1.5.5.7.3.2") || slices.Contains(effectiveekus, "2.5.29.37.0") || len(effectiveekus) == 0, nil
		}
	}
}
