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

// sampleEntry returns a fully populated Entry for mapping assertions.
func sampleEntry() services.Entry {
	return services.Entry{
		Action:          "POST /api/v2/roles/{role_id}",
		ActorID:         "actor-id",
		ActorName:       "actor-name",
		ActorEmail:      "actor@example.com",
		RequestID:       "req-1",
		SourceIPAddress: "10.0.0.1",
		Fields:          map[string]any{"key": "value"},
	}
}

// captureInsert wires the mock to record the single AuditRecord passed to
// InsertAuditLog and return the provided error.
func captureInsert(db *mocks.MockDatabase, captured *services.AuditRecord, returnErr error) {
	db.EXPECT().
		InsertAuditLog(mock.Anything, mock.Anything).
		Run(func(_ context.Context, record services.AuditRecord) {
			*captured = record
		}).
		Return(returnErr).
		Once()
}

func TestService_Intent_WritesIntentRowAndReturnsCommitID(t *testing.T) {
	var (
		db       = mocks.NewMockDatabase(t)
		service  = services.NewService(db)
		captured services.AuditRecord
		entry    = sampleEntry()
	)
	captureInsert(db, &captured, nil)

	commitID, err := service.Intent(context.Background(), entry)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.UUID{}, commitID, "a commit id should be generated")

	assert.Equal(t, services.StatusIntent, captured.Status)
	assert.Equal(t, commitID, captured.CommitID, "written row must carry the returned commit id")

	// Field mapping (toRecord) is exercised through the captured record.
	assert.Equal(t, entry.Action, captured.Action)
	assert.Equal(t, entry.ActorID, captured.ActorID)
	assert.Equal(t, entry.ActorName, captured.ActorName)
	assert.Equal(t, entry.ActorEmail, captured.ActorEmail)
	assert.Equal(t, entry.RequestID, captured.RequestID)
	assert.Equal(t, entry.SourceIPAddress, captured.SourceIPAddress)
	assert.Equal(t, map[string]any{"key": "value"}, captured.Fields)
	assert.Equal(t, services.SourceMiddleware, captured.Source, "source is always middleware")
}

func TestService_Success_WritesSuccessRowWithProvidedCommitID(t *testing.T) {
	var (
		db       = mocks.NewMockDatabase(t)
		service  = services.NewService(db)
		captured services.AuditRecord
		commitID = uuid.Must(uuid.NewV4())
	)
	captureInsert(db, &captured, nil)

	require.NoError(t, service.Success(context.Background(), commitID, sampleEntry()))
	assert.Equal(t, services.StatusSuccess, captured.Status)
	assert.Equal(t, commitID, captured.CommitID)
	assert.Equal(t, services.SourceMiddleware, captured.Source)
}

func TestService_Failure_WritesFailureRowWithProvidedCommitID(t *testing.T) {
	var (
		db       = mocks.NewMockDatabase(t)
		service  = services.NewService(db)
		captured services.AuditRecord
		commitID = uuid.Must(uuid.NewV4())
	)
	captureInsert(db, &captured, nil)

	require.NoError(t, service.Failure(context.Background(), commitID, sampleEntry()))
	assert.Equal(t, services.StatusFailure, captured.Status)
	assert.Equal(t, commitID, captured.CommitID)
}

func TestService_Intent_PropagatesInsertErrorAndReturnsCommitID(t *testing.T) {
	var (
		sentinel = errors.New("insert failed")
		db       = mocks.NewMockDatabase(t)
		service  = services.NewService(db)
		captured services.AuditRecord
	)
	captureInsert(db, &captured, sentinel)

	commitID, err := service.Intent(context.Background(), sampleEntry())
	require.ErrorIs(t, err, sentinel)
	// The current contract returns the generated id even on insert failure so
	// callers can correlate a failed intent if needed.
	assert.NotEqual(t, uuid.UUID{}, commitID)
}

func TestService_SuccessAndFailure_PropagateInsertError(t *testing.T) {
	var (
		sentinel = errors.New("insert failed")
		commitID = uuid.Must(uuid.NewV4())
	)

	t.Run("success", func(t *testing.T) {
		var (
			db       = mocks.NewMockDatabase(t)
			service  = services.NewService(db)
			captured services.AuditRecord
		)
		captureInsert(db, &captured, sentinel)
		require.ErrorIs(t, service.Success(context.Background(), commitID, sampleEntry()), sentinel)
	})

	t.Run("failure", func(t *testing.T) {
		var (
			db       = mocks.NewMockDatabase(t)
			service  = services.NewService(db)
			captured services.AuditRecord
		)
		captureInsert(db, &captured, sentinel)
		require.ErrorIs(t, service.Failure(context.Background(), commitID, sampleEntry()), sentinel)
	})
}

// TestService_RedactsSensitiveFields exercises redactSensitiveFields indirectly
// through Intent by asserting on the AuditRecord.Fields captured by the mock.
func TestService_RedactsSensitiveFields(t *testing.T) {
	const redacted = "[REDACTED]"

	var cases = []struct {
		name     string
		fields   map[string]any
		expected map[string]any
	}{
		{
			name:     "nil fields pass through",
			fields:   nil,
			expected: nil,
		},
		{
			name:     "empty fields pass through",
			fields:   map[string]any{},
			expected: map[string]any{},
		},
		{
			name:     "non-sensitive keys are unchanged",
			fields:   map[string]any{"role": "admin", "count": 3},
			expected: map[string]any{"role": "admin", "count": 3},
		},
		{
			name:     "exact sensitive keys are redacted",
			fields:   map[string]any{"password": "p", "secret": "s", "token": "t"},
			expected: map[string]any{"password": redacted, "secret": redacted, "token": redacted},
		},
		{
			name:     "underscore variants are redacted",
			fields:   map[string]any{"api_key": "a", "private_key": "k"},
			expected: map[string]any{"api_key": redacted, "private_key": redacted},
		},
		{
			name:     "no-underscore variants are redacted",
			fields:   map[string]any{"apikey": "a", "privatekey": "k"},
			expected: map[string]any{"apikey": redacted, "privatekey": redacted},
		},
		{
			name:     "matching is case-insensitive",
			fields:   map[string]any{"Password": "p", "X-API-KEY": "a"},
			expected: map[string]any{"Password": redacted, "X-API-KEY": redacted},
		},
		{
			name:     "matching is substring based",
			fields:   map[string]any{"user_password_hash": "h", "auth_token_id": "1"},
			expected: map[string]any{"user_password_hash": redacted, "auth_token_id": redacted},
		},
		{
			name:     "mixed sensitive and non-sensitive",
			fields:   map[string]any{"role": "admin", "secret": "s"},
			expected: map[string]any{"role": "admin", "secret": redacted},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				db       = mocks.NewMockDatabase(t)
				service  = services.NewService(db)
				captured services.AuditRecord
				entry    = sampleEntry()
			)
			entry.Fields = tc.fields
			captureInsert(db, &captured, nil)

			_, err := service.Intent(context.Background(), entry)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, captured.Fields)
		})
	}
}
