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

package database

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type EventData interface {
	CreateEvent(ctx context.Context, event model.Event) (model.Event, error)
	GetEvents(ctx context.Context, filter model.SQLFilter, sort model.Sort, skip, limit int) (model.Events, int, error)
	GetEvent(ctx context.Context, eventId uuid.UUID) (model.Event, error)
}

func (s *BloodhoundDB) CreateEvent(ctx context.Context, event model.Event) (model.Event, error) {
	result := s.db.WithContext(ctx).Create(&event)
	return event, CheckError(result)
}

func (s *BloodhoundDB) GetEvents(ctx context.Context, filter model.SQLFilter, sort model.Sort, skip, limit int) (model.Events, int, error) {
	var (
		events          = model.Events{}
		skipLimitString string
		orderSQL        string
		sortColumns     []string
		totalRowCount   int
	)

	if filter.SQLString != "" {
		filter.SQLString = " WHERE " + filter.SQLString
	}

	if len(sort) == 0 {
		sort = append(sort, model.SortItem{Column: "created_at", Direction: model.DescendingSortDirection})
	}

	for _, item := range sort {
		dirString := "ASC"
		if item.Direction == model.DescendingSortDirection {
			dirString = "DESC"
		}
		sortColumns = append(sortColumns, fmt.Sprintf("%s %s", item.Column, dirString))
	}

	orderSQL = "ORDER BY " + strings.Join(sortColumns, ", ")

	if limit > 0 {
		skipLimitString += fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip > 0 {
		skipLimitString += fmt.Sprintf(" OFFSET %d", skip)
	}

	sqlString := fmt.Sprintf(`SELECT * FROM %s%s %s %s`, model.Event{}.TableName(), filter.SQLString, orderSQL, skipLimitString)

	if result := s.db.WithContext(ctx).Raw(sqlString, filter.Params...).Find(&events); result.Error != nil {
		return events, 0, CheckError(result)
	}

	if limit > 0 || skip > 0 {
		sqlCountStr := fmt.Sprintf(
			"SELECT COUNT(*) FROM %s%s",
			(model.Event{}).TableName(),
			filter.SQLString)
		if result := s.db.WithContext(ctx).Raw(sqlCountStr, filter.Params...).Scan(&totalRowCount); result.Error != nil {
			return []model.Event{}, 0, CheckError(result)
		}
	} else {
		totalRowCount = len(events)
	}

	return events, totalRowCount, nil
}

func (s *BloodhoundDB) GetEvent(ctx context.Context, eventId uuid.UUID) (model.Event, error) {
	var event model.Event
	result := s.db.WithContext(ctx).First(&event, eventId)
	return event, CheckError(result)
}
