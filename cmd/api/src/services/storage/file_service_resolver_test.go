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
package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	api_storage "github.com/specterops/bloodhound/cmd/api/src/services/storage"
	"github.com/specterops/bloodhound/packages/go/storage"
	"github.com/specterops/bloodhound/packages/go/storage/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestStorageConfiguration(workDir string, collectorsBasePath string) config.Configuration {
	return config.Configuration{
		WorkDir:            workDir,
		CollectorsBasePath: collectorsBasePath,
	}
}

func TestNewFileServiceResolver(t *testing.T) {
	t.Parallel()

	type expected struct {
		errContains string
	}

	type testData struct {
		name          string
		buildServices func(t *testing.T, controller *gomock.Controller) map[storage.FileServiceName]storage.FileService
		expected      expected
	}

	tests := []testData{
		{
			name: "creates resolver with services",
			buildServices: func(t *testing.T, controller *gomock.Controller) map[storage.FileServiceName]storage.FileService {
				return map[storage.FileServiceName]storage.FileService{
					storage.FileServiceWork: mocks.NewMockFileService(controller),
				}
			},
		},
		{
			name: "empty service name returns error",
			buildServices: func(t *testing.T, controller *gomock.Controller) map[storage.FileServiceName]storage.FileService {
				return map[storage.FileServiceName]storage.FileService{
					"": mocks.NewMockFileService(controller),
				}
			},
			expected: expected{
				errContains: "file service name is required",
			},
		},
		{
			name: "nil service returns error",
			buildServices: func(t *testing.T, controller *gomock.Controller) map[storage.FileServiceName]storage.FileService {
				return map[storage.FileServiceName]storage.FileService{
					storage.FileServiceWork: nil,
				}
			},
			expected: expected{
				errContains: `file service "work" is nil`,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			services := testCase.buildServices(t, gomock.NewController(t))

			// Act
			resolver, err := api_storage.NewFileServiceResolver(services)

			// Assert
			if testCase.expected.errContains != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expected.errContains)
				require.Nil(t, resolver)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resolver)
		})
	}
}

func TestFileServiceResolver_Resolve(t *testing.T) {
	t.Parallel()

	type expected struct {
		errIs   error
		service storage.FileService
	}

	type testData struct {
		name        string
		resolveName storage.FileServiceName
		expected    expected
	}

	controller := gomock.NewController(t)
	workService := mocks.NewMockFileService(controller)
	services := map[storage.FileServiceName]storage.FileService{
		storage.FileServiceWork: workService,
	}

	tests := []testData{
		{
			name:        "resolves service",
			resolveName: storage.FileServiceWork,
			expected: expected{
				service: workService,
			},
		},
		{
			name:        "missing service returns not found",
			resolveName: storage.FileServiceIngest,
			expected: expected{
				errIs: storage.ErrFileServiceNotFound,
			},
		},
		{
			name:        "empty name returns not found",
			resolveName: "",
			expected: expected{
				errIs: storage.ErrFileServiceNotFound,
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			resolver, err := api_storage.NewFileServiceResolver(services)
			require.NoError(t, err)

			// Act
			fileService, err := resolver.Resolve(testCase.resolveName)

			// Assert
			if testCase.expected.errIs != nil {
				require.ErrorIs(t, err, testCase.expected.errIs)
				require.Nil(t, fileService)
				return
			}

			require.NoError(t, err)
			require.Same(t, testCase.expected.service, fileService)
		})
	}
}

func TestFileServiceResolver_CopiesServices(t *testing.T) {
	t.Parallel()

	// Arrange
	var (
		controller  = gomock.NewController(t)
		workService = mocks.NewMockFileService(controller)
		services    = map[storage.FileServiceName]storage.FileService{
			storage.FileServiceWork: workService,
		}
	)

	resolver, err := api_storage.NewFileServiceResolver(services)
	require.NoError(t, err)

	delete(services, storage.FileServiceWork)

	// Act
	fileService, err := resolver.Resolve(storage.FileServiceWork)

	// Assert
	require.NoError(t, err)
	require.Same(t, workService, fileService)
}

func TestNewDefaultFileServices(t *testing.T) {
	t.Parallel()

	// Arrange
	var (
		workDir            = t.TempDir()
		collectorsBasePath = t.TempDir()
		configuration      = newTestStorageConfiguration(workDir, collectorsBasePath)
	)

	require.NoError(t, os.MkdirAll(configuration.TempDirectory(), 0o750))
	require.NoError(t, os.MkdirAll(configuration.RetainedFilesDirectory(), 0o750))

	// Act
	fileServices, err := api_storage.NewDefaultFileServices(configuration)

	// Assert
	require.NoError(t, err)
	require.Contains(t, fileServices, storage.FileServiceWork)
	require.Contains(t, fileServices, storage.FileServiceIngest)
	require.Contains(t, fileServices, storage.FileServiceRetained)
	require.Contains(t, fileServices, storage.FileServiceCollectors)
	require.Len(t, fileServices, 4)

	for _, fileService := range fileServices {
		storageFileService, ok := fileService.(*storage.StorageFileService)
		require.True(t, ok)

		localStore, ok := storageFileService.Storage.(*storage.LocalStore)
		require.True(t, ok)
		require.NoError(t, localStore.Close())
	}
}

func TestNewDefaultFileServices_ReturnsError(t *testing.T) {
	t.Parallel()

	type testData struct {
		name  string
		setup func(t *testing.T) config.Configuration
	}

	tests := []testData{
		{
			name: "missing work directory",
			setup: func(t *testing.T) config.Configuration {
				return newTestStorageConfiguration(filepath.Join(t.TempDir(), "missing"), t.TempDir())
			},
		},
		{
			name: "missing temp directory",
			setup: func(t *testing.T) config.Configuration {
				workDir := t.TempDir()

				return newTestStorageConfiguration(workDir, t.TempDir())
			},
		},
		{
			name: "missing retained directory",
			setup: func(t *testing.T) config.Configuration {
				workDir := t.TempDir()
				configuration := newTestStorageConfiguration(workDir, t.TempDir())
				require.NoError(t, os.MkdirAll(configuration.TempDirectory(), 0o750))

				return configuration
			},
		},
		{
			name: "missing collectors directory",
			setup: func(t *testing.T) config.Configuration {
				workDir := t.TempDir()
				configuration := newTestStorageConfiguration(workDir, filepath.Join(t.TempDir(), "missing"))
				require.NoError(t, os.MkdirAll(configuration.TempDirectory(), 0o750))
				require.NoError(t, os.MkdirAll(configuration.RetainedFilesDirectory(), 0o750))

				return configuration
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			configuration := testCase.setup(t)

			// Act
			fileServices, err := api_storage.NewDefaultFileServices(configuration)

			// Assert
			require.Error(t, err)
			require.Nil(t, fileServices)
		})
	}
}
