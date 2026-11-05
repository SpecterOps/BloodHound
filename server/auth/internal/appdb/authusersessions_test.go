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
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedLogoutSQL is the literal SQL the Store issues for Logout.
const expectedLogoutSQL = `UPDATE user_sessions SET expires_at = current_timestamp, updated_at = current_timestamp WHERE id = $1`

func TestStore_Logout_SQL(t *testing.T) {
	var (
		ctx       = context.Background()
		sessionID = int64(42)
		execErr   = errors.New("connection reset")
	)

	testCases := []struct {
		name         string
		setupMock    func(pool pgxmock.PgxPoolIface)
		expectStrErr string
	}{
		{
			name: "Expires the session by id",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectExec(expectedLogoutSQL).WithArgs(sessionID).WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name: "Missing session is not an error",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectExec(expectedLogoutSQL).WithArgs(sessionID).WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
		},
		{
			name: "Exec error is wrapped",
			setupMock: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectExec(expectedLogoutSQL).WithArgs(sessionID).WillReturnError(execErr)
			},
			expectStrErr: "ending user session: " + execErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			testCase.setupMock(pool)

			err := store.Logout(ctx, sessionID)

			if testCase.expectStrErr != "" {
				assert.EqualError(t, err, testCase.expectStrErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
