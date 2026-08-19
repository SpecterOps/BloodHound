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

package bootstrap_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bootstrap"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFillAndPopulateDefaultAdminInfo(t *testing.T) {
	type Input struct {
		Config config.DefaultAdminConfiguration
	}

	cases := []struct {
		Input    Input
		Error    error
		NeedsLog bool
	}{
		{Input{config.DefaultAdminConfiguration{
			PrincipalName: "",
			Password:      "",
			EmailAddress:  "",
			FirstName:     "",
			LastName:      "",
			ExpireNow:     false,
		}}, nil, true},
		{Input{config.DefaultAdminConfiguration{
			PrincipalName: "",
			Password:      "abc123",
			EmailAddress:  "",
			FirstName:     "",
			LastName:      "",
			ExpireNow:     false,
		}}, nil, false},
		{Input{config.DefaultAdminConfiguration{
			PrincipalName: "abc123",
			Password:      "",
			EmailAddress:  "test@test.com",
			FirstName:     "",
			LastName:      "",
			ExpireNow:     false,
		}}, nil, true},
	}

	for _, c := range cases {
		cfg, needsLog, err := bootstrap.FillAndPopulateDefaultAdminInfo(c.Input.Config, config.NewDefaultAdminConfiguration)
		require.Equal(t, c.Error, err)
		require.Equal(t, c.NeedsLog, needsLog)
		require.NotEqual(t, "", cfg.EmailAddress)
		require.NotEqual(t, "", cfg.Password)
		require.NotEqual(t, "", cfg.FirstName)
		require.NotEqual(t, "", cfg.LastName)
		require.NotEqual(t, "", cfg.PrincipalName)
	}
}

func TestEnsureServerDirectoriesCreatesRequiredDirectories(t *testing.T) {
	t.Parallel()

	var (
		rootDirectory = t.TempDir()
		cfg           = config.Configuration{
			WorkDir:            filepath.Join(rootDirectory, "work"),
			CollectorsBasePath: filepath.Join(rootDirectory, "collectors"),
		}
	)

	require.NoError(t, bootstrap.EnsureServerDirectories(cfg))

	for _, directory := range []string{
		cfg.WorkDir,
		cfg.TempDirectory(),
		cfg.ScratchDirectory(),
		cfg.RetainedFilesDirectory(),
		cfg.ClientLogDirectory(),
		cfg.CollectorsDirectory(),
	} {
		requireDirectoryExists(t, directory)
	}
}

func requireDirectoryExists(t *testing.T, path string) {
	t.Helper()

	info, err := os.Stat(path)
	require.NoError(t, err)
	require.True(t, info.IsDir())
}

func TestMigrateDB(t *testing.T) {
	tests := []struct {
		name  string
		cfg   config.DefaultAdminConfiguration
		mocks func(mockDb *mocks.MockDatabase)
	}{
		{
			name: "Success - default admin created when enabled in config",
			cfg: config.DefaultAdminConfiguration{
				Enabled:  true,
				Password: "SFdzJoW2GT7Fn68aEieKn7S1S2DLdXnw",
			},
			mocks: func(mockDb *mocks.MockDatabase) {
				mockDb.EXPECT().GetAllRoles(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Roles{
					{
						Name:        auth.RoleAdministrator,
						Description: "Admin For Testing",
						Permissions: model.Permissions{},
						Serial:      model.Serial{},
					},
				}, nil)
				mockDb.EXPECT().LookupUser(gomock.Any(), gomock.Any()).Return(model.User{}, database.ErrNotFound)
				mockDb.EXPECT().GetConfigurationParameter(gomock.Any(), appcfg.PasswordExpirationWindow).Return(appcfg.Parameter{}, nil)
				mockDb.EXPECT().InitializeSecretAuth(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Installation{}, nil).Times(1)
			},
		},
		{
			name: "Success - default admin not created when disabled in config",
			cfg: config.DefaultAdminConfiguration{
				Enabled: false,
			},
			mocks: func(mockDb *mocks.MockDatabase) {
				mockDb.EXPECT().InitializeSecretAuth(gomock.Any(), gomock.Any(), gomock.Any()).Return(model.Installation{}, nil).Times(0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			mockDb := mocks.NewMockDatabase(mockCtrl)
			defer mockCtrl.Finish()

			mockDb.EXPECT().Migrate(gomock.Any()).Return(nil).Times(1)
			mockDb.EXPECT().HasInstallation(gomock.Any()).Return(false, nil).Times(1)
			mockDb.EXPECT().CreateInstallation(gomock.Any()).Return(model.Installation{}, nil).Times(1)

			tt.mocks(mockDb)

			err := bootstrap.MigrateDB(context.Background(), config.Configuration{
				Crypto: config.CryptoConfiguration{
					Argon2: config.Argon2Configuration{
						MemoryKibibytes: 16,
						NumIterations:   2,
						NumThreads:      1,
					},
				},
				DefaultAdmin: tt.cfg,
			}, mockDb, func() (config.DefaultAdminConfiguration, error) {
				return config.DefaultAdminConfiguration{}, nil
			})

			require.NoError(t, err)
		})
	}
}
