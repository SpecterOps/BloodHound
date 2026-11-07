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

package appdb_test

import (
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/auth/internal/appdb"
	"github.com/stretchr/testify/require"
)

// newTestStore returns a Store backed by a pgxmock pool that compares SQL with
// pgxmock.QueryMatcherEqual. Unmet expectations fail the test during cleanup.
func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()

	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(func() {
		err := pool.ExpectationsWereMet()
		pool.Close()
		require.NoError(t, err)
	})

	return appdb.NewStore(pool), pool
}
