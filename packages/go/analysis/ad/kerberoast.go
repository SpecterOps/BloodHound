// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"log/slog"
	"strings"

	"github.com/specterops/bloodhound/packages/go/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
	"github.com/specterops/dawgs/util/channels"
)

const krbtgtSIDSuffix = "-502"

// isKerberoastableCandidate applies the account-level exclusions shared with
// the prebuilt "Kerberoastable users" searches: the krbgt account and group
// managed service accounts cannot be kerberoasted in a useful way.
func isKerberoastableCandidate(node *graph.Node) bool {
	var (
		objectID, objectIDErr = node.Properties.Get(common.ObjectID.String()).String()
		isGMSA, gmsaErr       = node.Properties.Get(ad.GMSA.String()).Bool()
		isMSA, msaErr         = node.Properties.Get(ad.MSA.String()).Bool()
	)

	if objectIDErr != nil || strings.HasSuffix(objectID, krbtgtSIDSuffix) {
		return false
	}

	if gmsaErr == nil && isGMSA {
		return false
	}

	if msaErr == nil && isMSA {
		return false
	}

	return true
}

// PostKerberoastable creates a Kerberoastable edge from the domain's
// "Authenticated Users" well-known group to every collected user account that
// exposes a service principal name, mirroring the fact that any authenticated
// domain principal may request a TGS ticket for such an account and attempt an
// offline crack of its service account secret.
func PostKerberoastable(ctx context.Context, db graph.Database) (*post.AtomicPostProcessingStats, error) {
	defer measure.ContextLogAndMeasure(
		ctx,
		slog.LevelInfo,
		"Post-processing Kerberoastable",
		attr.Namespace("analysis"),
		attr.Function("PostKerberoastable"),
		attr.Scope("process"),
	)()

	if domainNodes, err := fetchCollectedDomainNodes(ctx, db); err != nil {
		return &post.AtomicPostProcessingStats{}, err
	} else {
		operation := post.NewPostRelationshipOperation(ctx, db, "Kerberoastable Post Processing")

		for _, domain := range domainNodes {
			innerDomain := domain

			operation.Operation.SubmitReader(func(ctx context.Context, tx graph.Transaction, outC chan<- post.EnsureRelationshipJob) error {
				domainName, err := innerDomain.Properties.Get(common.Name.String()).String()
				if err != nil || domainName == "" {
					return nil
				}

				domainSid, err := innerDomain.Properties.Get(common.ObjectID.String()).String()
				if err != nil || domainSid == "" {
					return nil
				}

				authenticatedUsers, err := tx.Nodes().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Node(), ad.Group),
						query.Equals(query.NodeProperty(common.ObjectID.String()), wellknown.DefineSID(domainName, wellknown.AuthenticatedUsersSIDSuffix)),
					)
				}).First()
				if err != nil {
					if graph.IsErrNotFound(err) {
						return nil
					}
					return err
				}

				candidates, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
					return query.And(
						query.Kind(query.Node(), ad.User),
						query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSid),
						query.Equals(query.NodeProperty(ad.HasSPN.String()), true),
						query.Equals(query.NodeProperty(common.Enabled.String()), true),
					)
				}))
				if err != nil {
					return err
				}

				for _, candidate := range candidates {
					if !isKerberoastableCandidate(candidate) {
						continue
					}

					channels.Submit(ctx, outC, post.EnsureRelationshipJob{
						FromID: authenticatedUsers.ID,
						ToID:   candidate.ID,
						Kind:   ad.Kerberoastable,
					})
				}

				return nil
			})
		}

		return &operation.Stats, operation.Done()
	}
}
