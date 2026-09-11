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

//go:build integration

package appdb_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// testSetupData carries the ids produced by seedNodeKind so table-driven cases can
// reference the seeded node kind (and its kind and owning extension) without re-querying.
type testSetupData struct {
	kindID      int32
	nodeKindID  int32
	extensionID int32
}

// seedNodeKind inserts a schema_extensions row, a kind row and a schema_node_kinds row
// wiring them together, registers a cascade-delete cleanup and returns the seeded ids.
func seedNodeKind(t *testing.T, ctx context.Context, pool *pgxpool.Pool) testSetupData {
	t.Helper()

	var extensionID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_extensions (name, display_name, version, is_builtin, namespace)
		VALUES ('TestExtension', 'Test Extension', '1.0.0', false, 'TST')
		RETURNING id`).Scan(&extensionID))

	var kindID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO kind (name) VALUES ('TestNodeKind') RETURNING id`).Scan(&kindID))

	var nodeKindID int32
	require.NoError(t, pool.QueryRow(ctx, `
		INSERT INTO schema_node_kinds (schema_extension_id, kind_id, display_name, description, is_display_kind, icon, icon_color)
		VALUES ($1, $2, 'Test Node Kind', 'a test node kind', true, 'user', '#fff')
		RETURNING id`, extensionID, kindID).Scan(&nodeKindID))

	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM schema_extensions WHERE name = 'TestExtension'`)
		_, _ = pool.Exec(ctx, `DELETE FROM kind WHERE name = 'TestNodeKind'`)
	})

	return testSetupData{kindID: kindID, nodeKindID: nodeKindID, extensionID: extensionID}
}
