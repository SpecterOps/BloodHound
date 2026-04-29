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

package pubsub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/pubsub"
	"github.com/specterops/bloodhound/cmd/api/src/services/pubsub/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testHandler struct {
	called bool
}

func (s *testHandler) HandleEvent(_ context.Context, _ model.Event) error {
	s.called = true
	return nil
}

type errorHandler struct {
	called bool
	err    error
}

func (s *errorHandler) HandleEvent(_ context.Context, _ model.Event) error {
	s.called = true
	return s.err
}

func TestPubSubService_Publish(t *testing.T) {

	t.Run("error - empty message", func(t *testing.T) {
		var (
			mockCtrl      = gomock.NewController(t)
			mockDatabase  = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService = pubsub.NewPubSubService(mockDatabase)
			ctx           = context.Background()
			eventInput    = model.EventInput{Type: "ingest.started"}
		)

		defer mockCtrl.Finish()

		_, err := pubSubService.Publish(ctx, eventInput)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event message cannot be empty")
	})

	t.Run("error - database create fails", func(t *testing.T) {
		var (
			mockCtrl      = gomock.NewController(t)
			mockDatabase  = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService = pubsub.NewPubSubService(mockDatabase)
			ctx           = context.Background()
			eventInput    = model.EventInput{Type: "ingest.complete", Message: "done"}
		)

		defer mockCtrl.Finish()

		mockDatabase.EXPECT().CreateEvent(ctx, gomock.Any()).Return(model.Event{}, fmt.Errorf("database connection lost"))

		_, err := pubSubService.Publish(ctx, eventInput)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error creating event: database connection lost")
	})
	t.Run("success - publishes event with all fields and fires handler", func(t *testing.T) {
		var (
			mockCtrl      = gomock.NewController(t)
			mockDatabase  = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService = pubsub.NewPubSubService(mockDatabase)
			ctx           = context.Background()
			handler       = &testHandler{}
			eventInput    = model.EventInput{Type: "analysis.completed", Message: "Analysis completed successfully", Data: types.JSONUntypedObject{"domain": "testlab.local"}}
		)

		defer mockCtrl.Finish()

		pubSubService.Subscribe("analysis.completed", handler)

		mockDatabase.EXPECT().CreateEvent(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, event model.Event) (model.Event, error) {
				assert.Equal(t, string(eventInput.Type), event.Type)
				assert.Equal(t, eventInput.Message, event.Message)
				assert.Equal(t, eventInput.Data, event.Data)
				assert.NotEmpty(t, event.ID)
				assert.True(t, event.ProcessedAt.Time.IsZero(), "processed_at should not be set")
				return event, nil
			},
		)

		createdEvent, err := pubSubService.Publish(ctx, eventInput)
		require.NoError(t, err)
		assert.Equal(t, string(eventInput.Type), createdEvent.Type)
		assert.Equal(t, eventInput.Message, createdEvent.Message)
		assert.Equal(t, eventInput.Data, createdEvent.Data)
		assert.NotEmpty(t, createdEvent.ID)
		assert.True(t, handler.called, "handler should have been called")
	})

	t.Run("success - publishes event with minimal fields (no data) and fires handler", func(t *testing.T) {
		var (
			mockCtrl      = gomock.NewController(t)
			mockDatabase  = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService = pubsub.NewPubSubService(mockDatabase)
			ctx           = context.Background()
			handler       = &testHandler{}
			eventInput    = model.EventInput{Type: "ingest.started", Message: "Ingest Started Successfully"}
		)

		defer mockCtrl.Finish()

		pubSubService.Subscribe("ingest.started", handler)

		mockDatabase.EXPECT().CreateEvent(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, event model.Event) (model.Event, error) {
				assert.Equal(t, string(eventInput.Type), event.Type)
				assert.Equal(t, eventInput.Message, event.Message)
				assert.Nil(t, event.Data)
				return event, nil
			},
		)

		createdEvent, err := pubSubService.Publish(ctx, eventInput)
		require.NoError(t, err)
		assert.Equal(t, string(eventInput.Type), createdEvent.Type)
		assert.Equal(t, eventInput.Message, createdEvent.Message)
		assert.True(t, handler.called, "handler should have been called")
	})

	t.Run("success - fires multiple handlers for the same event type", func(t *testing.T) {
		var (
			mockCtrl      = gomock.NewController(t)
			mockDatabase  = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService = pubsub.NewPubSubService(mockDatabase)
			ctx           = context.Background()
			handlerOne    = &testHandler{}
			handlerTwo    = &testHandler{}
			eventInput    = model.EventInput{Type: "ingest.started", Message: "Ingest started"}
		)

		defer mockCtrl.Finish()

		pubSubService.Subscribe("ingest.started", handlerOne)
		pubSubService.Subscribe("ingest.started", handlerTwo)

		mockDatabase.EXPECT().CreateEvent(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, event model.Event) (model.Event, error) {
				return event, nil
			},
		)

		_, err := pubSubService.Publish(ctx, eventInput)
		require.NoError(t, err)
		assert.True(t, handlerOne.called, "the first handler should have been called")
		assert.True(t, handlerTwo.called, "the second handler should have been called")
	})

	t.Run("success - does not fire handler for a different event type", func(t *testing.T) {
		var (
			mockCtrl      = gomock.NewController(t)
			mockDatabase  = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService = pubsub.NewPubSubService(mockDatabase)
			ctx           = context.Background()
			handler       = &testHandler{}
			eventInput    = model.EventInput{Type: "ingest.started", Message: "Ingest started"}
		)

		defer mockCtrl.Finish()

		pubSubService.Subscribe("analysis.completed", handler)

		mockDatabase.EXPECT().CreateEvent(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, event model.Event) (model.Event, error) {
				return event, nil
			},
		)

		_, err := pubSubService.Publish(ctx, eventInput)
		require.NoError(t, err)
		assert.False(t, handler.called, "the handler should not have been called for a different event type")
	})

	t.Run("success - handler error does not prevent publish from succeeding", func(t *testing.T) {
		var (
			mockCtrl       = gomock.NewController(t)
			mockDatabase   = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService  = pubsub.NewPubSubService(mockDatabase)
			ctx            = context.Background()
			failingHandler = &errorHandler{err: fmt.Errorf("handler failed")}
			passingHandler = &testHandler{}
			eventInput     = model.EventInput{Type: "ingest.started", Message: "Ingest started"}
		)

		defer mockCtrl.Finish()

		pubSubService.Subscribe("ingest.started", failingHandler)
		pubSubService.Subscribe("ingest.started", passingHandler)

		mockDatabase.EXPECT().CreateEvent(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, event model.Event) (model.Event, error) {
				return event, nil
			},
		)

		createdEvent, err := pubSubService.Publish(ctx, eventInput)
		require.NoError(t, err, "publish should still succeed even if a handler errors")
		assert.NotEmpty(t, createdEvent.ID)
		assert.True(t, failingHandler.called, "the failing handler should have been called")
		assert.True(t, passingHandler.called, "the passing handler should still be called after a prior handler error")
	})
}

func TestPubSubService_Subscribe(t *testing.T) {
	t.Run("success - subscribes a single handler", func(t *testing.T) {
		var (
			mockCtrl      = gomock.NewController(t)
			mockDatabase  = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService = pubsub.NewPubSubService(mockDatabase)
		)

		defer mockCtrl.Finish()

		handler := &testHandler{}
		pubSubService.Subscribe("ingest.started", handler)
	})

	t.Run("success - subscribes multiple handlers for the same event type", func(t *testing.T) {
		var (
			mockCtrl      = gomock.NewController(t)
			mockDatabase  = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService = pubsub.NewPubSubService(mockDatabase)
		)

		defer mockCtrl.Finish()

		handlerOne := &testHandler{}
		handlerTwo := &testHandler{}
		pubSubService.Subscribe("ingest.started", handlerOne)
		pubSubService.Subscribe("ingest.started", handlerTwo)
	})

	t.Run("success - subscribes handlers for different event types", func(t *testing.T) {
		var (
			mockCtrl      = gomock.NewController(t)
			mockDatabase  = mocks.NewMockPubSubRepository(mockCtrl)
			pubSubService = pubsub.NewPubSubService(mockDatabase)
		)

		defer mockCtrl.Finish()

		handlerOne := &testHandler{}
		handlerTwo := &testHandler{}
		pubSubService.Subscribe("ingest.started", handlerOne)
		pubSubService.Subscribe("analysis.completed", handlerTwo)
	})
}
