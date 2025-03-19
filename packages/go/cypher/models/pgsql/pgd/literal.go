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

import "github.com/specterops/bloodhound/cypher/models/pgsql"

func ExpressionArrayLiteral(values ...pgsql.Expression) pgsql.ArrayLiteral {
	return pgsql.ArrayLiteral{
		Values: values,
	}
}

func IntLiteral[T int | int16 | int32 | int64](literal T) pgsql.Literal {
	var dataType = pgsql.UnknownDataType

	switch any(literal).(type) {
	case int:
		dataType = pgsql.Int
	case int16:
		dataType = pgsql.Int2
	case int32:
		dataType = pgsql.Int4
	case int64:
		dataType = pgsql.Int8
	}

	return pgsql.NewLiteral(literal, dataType)
}

func TextLiteral(literal string) pgsql.Literal {
	return pgsql.NewLiteral(literal, pgsql.Text)
}
