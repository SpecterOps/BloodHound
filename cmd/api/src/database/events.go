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

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type EventData interface {
	CreateEvent(ctx context.Context, event model.Event) (model.Event, error)
	GetEvent(ctx context.Context, eventId uuid.UUID) (model.Event, error)
}

func (s *BloodhoundDB) CreateEvent(ctx context.Context, event model.Event) (model.Event, error) {
	result := s.db.WithContext(ctx).Create(&event)
	return event, CheckError(result)
}

func (s *BloodhoundDB) GetEvent(ctx context.Context, eventId uuid.UUID) (model.Event, error) {
	var event model.Event
	result := s.db.WithContext(ctx).First(&event, eventId)
	return event, CheckError(result)
}
