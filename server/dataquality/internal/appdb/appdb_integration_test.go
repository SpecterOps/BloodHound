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
	"time"

	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/dataquality/internal/appdb"
	"github.com/specterops/bloodhound/server/dataquality/internal/services"
	"github.com/specterops/bloodhound/server/internal/servertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAggregationsIntegration(t *testing.T) {
	var (
		ctx         = context.Background()
		database    = servertest.NewDatabase(t)
		pool        = database.Pool()
		store       = appdb.NewStore(pool)
		timestamp   = time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		kindID      int32
		otherKindID int32
	)
	require.NoError(t, pool.QueryRow(ctx, `INSERT INTO kind (name) VALUES ('AggregationEnvironment') RETURNING id`).Scan(&kindID))
	require.NoError(t, pool.QueryRow(ctx, `INSERT INTO kind (name) VALUES ('OtherAggregationEnvironment') RETURNING id`).Scan(&otherKindID))
	// There is no slice write endpoint: seed the persistence-only inputs directly.
	_, err := pool.Exec(ctx, `INSERT INTO data_quality_aggregations (run_id, schema_extension_id, schema_environment_kind_id, metric_type, metric_name, metric_value, kind_id, created_at, updated_at, deleted_at) VALUES
 ('first',42,$1,'count','nodes',1.25,NULL,$3,$3,NULL),
 ('second',42,$1,'count','nodes',2.5,$1,$3,$3,NULL),
 ('other-extension',43,$1,'count','nodes',3,NULL,$3,$3,NULL),
 ('other-kind',42,$2,'count','nodes',4,NULL,$3,$3,NULL),
 ('older',42,$1,'count','nodes',5,NULL,$3 - interval '1 second',$3,NULL),
 ('newer',42,$1,'count','nodes',6,NULL,$3 + interval '1 second',$3,NULL),
 ('deleted',42,$1,'count','nodes',7,NULL,$3,$3,$3)`, kindID, otherKindID, timestamp)
	require.NoError(t, err)
	extensionID := int32(42)
	testCases := []struct {
		name        string
		extensionID *int32
		sort        params.SortItems
		skip        int
		limit       int
		expected    []string
		count       int
	}{
		{name: "inclusive bounds exclude deleted and other kinds", expected: []string{"first", "second", "other-extension"}, count: 3},
		{name: "extension filter and stable tied timestamps", extensionID: &extensionID, limit: 1, skip: 1, expected: []string{"second"}, count: 2},
		{name: "offset without limit", extensionID: &extensionID, skip: 1, expected: []string{"second"}, count: 2},
		{name: "page beyond end retains total count", limit: 1, skip: 99, count: 3},
		{name: "updated at sorting keeps id tiebreaker", extensionID: &extensionID, sort: params.SortItems{{Field: "updated_at", Direction: params.Ascending}}, expected: []string{"first", "second"}, count: 2},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			records, count, err := store.ListAggregations(ctx, services.AggregationQuery{
				EnvironmentKindID: kindID, ExtensionID: testCase.extensionID, Start: timestamp, End: timestamp,
				Sort: testCase.sort, Skip: testCase.skip, Limit: testCase.limit,
			})
			require.NoError(t, err)
			assert.Equal(t, testCase.count, count)
			require.Len(t, records, len(testCase.expected))
			for index, runID := range testCase.expected {
				assert.Equal(t, runID, records[index].RunID)
			}
			if len(records) > 0 && records[0].RunID == "first" {
				assert.Equal(t, 1.25, records[0].MetricValue)
				assert.Nil(t, records[0].KindID)
			}
			if len(records) > 0 && records[0].RunID == "second" {
				require.NotNil(t, records[0].KindID)
				assert.Equal(t, kindID, *records[0].KindID)
			}
		})
	}
}
