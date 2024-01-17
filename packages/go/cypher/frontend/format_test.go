// Copyright 2023 Specter Ops, Inc.
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

package frontend

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCypherEmitter_StripLiterals(t *testing.T) {
	var (
		buffer            = &bytes.Buffer{}
		regularQuery, err = ParseCypher(DefaultCypherContext(), "match (n {value: 'PII'}) where n.other = 'more pii' and n.number = 411 return n.name, n")
		emitter           = CypherEmitter{
			StripLiterals: true,
		}
	)

	require.Nil(t, err)
	require.Nil(t, emitter.Write(regularQuery, buffer))
	require.Equal(t, "match (n {value: $STRIPPED}) where n.other = $STRIPPED and n.number = $STRIPPED return n.name, n", buffer.String())
}
