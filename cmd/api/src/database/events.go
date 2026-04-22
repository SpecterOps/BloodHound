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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"gorm.io/gorm"
)

type EventData interface {
	CreateEvent(ctx context.Context, event model.Event) (model.Event, error)
	GetAllEvents(ctx context.Context, skip, limit int, order string, filter model.SQLFilter) (model.Events, int, error)
	GetEvent(ctx context.Context, id int) (model.Event, error)
}

func (s *BloodhoundDB) CreateEvent(ctx context.Context, event model.Event) (model.Event, error) {
	result := s.db.WithContext(ctx).Create(&event)
	return event, CheckError(result)
}

func (s *BloodhoundDB) GetAllEvents(ctx context.Context, skip, limit int, order string, filter model.SQLFilter) (model.Events, int, error) {
	var (
		events model.Events
		result *gorm.DB
		count  int64
	)

	if filter.SQLString != "" {
		result = s.db.Model(&events).WithContext(ctx).Where(filter.SQLString, filter.Params...).Count(&count)
	} else {
		result = s.db.Model(&events).WithContext(ctx).Count(&count)
	}

	if result.Error != nil {
		return nil, 0, CheckError(result)
	}

	if order == "" {
		order = "created_at desc"
	}

	cursor := s.Scope(Paginate(skip, limit)).WithContext(ctx)

	if filter.SQLString != "" {
		result = cursor.Where(filter.SQLString, filter.Params...).Order(order).Find(&events)
	} else {
		result = cursor.Order(order).Find(&events)
	}

	return events, int(count), CheckError(result)
}

func (s *BloodhoundDB) GetEvent(ctx context.Context, id int) (model.Event, error) {
	var event model.Event
	result := s.db.WithContext(ctx).First(&event, id)
	return event, CheckError(result)
}
