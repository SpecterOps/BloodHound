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

package database_test

import (
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase_CreateEvent(t *testing.T) {
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		input model.Event
	}{
		{
			name: "success - create event with all fields",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			input: model.Event{
				ID:      uuid.Must(uuid.NewV7()),
				Type:    "analysis_complete",
				Message: "Analysis completed successfully",
				Data:    types.JSONUntypedObject{"domain": "testlab.local"},
			},
		},
		{
			name: "success - create event with minimal fields",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			input: model.Event{
				ID:   uuid.Must(uuid.NewV7()),
				Type: "ingest_started",
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			created, err := testSuite.BHDatabase.CreateEvent(testSuite.Context, testCase.input)

			require.NoError(t, err)
			assert.Equal(t, testCase.input.ID, created.ID)
			assert.Equal(t, testCase.input.Type, created.Type)
			assert.Equal(t, testCase.input.Message, created.Message)
			assert.False(t, created.CreatedAt.IsZero())

		})
	}
}

func TestDatabase_GetEvent(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (IntegrationTestSuite, uuid.UUID)
		wantErr bool
	}{
		{
			name: "fail - event not found",
			setup: func() (IntegrationTestSuite, uuid.UUID) {
				return setupIntegrationTestSuite(t), uuid.Must(uuid.NewV7())
			},
			wantErr: true,
		},
		{
			name: "success - retrieves event by id",
			setup: func() (IntegrationTestSuite, uuid.UUID) {
				testSuite := setupIntegrationTestSuite(t)
				eventID := uuid.Must(uuid.NewV7())
				_, err := testSuite.BHDatabase.CreateEvent(testSuite.Context, model.Event{
					ID:      eventID,
					Type:    "test_event",
					Message: "retrievable event",
				})
				require.NoError(t, err)
				return testSuite, eventID
			},
			wantErr: false,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite, eventID := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			event, err := testSuite.BHDatabase.GetEvent(testSuite.Context, eventID)
			if testCase.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, eventID, event.ID)
				assert.Equal(t, "test_event", event.Type)
				assert.Equal(t, "retrievable event", event.Message)
			}
		})
	}
}

func TestDatabase_GetEvents(t *testing.T) {
	t.Run("success - returns empty list when none exist", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		events, count, err := testSuite.BHDatabase.GetEvents(testSuite.Context, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
		assert.Empty(t, events)
	})

	t.Run("success - returns all created events", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		for i := 0; i < 3; i++ {
			_, err := testSuite.BHDatabase.CreateEvent(testSuite.Context, model.Event{
				ID:      uuid.Must(uuid.NewV7()),
				Type:    "test_event",
				Message: "event message",
			})
			require.NoError(t, err)
		}

		events, count, err := testSuite.BHDatabase.GetEvents(testSuite.Context, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 3, count)
		assert.Len(t, events, 3)
	})

	t.Run("success - default sort is created_at descending", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		for _, eventType := range []string{"first", "second", "third"} {
			_, err := testSuite.BHDatabase.CreateEvent(testSuite.Context, model.Event{
				ID:      uuid.Must(uuid.NewV7()),
				Type:    eventType,
				Message: eventType,
			})
			require.NoError(t, err)
		}

		events, _, err := testSuite.BHDatabase.GetEvents(testSuite.Context, model.SQLFilter{}, nil, 0, 0)
		require.NoError(t, err)
		require.Len(t, events, 3)
		assert.Equal(t, "third", events[0].Type)
		assert.Equal(t, "second", events[1].Type)
		assert.Equal(t, "first", events[2].Type)
	})

	t.Run("success - pagination with skip and limit", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		for i := 0; i < 5; i++ {
			_, err := testSuite.BHDatabase.CreateEvent(testSuite.Context, model.Event{
				ID:      uuid.Must(uuid.NewV7()),
				Type:    "paginated_event",
				Message: "event message",
			})
			require.NoError(t, err)
		}

		events, count, err := testSuite.BHDatabase.GetEvents(testSuite.Context, model.SQLFilter{}, nil, 2, 2)
		require.NoError(t, err)
		assert.Equal(t, 5, count)
		assert.Len(t, events, 2)
	})

	t.Run("success - filter by type", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		for _, eventType := range []string{"analysis_complete", "ingest_started", "analysis_complete"} {
			_, err := testSuite.BHDatabase.CreateEvent(testSuite.Context, model.Event{
				ID:   uuid.Must(uuid.NewV7()),
				Type: eventType,
			})
			require.NoError(t, err)
		}

		filter := model.SQLFilter{
			SQLString: "type = ?",
			Params:    []any{"analysis_complete"},
		}

		events, count, err := testSuite.BHDatabase.GetEvents(testSuite.Context, filter, nil, 0, 0)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Len(t, events, 2)
		for _, event := range events {
			assert.Equal(t, "analysis_complete", event.Type)
		}
	})

	t.Run("success - custom sort ascending", func(t *testing.T) {
		testSuite := setupIntegrationTestSuite(t)
		defer teardownIntegrationTestSuite(t, &testSuite)

		for _, eventType := range []string{"charlie", "alpha", "bravo"} {
			_, err := testSuite.BHDatabase.CreateEvent(testSuite.Context, model.Event{
				ID:   uuid.Must(uuid.NewV4()),
				Type: eventType,
			})
			require.NoError(t, err)
		}

		sortItems := model.Sort{
			{Column: "type", Direction: model.AscendingSortDirection},
		}

		events, _, err := testSuite.BHDatabase.GetEvents(testSuite.Context, model.SQLFilter{}, sortItems, 0, 0)
		require.NoError(t, err)
		require.Len(t, events, 3)
		assert.Equal(t, "alpha", events[0].Type)
		assert.Equal(t, "bravo", events[1].Type)
		assert.Equal(t, "charlie", events[2].Type)
	})
}
