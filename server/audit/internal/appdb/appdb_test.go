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
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/audit/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_ListAuditLogs(t *testing.T) {
	t.Parallel()
	var (
		ctx      = context.Background()
		before   = time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
		after    = before.Add(-time.Hour)
		commitID = uuid.Must(uuid.NewV4())
		query    = services.ListQuery{
			Before: before, After: after, Skip: 3, Limit: 10,
			Filters: params.Filters{"actor_name": {{Field: "actor_name", Operator: params.Equals, Value: "Ada", SetOperator: params.FilterAnd}}},
			Sort:    params.SortItems{{Field: "created_at", Direction: params.Descending}},
		}
	)

	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	countBuilder, err := newAuditLogSelectBuilder(query)
	require.NoError(t, err)
	countBuilder.Select("COUNT(*) AS count")
	countSQL, countArgs := countBuilder.Build()
	pool.ExpectQuery(regexp.QuoteMeta(countSQL)).WithArgs(countArgs...).WillReturnRows(pool.NewRows([]string{"count"}).AddRow(int64(2)))

	listBuilder, err := newAuditLogSelectBuilder(query)
	require.NoError(t, err)
	listBuilder.Select(selectedAuditLogColumns...).OrderBy("created_at DESC").Offset(3).Limit(10)
	listSQL, listArgs := listBuilder.Build()
	pool.ExpectQuery(regexp.QuoteMeta(listSQL)).WithArgs(listArgs...).WillReturnRows(
		pool.NewRows(selectedAuditLogColumns).AddRow(
			int64(7), before.Add(-time.Minute), "actor-id", "Ada", "ada@example.com", "CreateUser",
			map[string]any{"name": "user"}, "request-id", "127.0.0.1", "success", commitID.String(),
		),
	)

	logs, count, err := NewStore(pool).ListAuditLogs(ctx, query)

	require.NoError(t, err)
	assert.Equal(t, 2, count)
	require.Len(t, logs, 1)
	assert.Equal(t, int64(7), logs[0].ID)
	assert.Equal(t, commitID, logs[0].CommitID)
	assert.Contains(t, countSQL, "FROM audit_logs_old")
	assert.Contains(t, countSQL, "FROM audit_logs")
	require.NoError(t, pool.ExpectationsWereMet())
}

func TestApplyFilters_RejectsUnknownFields(t *testing.T) {
	t.Parallel()
	_, err := newAuditLogSelectBuilder(services.ListQuery{
		Filters: params.Filters{"unknown": {{Field: "unknown", Operator: params.Equals, Value: "value"}}},
	})
	assert.ErrorContains(t, err, "unknown field")
}
