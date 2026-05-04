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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/mock.go -package=mocks . PubSubRepository
package pubsub

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
)

type PubSubRepository interface {
	CreateEvent(ctx context.Context, event model.Event) (model.Event, error)
}

type EventHandler interface {
	HandleEvent(ctx context.Context) error
}

type PubSubService struct {
	database PubSubRepository
	handlers []EventHandler
}

func NewPubSubService(database PubSubRepository) *PubSubService {
	return &PubSubService{
		database: database,
		handlers: make([]EventHandler, 0),
	}
}

// Publish validates the event properties, creates a new event in the events table, and then fires an event to all handlers.
func (s *PubSubService) Publish(ctx context.Context, eventInput model.EventInput) (model.Event, error) {
	if eventID, err := uuid.NewV7(); err != nil {
		return model.Event{}, fmt.Errorf("error generating event ID: %w", err)
	} else if eventInput.Message == "" {
		return model.Event{}, errors.New("event message cannot be empty")
	} else {
		event := model.Event{
			ID:      eventID,
			Type:    string(eventInput.Type),
			Message: eventInput.Message,
			Data:    eventInput.Data,
		}

		if createdEvent, err := s.database.CreateEvent(ctx, event); err != nil {
			return model.Event{}, fmt.Errorf("error creating event: %w", err)
		} else {
			for _, eventHandler := range s.handlers {
				if err := eventHandler.HandleEvent(ctx); err != nil {
					slog.Error("Error handling event", attr.Error(err))
				}
			}
			return createdEvent, nil
		}
	}
}

// Subscribe registers an EventHandler.
func (s *PubSubService) Subscribe(handler EventHandler) {
	s.handlers = append(s.handlers, handler)
}
