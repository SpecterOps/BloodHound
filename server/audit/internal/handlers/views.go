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
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

// LogView preserves the existing GET /api/v2/audit response shape.
type LogView struct {
	ID              int64          `json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	ActorID         string         `json:"actor_id"`
	ActorName       string         `json:"actor_name"`
	ActorEmail      string         `json:"actor_email"`
	Action          string         `json:"action"`
	Fields          map[string]any `json:"fields"`
	RequestID       string         `json:"request_id"`
	SourceIPAddress string         `json:"source_ip_address"`
	Status          string         `json:"status"`
	CommitID        uuid.UUID      `json:"commit_id"`
}

// LogListView wraps audit logs under the existing data.logs key.
type LogListView struct {
	Logs []LogView `json:"logs"`
}

func buildLogListView(logs []services.Log) LogListView {
	var views = make([]LogView, 0, len(logs))
	for _, log := range logs {
		views = append(views, LogView{
			ID:              log.ID,
			CreatedAt:       log.CreatedAt,
			ActorID:         log.ActorID,
			ActorName:       log.ActorName,
			ActorEmail:      log.ActorEmail,
			Action:          log.Action,
			Fields:          log.Fields,
			RequestID:       log.RequestID,
			SourceIPAddress: log.SourceIPAddress,
			Status:          log.Status,
			CommitID:        log.CommitID,
		})
	}

	return LogListView{Logs: views}
}

// JSONView implements responses.JSONViewer.
func (s LogListView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}

// ValidFilters describes the legacy audit endpoint's filter contract.
func (s LogListView) ValidFilters() map[string]params.FilterableField {
	var comparisonOperators = []params.FilterOperator{
		params.Equals,
		params.NotEquals,
		params.GreaterThan,
		params.GreaterThanOrEquals,
		params.LessThan,
		params.LessThanOrEquals,
	}
	var equalityOperators = []params.FilterOperator{params.Equals, params.NotEquals}

	return map[string]params.FilterableField{
		"id":                {Operators: comparisonOperators},
		"actor_id":          {Operators: equalityOperators, IsStringData: true},
		"actor_name":        {Operators: equalityOperators, IsStringData: true},
		"actor_email":       {Operators: equalityOperators, IsStringData: true},
		"action":            {Operators: equalityOperators, IsStringData: true},
		"request_id":        {Operators: equalityOperators, IsStringData: true},
		"created_at":        {Operators: comparisonOperators},
		"source_ip_address": {Operators: equalityOperators, IsStringData: true},
		"status":            {Operators: equalityOperators, IsStringData: true},
	}
}

// IsSortable describes the legacy audit endpoint's sort contract.
func (s LogListView) IsSortable(field string) bool {
	switch field {
	case "id", "actor_name", "actor_email", "action", "created_at", "source_ip_address", "status":
		return true
	default:
		return false
	}
}
