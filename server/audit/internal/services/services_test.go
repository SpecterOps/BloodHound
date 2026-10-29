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

package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/server/audit/internal/services"
	"github.com/specterops/bloodhound/server/audit/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testEntry() services.Entry {
	return services.Entry{
		Action:          "POST/api/v2/roles/{role_id}",
		ActorID:         "actor-id",
		ActorName:       "actor-name",
		ActorEmail:      "actor@example.com",
		RequestID:       "request-id",
		SourceIPAddress: "10.0.0.1",
		Fields:          map[string]any{"role": "admin"},
	}
}

func captureRecord(database *mocks.MockDatabase, capturedRecord *services.AuditRecord, returnError error) {
	database.EXPECT().InsertAuditLog(mock.Anything, mock.Anything).
		Run(func(_ context.Context, record services.AuditRecord) { *capturedRecord = record }).
		Return(returnError).
		Once()
}

func TestService_Intent(t *testing.T) {
	var (
		insertError = errors.New("insert failed")
		testCases   = []struct {
			name          string
			databaseError error
		}{
			{name: "writes intent"},
			{name: "returns database error", databaseError: insertError},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				database       = mocks.NewMockDatabase(t)
				service        = services.NewService(database)
				capturedRecord services.AuditRecord
			)
			captureRecord(database, &capturedRecord, testCase.databaseError)

			commitID, err := service.Intent(context.Background(), testEntry())
			assert.NotEqual(t, uuid.Nil, commitID)
			if testCase.databaseError == nil {
				require.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testCase.databaseError)
			}
			assert.Equal(t, services.StatusIntent, capturedRecord.Status)
			assert.Equal(t, commitID, capturedRecord.CommitID)
			assert.Equal(t, services.SourceMiddleware, capturedRecord.Source)
			assert.Equal(t, testEntry().Action, capturedRecord.Action)
			assert.Equal(t, testEntry().ActorID, capturedRecord.ActorID)
			assert.Equal(t, testEntry().ActorName, capturedRecord.ActorName)
			assert.Equal(t, testEntry().ActorEmail, capturedRecord.ActorEmail)
			assert.Equal(t, testEntry().RequestID, capturedRecord.RequestID)
			assert.Equal(t, testEntry().SourceIPAddress, capturedRecord.SourceIPAddress)
		})
	}
}

func TestService_Outcome(t *testing.T) {
	testCases := []struct {
		name           string
		writeOutcome   func(*services.Service, context.Context, uuid.UUID, services.Entry) error
		expectedStatus services.Status
	}{
		{name: "writes success", writeOutcome: (*services.Service).Success, expectedStatus: services.StatusSuccess},
		{name: "writes failure", writeOutcome: (*services.Service).Failure, expectedStatus: services.StatusFailure},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				database       = mocks.NewMockDatabase(t)
				service        = services.NewService(database)
				capturedRecord services.AuditRecord
				commitID       = uuid.Must(uuid.NewV4())
			)
			captureRecord(database, &capturedRecord, nil)

			require.NoError(t, testCase.writeOutcome(service, context.Background(), commitID, testEntry()))
			assert.Equal(t, testCase.expectedStatus, capturedRecord.Status)
			assert.Equal(t, commitID, capturedRecord.CommitID)
		})
	}
}

func TestService_DefaultsUnknownActor(t *testing.T) {
	var (
		database       = mocks.NewMockDatabase(t)
		service        = services.NewService(database)
		capturedRecord services.AuditRecord
		entry          = testEntry()
	)
	entry.ActorID = ""
	entry.ActorName = ""
	entry.ActorEmail = ""
	captureRecord(database, &capturedRecord, nil)

	_, err := service.Intent(context.Background(), entry)
	require.NoError(t, err)
	assert.Equal(t, services.UnknownActorName, capturedRecord.ActorName)
}

func TestService_RedactsSensitiveFields(t *testing.T) {
	testCases := []struct {
		name           string
		fields         map[string]any
		expectedFields map[string]any
	}{
		{name: "keeps nil fields", fields: nil, expectedFields: nil},
		{name: "keeps empty fields", fields: map[string]any{}, expectedFields: map[string]any{}},
		{name: "keeps safe fields", fields: map[string]any{"role": "admin"}, expectedFields: map[string]any{"role": "admin"}},
		{name: "redacts case and separator variants", fields: map[string]any{"Password": "one", "api_key": "two", "private-key": "three"}, expectedFields: map[string]any{"Password": "[REDACTED]", "api_key": "[REDACTED]", "private-key": "[REDACTED]"}},
		{name: "redacts substring matches", fields: map[string]any{"auth_token_id": "one"}, expectedFields: map[string]any{"auth_token_id": "[REDACTED]"}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				database       = mocks.NewMockDatabase(t)
				service        = services.NewService(database)
				capturedRecord services.AuditRecord
				entry          = testEntry()
			)
			entry.Fields = testCase.fields
			captureRecord(database, &capturedRecord, nil)

			_, err := service.Intent(context.Background(), entry)
			require.NoError(t, err)
			assert.Equal(t, testCase.expectedFields, capturedRecord.Fields)
		})
	}
}
