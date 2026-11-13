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

package database

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestBuildSQLFilterQualifiedColumns(t *testing.T) {
	for _, testCase := range []struct {
		column string
		want   string
	}{
		{column: "schema_extension_id", want: "schema_extension_id = 3"},
		{column: "se.schema_extension_id", want: "se.schema_extension_id = 3"},
		{column: "public.se.schema_extension_id", want: "public.se.schema_extension_id = 3"},
		{column: `se.quoted"column`, want: `se."quoted""column" = 3`},
	} {
		t.Run(testCase.column, func(t *testing.T) {
			filter, err := buildSQLFilter(model.Filters{
				testCase.column: {{Operator: model.Equals, Value: "3"}},
			})
			require.NoError(t, err)
			require.Equal(t, testCase.want, filter.sqlString)
		})
	}
}
