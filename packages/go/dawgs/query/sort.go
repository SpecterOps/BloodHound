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
	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type SortDirection string

const SortDirectionAscending SortDirection = "asc"
const SortDirectionDescending SortDirection = "desc"

type SortItem struct {
	SortCriteria graph.Criteria
	Direction    SortDirection
}

type SortItems []SortItem

func (s SortItems) FormatCypherOrder() *cypher.Order {
	var orderCriteria []graph.Criteria

	for _, sortItem := range s {
		switch sortItem.Direction {
		case SortDirectionAscending:
			orderCriteria = append(orderCriteria, Order(sortItem.SortCriteria, Ascending()))
		case SortDirectionDescending:
			orderCriteria = append(orderCriteria, Order(sortItem.SortCriteria, Descending()))
		}
	}
	return OrderBy(orderCriteria...)
}
