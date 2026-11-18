// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/server/audit/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mock struct {
	database         *mockDatabase
	identityResolver mockIdentityResolver
}

type mockDatabase struct {
	logs             []services.AuditLog
	contextSnapshots []contextSnapshot
	err              error
}

type contextSnapshot struct {
	err         error
	deadline    time.Time
	hasDeadline bool
}

func (s *mockDatabase) CreateAuditLog(ctx context.Context, auditLog services.AuditLog) error {
	deadline, hasDeadline := ctx.Deadline()
	s.contextSnapshots = append(s.contextSnapshots, contextSnapshot{
		err:         ctx.Err(),
		deadline:    deadline,
		hasDeadline: hasDeadline,
	})
	s.logs = append(s.logs, auditLog)
	return s.err
}

type mockIdentityResolver struct {
	identity          auth.SimpleIdentity
	expectedAuthOwner any
	err               error
}

func TestNewService_NilDependenciesPanic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		construct     func(*mockDatabase, mockIdentityResolver)
		expectedPanic string
	}{
		{
			name: "Error: nil Database",
			construct: func(_ *mockDatabase, identityResolver mockIdentityResolver) {
				services.NewService(nil, identityResolver)
			},
			expectedPanic: "audit: service requires a non-nil Database",
		},
		{
			name: "Error: nil IdentityResolver",
			construct: func(database *mockDatabase, _ mockIdentityResolver) {
				services.NewService(database, nil)
			},
			expectedPanic: "audit: service requires a non-nil IdentityResolver",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assert.PanicsWithValue(t, testCase.expectedPanic, func() {
				testCase.construct(&mockDatabase{}, mockIdentityResolver{})
			})
		})
	}
}

func (s mockIdentityResolver) GetIdentity(authContext auth.Context) (auth.SimpleIdentity, error) {
	if s.expectedAuthOwner != nil && authContext.Owner != s.expectedAuthOwner {
		return auth.SimpleIdentity{}, errors.New("unexpected auth context owner")
	}

	return s.identity, s.err
}

func TestService_Begin(t *testing.T) {
	t.Parallel()

	var (
		actorID        = uuid.Must(uuid.NewV4())
		writeErr       = errors.New("write failed")
		requestContext = bhctx.Set(context.Background(), &bhctx.Context{
			RequestID: "request-id",
			RequestIP: "127.0.0.1",
		})
		entry = services.AuditLogEntry{
			Action: services.ActionUpdateParameter,
			Fields: map[string]any{"key": "value"},
		}
	)

	tests := []struct {
		name       string
		setupMocks func(*mock)
		expected   struct {
			error      error
			logCount   int
			actorID    string
			actorName  string
			actorEmail string
		}
	}{
		{
			name: "Success: writes an authenticated intent",
			setupMocks: func(m *mock) {
				m.identityResolver.identity = auth.SimpleIdentity{ID: actorID, Name: "actor", Email: "actor@example.com"}
				m.identityResolver.err = nil
			},
			expected: struct {
				error      error
				logCount   int
				actorID    string
				actorName  string
				actorEmail string
			}{logCount: 1, actorID: actorID.String(), actorName: "actor", actorEmail: "actor@example.com"},
		},
		{
			name: "Success: writes an anonymous intent when identity resolution fails",
			setupMocks: func(m *mock) {
				m.identityResolver.err = errors.New("identity unavailable")
			},
			expected: struct {
				error      error
				logCount   int
				actorID    string
				actorName  string
				actorEmail string
			}{logCount: 1},
		},
		{
			name: "Error: audit insert failure",
			setupMocks: func(m *mock) {
				m.database.err = writeErr
			},
			expected: struct {
				error      error
				logCount   int
				actorID    string
				actorName  string
				actorEmail string
			}{error: services.ErrAuditLog, logCount: 1},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var mocks = mock{
				database:         &mockDatabase{},
				identityResolver: mockIdentityResolver{err: errors.New("identity unavailable")},
			}
			testCase.setupMocks(&mocks)

			service := services.NewService(mocks.database, mocks.identityResolver)
			commitID, err := service.Begin(requestContext, entry)

			if testCase.expected.error != nil {
				assert.ErrorIs(t, err, testCase.expected.error)
				assert.NotErrorIs(t, err, writeErr)
				assert.ErrorContains(t, err, writeErr.Error())
				assert.True(t, commitID.IsNil())
			} else {
				require.NoError(t, err)
				assert.False(t, commitID.IsNil())
			}
			require.Len(t, mocks.database.logs, testCase.expected.logCount)
			auditLog := mocks.database.logs[0]
			assert.Equal(t, services.StatusIntent, auditLog.Status)
			assert.Equal(t, entry.Action, auditLog.Action)
			assert.Equal(t, entry.Fields, auditLog.Fields)
			assert.Equal(t, "request-id", auditLog.RequestID)
			assert.Equal(t, "127.0.0.1", auditLog.SourceIPAddress)
			assert.Equal(t, testCase.expected.actorID, auditLog.ActorID)
			assert.Equal(t, testCase.expected.actorName, auditLog.ActorName)
			assert.Equal(t, testCase.expected.actorEmail, auditLog.ActorEmail)
		})
	}
}

func TestService_Complete(t *testing.T) {
	t.Parallel()

	type expected struct {
		status          services.Status
		error           error
		fields          map[string]any
		requestID       string
		sourceIPAddress string
		actorID         string
	}

	var (
		actorID      = uuid.Must(uuid.NewV4())
		commitID     = uuid.Must(uuid.NewV4())
		operationErr = errors.New("operation failed")
		writeErr     = errors.New("write failed")
		entry        = services.AuditLogEntry{
			Action: services.ActionUpdateParameter,
			Fields: map[string]any{"key": "value"},
		}
	)

	tests := []struct {
		name             string
		buildContext     func() context.Context
		identityResolver mockIdentityResolver
		operationErr     error
		writeErr         error
		expected         expected
	}{
		{
			name:     "Success: writes success outcome",
			expected: expected{status: services.StatusSuccess, fields: map[string]any{"key": "value"}},
		},
		{
			name:         "Success: writes failure outcome with operation error",
			operationErr: operationErr,
			expected:     expected{status: services.StatusFailure, fields: map[string]any{"key": "value", "error": operationErr.Error()}},
		},
		{
			name:     "Error: audit insert failure",
			writeErr: writeErr,
			expected: expected{status: services.StatusSuccess, error: services.ErrAuditLog, fields: map[string]any{"key": "value"}},
		},
		{
			name: "Success: writes outcome after request cancellation and preserves request values",
			buildContext: func() context.Context {
				requestContext := bhctx.Set(context.Background(), &bhctx.Context{
					RequestID: "request-id",
					RequestIP: "127.0.0.1",
					AuthCtx:   auth.Context{Owner: "actor-owner"},
				})
				canceledContext, cancel := context.WithCancel(requestContext)
				cancel()

				return canceledContext
			},
			identityResolver: mockIdentityResolver{
				identity:          auth.SimpleIdentity{ID: actorID},
				expectedAuthOwner: "actor-owner",
			},
			expected: expected{
				status:          services.StatusSuccess,
				fields:          map[string]any{"key": "value"},
				requestID:       "request-id",
				sourceIPAddress: "127.0.0.1",
				actorID:         actorID.String(),
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var (
				ctx              = context.Background()
				database         = &mockDatabase{err: testCase.writeErr}
				identityResolver = mockIdentityResolver{err: errors.New("identity unavailable")}
			)
			if testCase.buildContext != nil {
				ctx = testCase.buildContext()
				identityResolver = testCase.identityResolver
			}
			service := services.NewService(database, identityResolver)

			err := service.Complete(ctx, commitID, entry, testCase.operationErr)

			if testCase.expected.error != nil {
				assert.ErrorIs(t, err, testCase.expected.error)
				assert.NotErrorIs(t, err, writeErr)
				assert.ErrorContains(t, err, writeErr.Error())
			} else {
				require.NoError(t, err)
			}
			require.Len(t, database.logs, 1)
			require.Len(t, database.contextSnapshots, 1)
			assert.NoError(t, database.contextSnapshots[0].err)
			assert.True(t, database.contextSnapshots[0].hasDeadline)
			assert.Positive(t, time.Until(database.contextSnapshots[0].deadline))
			assert.LessOrEqual(t, time.Until(database.contextSnapshots[0].deadline), 30*time.Second)
			assert.Equal(t, testCase.expected.status, database.logs[0].Status)
			assert.Equal(t, commitID, database.logs[0].CommitID)
			assert.Equal(t, testCase.expected.fields, database.logs[0].Fields)
			assert.Equal(t, testCase.expected.requestID, database.logs[0].RequestID)
			assert.Equal(t, testCase.expected.sourceIPAddress, database.logs[0].SourceIPAddress)
			assert.Equal(t, testCase.expected.actorID, database.logs[0].ActorID)
			assert.Equal(t, map[string]any{"key": "value"}, entry.Fields)
		})
	}
}
