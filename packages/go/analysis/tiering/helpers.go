// Copyright 2025 Specter Ops, Inc.
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

package tiering

import (
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

type SearchTierNodesCriteria struct {
	PrimaryTierKind           graph.Kind
	SearchTierNodes           graph.Criteria
	SearchTierNodesRel        graph.Criteria
	SearchPrimaryTierNodes    graph.Criteria
	SearchPrimaryTierNodesRel graph.Criteria
}

func SearchTierNodes(tieringEnabled bool, tierKinds ...graph.Kind) graph.Criteria {
	if tieringEnabled {
		// Default to tier zero in the event no tierKinds are supplied
		if len(tierKinds) == 0 {
			tierKinds = append(tierKinds, KindTagTierZero)
		}
		return query.KindIn(query.Node(), tierKinds...)
	} else {
		return query.StringContains(query.NodeProperty(common.SystemTags.String()), ad.AdminTierZero)
	}
}

func SearchTierNodesRel(tieringEnabled bool, tierKinds ...graph.Kind) graph.Criteria {
	if tieringEnabled {
		// Default to tier zero in the event no tierKinds are supplied
		if len(tierKinds) == 0 {
			tierKinds = append(tierKinds, KindTagTierZero)
		}
		return query.KindIn(query.Start(), tierKinds...)
	} else {
		return query.StringContains(query.StartProperty(common.SystemTags.String()), ad.AdminTierZero)
	}
}
