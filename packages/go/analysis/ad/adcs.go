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
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
)

func PostIssuedSignedBy(ctx context.Context, db graph.Database, enterpriseCertAuthorities []graph.Node, rootCertAuthorities []graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "IssuedSignBy Post Processing")

	for _, node := range enterpriseCertAuthorities {
		innerNode := node
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			if certChain, err := node.Properties.Get(ad.CertChain.String()).StringSlice(); err != nil {
				return err
			} else if len(certChain) > 1 {
				parentCert := certChain[1]
				if targetNode, err := findMatchingCertChainID(parentCert, tx); err != nil {
					return err
				} else {
					if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
						FromID: innerNode.ID,
						ToID:   targetNode,
						Kind:   ad.IssuedSignedBy,
					}) {
						return nil
					}
				}
			}
			return nil
		})
	}

	for _, node := range rootCertAuthorities {
		innerNode := node
		operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
			if certChain, err := node.Properties.Get(ad.CertChain.String()).StringSlice(); err != nil {
				return err
			} else if len(certChain) > 1 {
				parentCert := certChain[1]
				if targetNode, err := findMatchingCertChainID(parentCert, tx); err != nil {
					return err
				} else {
					if !channels.Submit(ctx, outC, analysis.CreatePostRelationshipJob{
						FromID: innerNode.ID,
						ToID:   targetNode,
						Kind:   ad.IssuedSignedBy,
					}) {
						return nil
					}
				}
			}
			return nil
		})
	}

	return &operation.Stats, operation.Done()
}

func findMatchingCertChainID(certThumbprint string, tx graph.Transaction) (graph.ID, error) {
	if targetNode, err := tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), ad.Entity),
			query.Equals(
				query.NodeProperty(ad.CertThumbprint.String()),
				certThumbprint,
			),
		)
	}).First(); err != nil {
		return graph.ID(0), err
	} else {
		return targetNode.ID, nil
	}
}
