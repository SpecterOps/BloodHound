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

package audit

import (
	"context"
	"errors"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuditService struct {
	begin    func(context.Context, Entry) (uuid.UUID, error)
	complete func(context.Context, uuid.UUID, Entry, error) error
}

func (s mockAuditService) Begin(ctx context.Context, entry Entry) (uuid.UUID, error) {
	return s.begin(ctx, entry)
}

func (s mockAuditService) Complete(ctx context.Context, commitID uuid.UUID, entry Entry, operationErr error) error {
	return s.complete(ctx, commitID, entry, operationErr)
}

func TestNewAuditLogAdapter_Dependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		pool             *pgxpool.Pool
		identityResolver auth.IdentityResolver
		expectedPanic    string
	}{
		{
			name:             "Success: constructs audit service",
			pool:             new(pgxpool.Pool),
			identityResolver: auth.NewIdentityResolver(),
		},
		{
			name:             "Error: nil pool",
			identityResolver: auth.NewIdentityResolver(),
			expectedPanic:    "audit: NewAuditLogAdapter requires a non-nil Pool",
		},
		{
			name:          "Error: nil identity resolver",
			pool:          new(pgxpool.Pool),
			expectedPanic: "audit: NewAuditLogAdapter requires a non-nil IdentityResolver",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			construct := func() {
				adapter := NewAuditLogAdapter(testCase.pool, testCase.identityResolver)
				assert.NotNil(t, adapter)
			}

			if testCase.expectedPanic != "" {
				assert.PanicsWithValue(t, testCase.expectedPanic, construct)
			} else {
				assert.NotPanics(t, construct)
			}
		})
	}
}

func TestAuditLogAdapter_Begin(t *testing.T) {
	t.Parallel()

	var (
		ctx              = context.Background()
		entry            = Entry{Action: ActionUpdateParameter}
		expectedCommitID = uuid.Must(uuid.NewV4())
		expectedErr      = errors.New("begin failed")
	)

	tests := []struct {
		name             string
		expectedCommitID uuid.UUID
		expectedErr      error
	}{
		{name: "Success: delegates audit intent", expectedCommitID: expectedCommitID},
		{name: "Error: returns service error", expectedErr: expectedErr},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			adapter := &AuditLogAdapter{service: mockAuditService{
				begin: func(actualCtx context.Context, actualEntry Entry) (uuid.UUID, error) {
					assert.Equal(t, ctx, actualCtx)
					assert.Equal(t, entry, actualEntry)
					return testCase.expectedCommitID, testCase.expectedErr
				},
			}}

			commitID, err := adapter.Begin(ctx, entry)

			assert.Equal(t, testCase.expectedCommitID, commitID)
			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuditLogAdapter_Complete(t *testing.T) {
	t.Parallel()

	var (
		ctx          = context.Background()
		commitID     = uuid.Must(uuid.NewV4())
		entry        = Entry{Action: ActionUpdateParameter}
		operationErr = errors.New("operation failed")
		expectedErr  = errors.New("complete failed")
	)

	tests := []struct {
		name        string
		serviceErr  error
		expectedErr error
	}{
		{name: "Success: delegates audit outcome"},
		{name: "Error: returns service error", serviceErr: expectedErr, expectedErr: expectedErr},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			adapter := &AuditLogAdapter{service: mockAuditService{
				complete: func(actualCtx context.Context, actualCommitID uuid.UUID, actualEntry Entry, actualOperationErr error) error {
					assert.Equal(t, ctx, actualCtx)
					assert.Equal(t, commitID, actualCommitID)
					assert.Equal(t, entry, actualEntry)
					assert.ErrorIs(t, actualOperationErr, operationErr)
					return testCase.serviceErr
				},
			}}

			err := adapter.Complete(ctx, commitID, entry, operationErr)

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
