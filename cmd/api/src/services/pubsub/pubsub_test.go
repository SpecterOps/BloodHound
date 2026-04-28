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

func TestService_Publish(t *testing.T) {

	t.Run("error - empty message", func(t *testing.T) {
		var (
			mockCtrl     = gomock.NewController(t)
			mockDatabase = mocks.NewMockPubSubRepository(mockCtrl)
			service      = pubsub.NewPubSubService(mockDatabase)
			ctx          = context.Background()
			eventInput   = model.EventInput{Type: "ingest.started"}
		)

		defer mockCtrl.Finish()

		_, err := service.Publish(ctx, eventInput)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "event message cannot be empty")
	})

	t.Run("error - database create fails", func(t *testing.T) {
		var (
			mockCtrl     = gomock.NewController(t)
			mockDatabase = mocks.NewMockPubSubRepository(mockCtrl)
			service      = pubsub.NewPubSubService(mockDatabase)
			ctx          = context.Background()
			eventInput   = model.EventInput{Type: "ingest.complete", Message: "done"}
		)

		defer mockCtrl.Finish()

		mockDatabase.EXPECT().CreateEvent(ctx, gomock.Any()).Return(model.Event{}, fmt.Errorf("database connection lost"))

		_, err := service.Publish(ctx, eventInput)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error creating event: database connection lost")
	})
	t.Run("success - publishes event with all fields", func(t *testing.T) {
		var (
			mockCtrl     = gomock.NewController(t)
			mockDatabase = mocks.NewMockPubSubRepository(mockCtrl)
			service      = pubsub.NewPubSubService(mockDatabase)
			ctx          = context.Background()
			eventInput   = model.EventInput{Type: "analysis.completed", Message: "Analysis completed successfully", Data: types.JSONUntypedObject{"domain": "testlab.local"}}
		)

		defer mockCtrl.Finish()

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

		createdEvent, err := service.Publish(ctx, eventInput)
		require.NoError(t, err)
		assert.Equal(t, string(eventInput.Type), createdEvent.Type)
		assert.Equal(t, eventInput.Message, createdEvent.Message)
		assert.Equal(t, eventInput.Data, createdEvent.Data)
		assert.NotEmpty(t, createdEvent.ID)
	})

	t.Run("success - publishes event with minimal fields", func(t *testing.T) {
		var (
			mockCtrl     = gomock.NewController(t)
			mockDatabase = mocks.NewMockPubSubRepository(mockCtrl)
			service      = pubsub.NewPubSubService(mockDatabase)
			ctx          = context.Background()
			eventInput   = model.EventInput{Type: "ingest.started", Message: "Ingest Started Successfully"}
		)

		defer mockCtrl.Finish()

		mockDatabase.EXPECT().CreateEvent(ctx, gomock.Any()).DoAndReturn(
			func(_ context.Context, event model.Event) (model.Event, error) {
				assert.Equal(t, string(eventInput.Type), event.Type)
				assert.Equal(t, eventInput.Message, event.Message)
				assert.Nil(t, event.Data)
				return event, nil
			},
		)

		createdEvent, err := service.Publish(ctx, eventInput)
		require.NoError(t, err)
		assert.Equal(t, string(eventInput.Type), createdEvent.Type)
		assert.Equal(t, eventInput.Message, createdEvent.Message)
	})
}
