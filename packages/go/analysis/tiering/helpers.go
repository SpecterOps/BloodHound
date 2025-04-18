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
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func SearchTierZeroNodes(tieringEnabled bool) graph.Criteria {
	if tieringEnabled {
		return query.Kind(query.Node(), KindTagTierZero)
	} else {
		return query.StringContains(query.NodeProperty(common.SystemTags.String()), ad.AdminTierZero)
	}
}

func SearchTierZeroNodesRel(tieringEnabled bool) graph.Criteria {
	if tieringEnabled {
		return query.Kind(query.Start(), KindTagTierZero)
	} else {
		return query.StringContains(query.StartProperty(common.SystemTags.String()), ad.AdminTierZero)
	}
}
