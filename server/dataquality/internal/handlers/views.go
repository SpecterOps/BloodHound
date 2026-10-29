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

package handlers

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/specterops/bloodhound/server/dataquality/internal/services"
)

type AggregationView struct {
	RunID             string    `json:"run_id"`
	ExtensionID       int32     `json:"extension_id"`
	EnvironmentKindID int32     `json:"environment_kind_id"`
	MetricType        string    `json:"metric_type"`
	MetricName        string    `json:"metric_name"`
	MetricValue       float64   `json:"metric_value"`
	KindID            *int32    `json:"kind_id"`
	ID                int32     `json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	// The legacy Serial model exposes sql.NullTime even though deleted rows are excluded.
	DeletedAt sql.NullTime `json:"deleted_at"`
}

// AggregationsView owns the complete time-windowed, paginated response envelope.
type AggregationsView struct {
	Start *time.Time        `json:"start,omitempty"`
	End   *time.Time        `json:"end,omitempty"`
	Count int               `json:"count"`
	Limit int               `json:"limit"`
	Skip  int               `json:"skip"`
	Data  []AggregationView `json:"data"`
}

func BuildAggregationsView(aggregations []services.Aggregation, query services.AggregationQuery, count int) AggregationsView {
	var view = AggregationsView{Count: count, Limit: query.Limit, Skip: query.Skip}
	if !query.Start.IsZero() {
		view.Start = &query.Start
	}
	if !query.End.IsZero() {
		view.End = &query.End
	}
	for _, aggregation := range aggregations {
		view.Data = append(view.Data, AggregationView{
			ID: aggregation.ID, RunID: aggregation.RunID, ExtensionID: aggregation.SchemaExtensionID,
			EnvironmentKindID: aggregation.SchemaEnvironmentKindID, MetricType: aggregation.MetricType,
			MetricName: aggregation.MetricName, MetricValue: aggregation.MetricValue, KindID: aggregation.KindID,
			CreatedAt: aggregation.CreatedAt, UpdatedAt: aggregation.UpdatedAt,
		})
	}
	return view
}

func (s AggregationsView) IsSortable(field string) bool {
	return field == "created_at" || field == "updated_at"
}
func (s AggregationsView) JSONView() ([]byte, error) { return json.Marshal(s) }
