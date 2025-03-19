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

package pgd

import (
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

func Any(expr pgsql.Expression, castType pgsql.DataType) *pgsql.AnyExpression {
	return pgsql.NewAnyExpression(expr, castType)
}

func Column(root, column pgsql.Identifier) pgsql.CompoundIdentifier {
	return pgsql.CompoundIdentifier{root, column}
}

func EntityID(identifier pgsql.Identifier) pgsql.CompoundIdentifier {
	return Column(identifier, pgsql.ColumnID)
}

func StartID(identifier pgsql.Identifier) pgsql.CompoundIdentifier {
	return Column(identifier, pgsql.ColumnStartID)
}

func EndID(identifier pgsql.Identifier) pgsql.CompoundIdentifier {
	return Column(identifier, pgsql.ColumnEndID)
}

func Properties(identifier pgsql.Identifier) pgsql.CompoundIdentifier {
	return Column(identifier, pgsql.ColumnProperties)
}
