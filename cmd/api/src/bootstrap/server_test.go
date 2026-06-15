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
	"os"
	"path/filepath"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/bootstrap"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/stretchr/testify/require"
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
