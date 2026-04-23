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
