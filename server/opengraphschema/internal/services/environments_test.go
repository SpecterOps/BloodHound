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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services/mocks"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetEnvironmentKindsAndSchemaEnvironmentData(t *testing.T) {
	var (
		ctx           = context.Background()
		unexpectedErr = errors.New("connection refused")
		azureEnv      = model.SchemaEnvironment{EnvironmentKindName: "AZBase", EnvironmentKindId: 1}
		adEnv         = model.SchemaEnvironment{EnvironmentKindName: "Base", EnvironmentKindId: 2}
	)

	tests := []struct {
		name        string
		onlyBuiltin bool
		setupMock   func(databaseMock *mocks.MockDatabase)
		wantKinds   graph.Kinds
		wantEnvMap  model.EnvironmentKindsToEnvironment
		wantErr     error
	}{
		{
			name:        "success_-_maps_environments_to_kinds_and_map",
			onlyBuiltin: false,
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetEnvironmentsFiltered(ctx, false).
					Return([]model.SchemaEnvironment{azureEnv, adEnv}, nil)
			},
			wantKinds: graph.Kinds{graph.StringKind("AZBase"), graph.StringKind("Base")},
			wantEnvMap: model.EnvironmentKindsToEnvironment{
				"AZBase": azureEnv,
				"Base":   adEnv,
			},
		},
		{
			name:        "success_-_forwards_only_builtin_flag_and_handles_empty",
			onlyBuiltin: true,
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetEnvironmentsFiltered(ctx, true).
					Return([]model.SchemaEnvironment{}, nil)
			},
			wantKinds:  graph.Kinds{},
			wantEnvMap: model.EnvironmentKindsToEnvironment{},
		},
		{
			name:        "error_-_propagates_database_error",
			onlyBuiltin: false,
			setupMock: func(databaseMock *mocks.MockDatabase) {
				databaseMock.EXPECT().GetEnvironmentsFiltered(ctx, false).
					Return(nil, unexpectedErr)
			},
			wantErr: unexpectedErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				databaseMock = mocks.NewMockDatabase(t)
				svc          = services.NewService(databaseMock)
			)

			tt.setupMock(databaseMock)

			kinds, envMap, err := svc.GetEnvironmentKindsAndSchemaEnvironmentData(ctx, tt.onlyBuiltin)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, kinds)
				assert.Nil(t, envMap)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantKinds, kinds)
			assert.Equal(t, tt.wantEnvMap, envMap)
		})
	}
}
