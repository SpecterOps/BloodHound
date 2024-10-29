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
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
)

func PostOwner(ctx context.Context, db graph.Database, groupExpansions impact.Aggregator) (*analysis.AtomicPostProcessingStats, error) {
	if dsHeuristicsCache, err := GetDsHeuristicsCache(ctx, db); err != nil {
		return &analysis.AtomicPostProcessingStats{}, fmt.Errorf("failed fetching dsheuristics values: %w", err)
	} else {
		operation := analysis.NewPostRelationshipOperation(ctx, db, "Owns Post Processing")

	}
	return nil, nil
}

func GetDsHeuristicsCache(ctx context.Context, db graph.Database) (map[string]bool, error) {
	var (
		dsHeuristicValues = make(map[string]bool)
	)
	return dsHeuristicValues, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if domains, err := ops.FetchNodes(tx.Nodes().Filter(query.Kind(query.Node(), ad.Domain))); err != nil {
			return err
		} else {
			for _, domain := range domains {
				if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					continue
				} else if rawDsHeuristics, err := domain.Properties.Get(ad.DSHeuristics.String()).String(); err != nil {
					continue
				} else if len(rawDsHeuristics) < 29 {
					continue
				} else {
					enforcedChar := string(rawDsHeuristics[28])
					switch enforcedChar {
					case "0":
					case "2":
						dsHeuristicValues[domainSid] = false
					case "1":
						dsHeuristicValues[domainSid] = true
					default:
						continue
					}
				}
			}
		}

		return nil
	})
}
