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

package appdb

import (
	"strings"
	"testing"

	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/stretchr/testify/require"
)

func TestBuildGraphExpansionQuery(t *testing.T) {
	t.Parallel()

	t.Run("outbound", func(t *testing.T) {
		t.Parallel()

		query, err := buildGraphExpansionQuery(42, services.GraphExpansionDirectionOutbound, []string{"CustomTraversable"}, 501)

		require.NoError(t, err)
		require.Equal(t, `MATCH (source)
WHERE ID(source) = 42
MATCH (source)-[r:ALL_ATTACK_PATHS|CustomTraversable]->(target)
RETURN source, r, target
ORDER BY ID(r)
LIMIT 501`, query)
	})

	t.Run("inbound", func(t *testing.T) {
		t.Parallel()

		query, err := buildGraphExpansionQuery(42, services.GraphExpansionDirectionInbound, nil, 501)

		require.NoError(t, err)
		require.Contains(t, query, "MATCH (target)-[r:ALL_ATTACK_PATHS]->(source)")
		require.Contains(t, query, "ORDER BY ID(r)")
	})

	t.Run("deduplicates relationship kinds", func(t *testing.T) {
		t.Parallel()

		query, err := buildGraphExpansionQuery(42, services.GraphExpansionDirectionOutbound, []string{"Custom", "Custom"}, 501)

		require.NoError(t, err)
		require.Equal(t, 1, strings.Count(query, "Custom"))
	})

	t.Run("rejects invalid relationship kinds", func(t *testing.T) {
		t.Parallel()

		_, err := buildGraphExpansionQuery(42, services.GraphExpansionDirectionOutbound, []string{"Unsafe-Kind"}, 501)

		require.ErrorIs(t, err, services.ErrInvalidGraphExpansionRelationshipKind)
		require.Contains(t, err.Error(), "Unsafe-Kind")
	})

	t.Run("rejects invalid direction", func(t *testing.T) {
		t.Parallel()

		_, err := buildGraphExpansionQuery(42, "sideways", nil, 501)

		require.ErrorIs(t, err, services.ErrInvalidGraphExpansionDirection)
	})
}
