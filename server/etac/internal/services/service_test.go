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
	t.Parallel()
	type args struct {
		appdb          services.AppDatabase
		dogtagsService dogtags.Service
	}
	type want struct {
		panic string
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Error: panics with nil appdb",
			args: args{
				appdb:          nil,
				dogtagsService: dogtagsWith(false),
			},
			want: want{panic: "etac: service requires a non-nil appdb"},
		},
		{
			name: "Error: panics with nil dogtags service",
			args: args{
				appdb:          fakeETACDatabase{},
				dogtagsService: nil,
			},
			want: want{panic: "etac: service requires a non-nil dogtagsService"},
		},
		{
			name: "Success: returns service with valid dependencies",
			args: args{
				appdb:          fakeETACDatabase{},
				dogtagsService: dogtagsWith(false),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if test.want.panic != "" {
				assert.PanicsWithValue(t, test.want.panic, func() {
					services.NewService(test.args.appdb, test.args.dogtagsService)
				})
				return
			}

			assert.NotNil(t, services.NewService(test.args.appdb, test.args.dogtagsService))
		})
	}
}

func TestService_CheckUserAccess(t *testing.T) {
	t.Parallel()
	var (
		ctx    = context.Background()
		userID = uuid.Must(uuid.NewV4())
		dbErr  = errors.New("connection refused")
	)

	type args struct {
		etacEnabled  bool
		user         users.User
		db           fakeETACDatabase
		environments []string
	}

	type want struct {
		result bool
		err    error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "allows when ETAC feature is disabled",
			args: args{
				etacEnabled:  false,
				user:         &fakeUser{ID: userID},
				environments: []string{"env-1"},
			},
			want: want{result: true},
		},
		{
			name: "allows when user has access to all environments",
			args: args{
				etacEnabled:  true,
				user:         &fakeUser{ID: userID, AllEnvironments: true},
				environments: []string{"env-1"},
			},
			want: want{result: true},
		},
		{
			name: "allows when user is permitted every requested environment",
			args: args{
				etacEnabled: true,
				user:        &fakeUser{ID: userID},
				db: fakeETACDatabase{rows: []services.EnvironmentTargetedAccessControl{
					{UserID: userID.String(), EnvironmentID: "env-1"},
					{UserID: userID.String(), EnvironmentID: "env-2"},
				}},
				environments: []string{"env-1", "env-2"},
			},
			want: want{result: true},
		},
		{
			name: "denies when a requested environment is not permitted",
			args: args{
				etacEnabled: true,
				user:        &fakeUser{ID: userID},
				db: fakeETACDatabase{rows: []services.EnvironmentTargetedAccessControl{
					{UserID: userID.String(), EnvironmentID: "env-1"},
				}},
				environments: []string{"env-1", "env-2"},
			},
			want: want{result: false},
		},
		{
			name: "propagates database errors",
			args: args{
				etacEnabled:  true,
				user:         &fakeUser{ID: userID},
				db:           fakeETACDatabase{err: dbErr},
				environments: []string{"env-1"},
			},
			want: want{err: dbErr},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := services.NewService(tt.args.db, dogtagsWith(tt.args.etacEnabled))
			got, err := svc.CheckUserAccess(ctx, tt.args.user, tt.args.environments...)

			if tt.want.err != nil {
				assert.ErrorIs(t, err, tt.want.err)
				assert.False(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.result, got)
			}
		})
	}
}

func TestService_FilterEnvironmentsByAccess(t *testing.T) {
	t.Parallel()
	var (
		ctx    = context.Background()
		userID = uuid.Must(uuid.NewV4())
		dbErr  = errors.New("connection refused")
		envs   = []services.EnvironmentTargetedAccessControl{
			{UserID: userID.String(), EnvironmentID: "env-1"},
			{UserID: userID.String(), EnvironmentID: "env-2"},
		}
	)

	type args struct {
		etacEnabled     bool
		user            users.User
		db              fakeETACDatabase
		requestedEnvIDs []string
	}
	type want struct {
		result []string
		err    error
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ETAC disabled, no requested IDs: no filter",
			args: args{
				etacEnabled: false,
				user:        &fakeUser{ID: userID},
			},
			want: want{result: nil},
		},
		{
			name: "ETAC disabled, requested IDs passed through",
			args: args{
				etacEnabled:     false,
				user:            &fakeUser{ID: userID},
				requestedEnvIDs: []string{"env-1"},
			},
			want: want{result: []string{"env-1"}},
		},
		{
			name: "all-environments user: no filter",
			args: args{
				etacEnabled: true,
				user:        &fakeUser{ID: userID, AllEnvironments: true},
			},
			want: want{result: nil},
		},
		{
			name: "ETAC user, no requested IDs: returns full allowlist",
			args: args{
				etacEnabled: true,
				user:        &fakeUser{ID: userID},
				db:          fakeETACDatabase{rows: envs},
			},
			want: want{result: []string{"env-1", "env-2"}},
		},
		{
			name: "ETAC user, requested ID in allowlist: returns intersection",
			args: args{
				etacEnabled:     true,
				user:            &fakeUser{ID: userID},
				db:              fakeETACDatabase{rows: envs},
				requestedEnvIDs: []string{"env-1"},
			},
			want: want{result: []string{"env-1"}},
		},
		{
			name: "ETAC user, requested ID outside allowlist: returns sentinel",
			args: args{
				etacEnabled:     true,
				user:            &fakeUser{ID: userID},
				db:              fakeETACDatabase{rows: envs},
				requestedEnvIDs: []string{"env-99"},
			},
			want: want{result: []string{""}},
		},
		{
			name: "ETAC user, empty allowlist: returns sentinel",
			args: args{
				etacEnabled: true,
				user:        &fakeUser{ID: userID},
				db:          fakeETACDatabase{rows: nil},
			},
			want: want{result: []string{""}},
		},
		{
			name: "database error propagates",
			args: args{
				etacEnabled: true,
				user:        &fakeUser{ID: userID},
				db:          fakeETACDatabase{err: dbErr},
			},
			want: want{err: dbErr},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := services.NewService(tt.args.db, dogtagsWith(tt.args.etacEnabled))
			got, err := svc.FilterEnvironmentsByAccess(ctx, tt.args.user, tt.args.requestedEnvIDs)

			if tt.want.err != nil {
				assert.ErrorIs(t, err, tt.want.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want.result, got)
			}
		})
	}
}

// TestService_checkUserAccessToEnvironments_AllEnvironments covers the
// short-circuit branch that is unreachable through CheckUserAccess (because
// shouldFilterForETAC already excludes all-environment users).
func TestService_checkUserAccessToEnvironments_AllEnvironments(t *testing.T) {
	t.Parallel()
	svc := services.NewService(fakeETACDatabase{err: errors.New("should not be called")}, dogtagsWith(true))

	got, err := svc.CheckUserAccessToEnvironments(context.Background(), &fakeUser{AllEnvironments: true}, "env-1")

	require.NoError(t, err)
	assert.True(t, got)
}
