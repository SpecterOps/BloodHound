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
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/channels"
	"github.com/specterops/bloodhound/graphschema/ad"
)

var (
	ErrNoCertParent = errors.New("cert has no parent")
)

func PostIssuedSignedBy(ctx context.Context, db graph.Database, enterpriseCertAuthorities []graph.Node, rootCertAuthorities []graph.Node) (*analysis.AtomicPostProcessingStats, error) {
	operation := analysis.NewPostRelationshipOperation(ctx, db, "IssuedSignBy Post Processing")

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, node := range enterpriseCertAuthorities {
			if postRel, err := processCertChainParent(node, tx); err != nil && !errors.Is(err, ErrNoCertParent) {
				return err
			} else if errors.Is(err, ErrNoCertParent) {
				continue
			} else if !channels.Submit(ctx, outC, postRel) {
				return nil
			}
		}

		return nil
	})

	operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- analysis.CreatePostRelationshipJob) error {
		for _, node := range rootCertAuthorities {
			if postRel, err := processCertChainParent(node, tx); err != nil && !errors.Is(err, ErrNoCertParent) {
				return err
			} else if errors.Is(err, ErrNoCertParent) {
				continue
			} else if !channels.Submit(ctx, outC, postRel) {
				return nil
			}
		}

		return nil
	})

	return &operation.Stats, operation.Done()
}

func processCertChainParent(node graph.Node, tx graph.Transaction) (analysis.CreatePostRelationshipJob, error) {
	if certChain, err := node.Properties.Get(ad.CertChain.String()).StringSlice(); err != nil {
		return analysis.CreatePostRelationshipJob{}, err
	} else if len(certChain) > 1 {
		parentCert := certChain[1]
		if targetNode, err := findMatchingCertChainID(parentCert, tx); err != nil {
			return analysis.CreatePostRelationshipJob{}, err
		} else {
			return analysis.CreatePostRelationshipJob{
				FromID: node.ID,
				ToID:   targetNode,
				Kind:   ad.IssuedSignedBy,
			}, nil
		}
	} else {
		return analysis.CreatePostRelationshipJob{}, ErrNoCertParent
	}
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
