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

package model

import (
	"fmt"
	"time"

	"github.com/specterops/bloodhound/src/database/types"
)

type AuditLog struct {
	ID        int64                   `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time               `json:"created_at" gorm:"index"`
	ActorID   string                  `json:"actor_id" gorm:"index"`
	ActorName string                  `json:"actor_name"`
	Action    string                  `json:"action" gorm:"index"`
	Fields    types.JSONUntypedObject `json:"fields"`
	RequestID string                  `json:"request_id"`
}

func (s AuditLog) String() string {
	return fmt.Sprintf("actor %s %s executed action %s", s.ActorID, s.ActorName, s.Action)
}

type AuditLogs []AuditLog

func (s AuditLogs) IsSortable(column string) bool {
	switch column {
	case "id",
		"actor_id",
		"actor_name",
		"action",
		"request_id",
		"created_at":
		return true
	default:
		return false
	}
}

func (s AuditLogs) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"id":         {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"actor_id":   {Equals, NotEquals},
		"actor_name": {Equals, NotEquals},
		"action":     {Equals, NotEquals},
		"request_id": {Equals, NotEquals},
		"created_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
	}
}

func (s AuditLogs) IsString(column string) bool {
	switch column {
	case "actor_id",
		"actor_name",
		"action",
		"request_id":
		return true
	default:
		return false
	}
}

func (s AuditLogs) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s AuditLogs) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

type AuditData map[string]any

func (s AuditData) AuditData() AuditData {
	return s
}

// MergeLeft merges the given Auditable interface and merges its AuditData fields with the fields stored in this
// map. Conflicting keys are overwritten with the Auditable interface's AuditData fields.
func (s AuditData) MergeLeft(rightSide Auditable) AuditData {
	var (
		rightData = rightSide.AuditData()
		dest      = make(AuditData, len(s)+len(rightData))
	)

	for key, value := range s {
		dest[key] = value
	}

	for key, value := range rightData {
		dest[key] = value
	}

	return dest
}

type Auditable interface {
	AuditData() AuditData
}
