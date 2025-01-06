// Copyright 2024 Specter Ops, Inc.
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

package translate_test

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/dawgs/drivers/pg/pgutil"

	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/test"
	"github.com/specterops/bloodhound/dawgs/graph"
)

var (
	NodeKind1 = graph.StringKind("NodeKind1")
	NodeKind2 = graph.StringKind("NodeKind2")
	EdgeKind1 = graph.StringKind("EdgeKind1")
	EdgeKind2 = graph.StringKind("EdgeKind2")
)

func newKindMapper() pgsql.KindMapper {
	mapper := pgutil.NewInMemoryKindMapper()

	// This is here to make SQL output a little more predictable for test cases
	mapper.Put(NodeKind1)
	mapper.Put(NodeKind2)
	mapper.Put(EdgeKind1)
	mapper.Put(EdgeKind2)

	return mapper
}

func TestTranslate(t *testing.T) {
	var (
		casesRun   = 0
		kindMapper = newKindMapper()
	)

	if testCases, err := test.ReadTranslationTestCases(); err != nil {
		t.Fatal(err)
	} else {
		for _, testCase := range testCases {
			t.Run(testCase.Name, func(t *testing.T) {
				testCase.Assert(t, testCase.PgSQL, kindMapper)
			})

			casesRun += 1
		}
	}

	fmt.Printf("Ran %d test cases\n", casesRun)
}
