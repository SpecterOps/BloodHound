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
	"path/filepath"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/bootstrap"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/services"
	"github.com/specterops/bloodhound/cmd/api/src/services/storage"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func runtimeDependencyTestConfig(t *testing.T) config.Configuration {
	t.Helper()

	rootDirectory := t.TempDir()

	return config.Configuration{
		WorkDir:            filepath.Join(rootDirectory, "work"),
		CollectorsBasePath: filepath.Join(rootDirectory, "collectors"),
	}
}

func TestCreateRuntimeDependenciesInitializesDefaultFileServices(t *testing.T) {
	t.Parallel()

	var (
		cfg         = runtimeDependencyTestConfig(t)
		connections = bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch]{}
	)

	require.NoError(t, bootstrap.EnsureServerDirectories(cfg))

	deps, err := services.CreateRuntimeDependencies(context.Background(), cfg, connections)

	require.NoError(t, err)
	require.NotNil(t, deps.FileServiceResolver)

	for _, serviceName := range []storage.FileServiceName{
		storage.FileServiceWork,
		storage.FileServiceIngest,
		storage.FileServiceRetained,
		storage.FileServiceCollectors,
	} {
		fileService, err := deps.FileServiceResolver.Resolve(serviceName)
		require.NoError(t, err)
		require.NotNil(t, fileService)
	}
}
