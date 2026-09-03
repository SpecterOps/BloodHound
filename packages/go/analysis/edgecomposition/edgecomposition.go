// Copyright 2026 Specter Ops, Inc.
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

package edgecomposition

import (
	"context"

	"github.com/specterops/bloodhound/packages/go/analysis/ad"
	analysisAzure "github.com/specterops/bloodhound/packages/go/analysis/azure"
	"github.com/specterops/bloodhound/packages/go/analysis/hybrid"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
)

func GetEdgeCompositionPath(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	if edge == nil {
		return ad.GetEdgeCompositionPath(ctx, db, edge)
	}

	switch edge.Kind {
	case azure.ManageEntraDS:
		return analysisAzure.GetManageEntraDSEdgeComposition(ctx, db, edge)
	case azure.AddEntraDSGroupMember:
		return hybrid.GetAddEntraDSGroupMemberEdgeComposition(ctx, db, edge)
	case azure.ManageEntraDSSync:
		return hybrid.GetManageEntraDSSyncEdgeComposition(ctx, db, edge)
	default:
		return ad.GetEdgeCompositionPath(ctx, db, edge)
	}
}
