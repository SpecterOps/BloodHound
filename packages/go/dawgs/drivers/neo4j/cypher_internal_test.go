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

package neo4j

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_StripCypher(t *testing.T) {
	var (
		query = "match (u1:User {domain: \"DOMAIN1\"}), (u2:User {domain: \"DOMAIN2\"}) where u1.samaccountname <> \"krbtgt\" and u1.samaccountname = u2.samaccountname with u2 match p1 = (u2)-[*1..]->(g:Group) with p1 match p2 = (u2)-[*1..]->(g:Group) return p1, p2"
	)

	result := stripCypherQuery(query)

	require.Equalf(t, false, strings.Contains(result, "DOMAIN1"), "Cypher query not sanitized. Contains sensitive value: %s", result)
	require.Equalf(t, false, strings.Contains(result, "DOMAIN2"), "Cypher query not sanitized. Contains sensitive value: %s", result)
}
