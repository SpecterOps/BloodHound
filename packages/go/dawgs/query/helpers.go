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

package query

import (
	"github.com/specterops/bloodhound/dawgs/graph"
)

// TODO Cleanup after Tiering GA
func SearchTierZeroNodes(tieringEnabled bool) graph.Criteria {
	if tieringEnabled {
		return Kind(Node(), graph.StringKind("Tag_Tier_Zero"))
	} else {
		return StringContains(NodeProperty("system_tags"), "admin_tier_0")
	}
}

// TODO Cleanup after Tiering GA
func SearchTierZeroNodesRel(tieringEnabled bool) graph.Criteria {
	if tieringEnabled {
		return Kind(Start(), graph.StringKind("Tag_Tier_Zero"))
	} else {
		return StringContains(StartProperty("system_tags"), "admin_tier_0")
	}
}
