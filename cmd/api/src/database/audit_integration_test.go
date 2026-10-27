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

//go:build integration

package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestDatabase_ListAuditLogs(t *testing.T) {
	var (
		dbInst   = integration.SetupDB(t)
		after    = time.Now().Add(-time.Hour)
		midpoint time.Time
		before   time.Time

		auditLogIdFilter = model.QueryParameterFilter{
			Name:         "id",
			Operator:     model.GreaterThan,
			Value:        "2",
			IsStringData: false,
		}
		auditLogIdFilterMap = model.QueryParameterFilterMap{auditLogIdFilter.Name: model.QueryParameterFilters{auditLogIdFilter}}

		mockCtx = bhctx.Context{
			RequestID: "requestID",
			AuthCtx: auth.Context{
				Owner:   model.User{},
				Session: model.UserSession{},
			},
		}
		testCtx = bhctx.Set(context.Background(), &mockCtx)
	)

	appendAuditLogs := func(numLogs int) {
		for i := 0; i < numLogs; i++ {
			require.NoError(t, dbInst.AppendAuditLog(testCtx, model.AuditEntry{Model: &model.User{}, Action: model.AuditLogActionCreateUser, Status: model.AuditLogStatusSuccess}))
		}
	}

	// IDs 1-4 are created before the midpoint and IDs 5-7 are created after it
	appendAuditLogs(4)
	midpoint = time.Now()
	appendAuditLogs(3)
	before = time.Now()

	idFilter, err := auditLogIdFilterMap.BuildSQLFilter()
	require.NoError(t, err)

	testCases := []struct {
		name          string
		before        time.Time
		after         time.Time
		limit         int
		filter        model.SQLFilter
		expectedCount int
		expectedLogs  int
	}{
		{
			name:          "full window",
			before:        before,
			after:         after,
			limit:         10,
			expectedCount: 7,
			expectedLogs:  7,
		},
		{
			// Limit is set to 1 to verify that count is total filtered count, not response size
			name:          "full window with filter",
			before:        before,
			after:         after,
			limit:         1,
			filter:        idFilter,
			expectedCount: 5,
			expectedLogs:  1,
		},
		{
			name:          "partial window",
			before:        before,
			after:         midpoint,
			limit:         10,
			expectedCount: 3,
			expectedLogs:  3,
		},
		{
			name:          "partial window with filter",
			before:        midpoint,
			after:         after,
			limit:         1,
			filter:        idFilter,
			expectedCount: 2,
			expectedLogs:  1,
		},
		{
			name:          "empty window",
			before:        before.Add(2 * time.Hour),
			after:         before.Add(time.Hour),
			limit:         10,
			expectedCount: 0,
			expectedLogs:  0,
		},
		{
			name:          "empty window with filter",
			before:        before.Add(2 * time.Hour),
			after:         before.Add(time.Hour),
			limit:         1,
			filter:        idFilter,
			expectedCount: 0,
			expectedLogs:  0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			logs, count, err := dbInst.ListAuditLogs(testCtx, testCase.before, testCase.after, 0, testCase.limit, "", testCase.filter)
			require.NoError(t, err)
			require.Equal(t, testCase.expectedCount, count)
			require.Len(t, logs, testCase.expectedLogs)
		})
	}
}
