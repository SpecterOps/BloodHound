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

package visualization

import (
	"bytes"
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/drivers/pg/pgutil"

	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/cypher/models/pgsql/translate"
	"github.com/stretchr/testify/require"
)

func TestGraphToPUMLDigraph(t *testing.T) {
	kindMapper := pgutil.NewInMemoryKindMapper()

	regularQuery, err := frontend.ParseCypher(frontend.NewContext(), "match (s), (e) where s.name = s.other + 1 / s.last return s")
	require.Nil(t, err)

	translation, err := translate.Translate(context.Background(), regularQuery, kindMapper, nil)
	require.Nil(t, err)

	graph, err := SQLToDigraph(translation.Statement)
	require.Nil(t, err)

	require.Nil(t, err)
	require.Nil(t, GraphToPUMLDigraph(graph, &bytes.Buffer{}))
}
