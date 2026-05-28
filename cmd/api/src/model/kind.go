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

package model

import (
	"strings"

	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

// IsExtendedNodeKind reports whether a kind is an internal or framework kind
// that should not be treated as a user-ingested node kind.
func IsExtendedNodeKind(kind graph.Kind) bool {
	return strings.HasPrefix(kind.String(), AssetGroupTagKindPrefix) ||
		kind.Is(common.MigrationData, graphschema.Meta, graphschema.MetaDetail)
}

type Kind struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func (k Kind) TableName() string {
	return "kind"
}

func (k Kind) ToKind() graph.Kind {
	return graph.StringKind(k.Name)
}

func (k Kind) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"type": {Equals, NotEquals},
	}
}
