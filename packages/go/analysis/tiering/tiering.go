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
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

const (
	StrTagTierZero = "Tag_Tier_Zero"
	StrTagOwned    = "Tag_Owned"
)

var (
	KindTagTierZero = graph.StringKind(StrTagTierZero)
	KindTagOwned    = graph.StringKind(StrTagOwned)
)

func IsTierZero(node *graph.Node) bool {
	if node.Kinds.ContainsOneOf(KindTagTierZero) {
		return true
	} else {
		// We can safely ignore the error here
		startSystemTags, _ := node.Properties.Get(common.SystemTags.String()).String()
		return strings.Contains(startSystemTags, ad.AdminTierZero)
	}
}

func IsOwned(node *graph.Node) bool {
	if node.Kinds.ContainsOneOf(KindTagOwned) {
		return true
	} else {
		// We can safely ignore the error here
		startSystemTags, _ := node.Properties.Get(common.SystemTags.String()).String()
		return strings.Contains(startSystemTags, ad.Owned)
	}
}
