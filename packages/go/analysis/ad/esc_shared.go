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
	"strings"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/slicesext"
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
						log.Warnf("unable to post-process TrustedForNTAuth edge for NTAuthStore node %d due to missing adcs data: %v", innerNode.ID, err)
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

func PostIssuedSignedBy(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node, rootCertAuthorities []*graph.Node) error {
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

	return nil
}

func PostEnterpriseCAFor(operation analysis.StatTrackedOperation[analysis.CreatePostRelationshipJob], enterpriseCertAuthorities []*graph.Node) error {
	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, ecaNode := range enterpriseCertAuthorities {
			if thumbprint, err := ecaNode.Properties.Get(ad.CertThumbprint.String()).String(); err != nil {
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
			}
		}
		return nil
	})
	return nil
}

func processCertChainParent(node *graph.Node, tx graph.Transaction) ([]analysis.CreatePostRelationshipJob, error) {
	if certChain, err := node.Properties.Get(ad.CertChain.String()).StringSlice(); err != nil {
		if errors.Is(err, graph.ErrPropertyNotFound) {
			return []analysis.CreatePostRelationshipJob{}, nil
		}
		return []analysis.CreatePostRelationshipJob{}, err
	} else if len(certChain) > 1 {
		parentCert := certChain[1]
		if targetNodes, err := findNodesByCertThumbprint(parentCert, tx, ad.EnterpriseCA, ad.RootCA); err != nil {
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
