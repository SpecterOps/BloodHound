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

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/server/etac/internal/services"
	"github.com/specterops/bloodhound/server/etac/internal/services/mocks"
	"github.com/specterops/bloodhound/server/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeUser struct {
	ID              uuid.UUID
	AllEnvironments bool
}

func (f *fakeUser) HasAllEnvironments() bool {
	return f.AllEnvironments
}

func (f *fakeUser) GetID() uuid.UUID {
	return f.ID
}

// fakeETACDatabase is a minimal in-memory implementation of the Database port
// used to drive the Service access checks without a real connection pool.
type fakeETACDatabase struct {
	rows []services.EnvironmentTargetedAccessControl
	err  error
}

func (f fakeETACDatabase) GetEnvironmentTargetedAccessControlForUser(_ context.Context, _ uuid.UUID) ([]services.EnvironmentTargetedAccessControl, error) {
	return f.rows, f.err
}

// dogtagsWith returns a real dogtags Service whose ETAC_ENABLED flag is set to
// the supplied value, exercising the production GetFlagAsBool path.
func dogtagsWith(etacEnabled bool) dogtags.Service {
	return dogtags.NewTestService(dogtags.TestOverrides{
		Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: etacEnabled},
	})
}

func TestNewService(t *testing.T) {
	mockAppDb := mocks.NewMockAppDatabase(t)
	assert.NotNil(t, services.NewService(mockAppDb, dogtagsWith(false)))
}

func TestService_CheckUserAccess(t *testing.T) {
	var (
		ctx    = context.Background()
		userID = uuid.Must(uuid.NewV4())
		dbErr  = errors.New("connection refused")
	)

	tests := []struct {
		name         string
		etacEnabled  bool
		user         users.User
		db           fakeETACDatabase
		environments []string
		want         bool
		wantErr      error
	}{
		{
			name:         "allows when ETAC feature is disabled",
			etacEnabled:  false,
			user:         &fakeUser{ID: userID},
			environments: []string{"env-1"},
			want:         true,
		},
		{
			name:         "allows when user has access to all environments",
			etacEnabled:  true,
			user:         &fakeUser{ID: userID, AllEnvironments: true},
			environments: []string{"env-1"},
			want:         true,
		},
		{
			name:        "allows when user is permitted every requested environment",
			etacEnabled: true,
			user:        &fakeUser{ID: userID},
			db: fakeETACDatabase{rows: []services.EnvironmentTargetedAccessControl{
				{UserID: userID.String(), EnvironmentID: "env-1"},
				{UserID: userID.String(), EnvironmentID: "env-2"},
			}},
			environments: []string{"env-1", "env-2"},
			want:         true,
		},
		{
			name:        "denies when a requested environment is not permitted",
			etacEnabled: true,
			user:        &fakeUser{ID: userID},
			db: fakeETACDatabase{rows: []services.EnvironmentTargetedAccessControl{
				{UserID: userID.String(), EnvironmentID: "env-1"},
			}},
			environments: []string{"env-1", "env-2"},
			want:         false,
		},
		{
			name:         "propagates database errors",
			etacEnabled:  true,
			user:         &fakeUser{ID: userID},
			db:           fakeETACDatabase{err: dbErr},
			environments: []string{"env-1"},
			wantErr:      dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := services.NewService(tt.db, dogtagsWith(tt.etacEnabled))
			got, err := svc.CheckUserAccess(ctx, tt.user, tt.environments...)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.False(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// TestService_checkUserAccessToEnvironments_AllEnvironments covers the
// short-circuit branch that is unreachable through CheckUserAccess (because
// shouldFilterForETAC already excludes all-environment users).
func TestService_checkUserAccessToEnvironments_AllEnvironments(t *testing.T) {
	svc := services.NewService(fakeETACDatabase{err: errors.New("should not be called")}, dogtagsWith(true))

	got, err := svc.CheckUserAccessToEnvironments(context.Background(), &fakeUser{AllEnvironments: true}, "env-1")

	require.NoError(t, err)
	assert.True(t, got)
}
