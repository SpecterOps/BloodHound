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
package storage_test

import (
	"context"
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
		buildServices func(t *testing.T, controller *gomock.Controller) api_storage.FileServiceMap
		expected      expected
	}

	tests := []testData{
		{
			name: "creates resolver with services",
			buildServices: func(t *testing.T, controller *gomock.Controller) api_storage.FileServiceMap {
				return api_storage.FileServiceMap{
					storage.FileServiceWork: mocks.NewMockFileService(controller),
				}
			},
		},
		{
			name: "empty service name returns error",
			buildServices: func(t *testing.T, controller *gomock.Controller) api_storage.FileServiceMap {
				return api_storage.FileServiceMap{
					"": mocks.NewMockFileService(controller),
				}
			},
			expected: expected{
				errContains: "file service name is required",
			},
		},
		{
			name: "nil service returns error",
			buildServices: func(t *testing.T, controller *gomock.Controller) api_storage.FileServiceMap {
				return api_storage.FileServiceMap{
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
	services := api_storage.FileServiceMap{
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
		services    = api_storage.FileServiceMap{
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
	fileServices, err := api_storage.NewDefaultFileServices(context.Background(), configuration)

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

func TestNewDefaultFileServices_ConfiguredS3Provider(t *testing.T) {
	// Arrange
	var (
		workDir            = t.TempDir()
		collectorsBasePath = t.TempDir()
		configuration      = newTestStorageConfiguration(workDir, collectorsBasePath)
	)

	require.NoError(t, os.MkdirAll(configuration.TempDirectory(), 0o750))
	require.NoError(t, os.MkdirAll(configuration.RetainedFilesDirectory(), 0o750))

	configuration.Storage = config.StorageConfiguration{
		InstanceBucket: config.BucketConfiguration{
			Name:   "test-bucket",
			Region: "us-east-1",
		},
		FileServices: map[string]config.FileServiceConfiguration{
			string(storage.FileServiceIngest): {
				Provider: "s3",
				Prefix:   "/test/ingest/",
			},
		},
	}

	// Act
	fileServices, err := api_storage.NewDefaultFileServices(context.Background(), configuration)

	// Assert
	require.NoError(t, err)
	require.Len(t, fileServices, 4)

	ingestFileService, ok := fileServices[storage.FileServiceIngest].(*storage.StorageFileService)
	require.True(t, ok)
	require.IsType(t, &storage.Store{}, ingestFileService.Storage)

	for serviceName, fileService := range fileServices {
		if serviceName == storage.FileServiceIngest {
			continue
		}

		storageFileService, ok := fileService.(*storage.StorageFileService)
		require.True(t, ok)

		localStore, ok := storageFileService.Storage.(*storage.LocalStore)
		require.True(t, ok)
		require.NoError(t, localStore.Close())
	}
}

func TestNewDefaultFileServices_RejectsInvalidConfiguration(t *testing.T) {
	t.Parallel()

	type testData struct {
		name                  string
		storageConfiguration  config.StorageConfiguration
		additionalDefinitions []api_storage.FileServiceDefinition
		errContains           string
	}

	validBucketConfiguration := config.BucketConfiguration{
		Name:   "test-bucket",
		Region: "us-east-1",
	}

	tests := []testData{
		{
			name: "unsupported provider",
			storageConfiguration: config.StorageConfiguration{
				FileServices: map[string]config.FileServiceConfiguration{
					string(storage.FileServiceIngest): {Provider: "unknown"},
				},
			},
			errContains: `file service "ingest": unsupported file service provider "unknown"`,
		},
		{
			name: "missing S3 prefix",
			storageConfiguration: config.StorageConfiguration{
				InstanceBucket: validBucketConfiguration,
				FileServices: map[string]config.FileServiceConfiguration{
					string(storage.FileServiceIngest): {Provider: "s3"},
				},
			},
			errContains: `file service "ingest": s3 prefix is required`,
		},
		{
			name: "missing bucket name",
			storageConfiguration: config.StorageConfiguration{
				InstanceBucket: config.BucketConfiguration{Region: "us-east-1"},
				FileServices: map[string]config.FileServiceConfiguration{
					string(storage.FileServiceIngest): {Provider: "s3", Prefix: "ingest"},
				},
			},
			errContains: "storage.instance_bucket.name is required",
		},
		{
			name: "missing bucket region",
			storageConfiguration: config.StorageConfiguration{
				InstanceBucket: config.BucketConfiguration{Name: "test-bucket"},
				FileServices: map[string]config.FileServiceConfiguration{
					string(storage.FileServiceIngest): {Provider: "s3", Prefix: "ingest"},
				},
			},
			errContains: "storage.instance_bucket.region is required",
		},
		{
			name: "unknown file service",
			storageConfiguration: config.StorageConfiguration{
				FileServices: map[string]config.FileServiceConfiguration{
					"unknown": {Provider: "local"},
				},
			},
			errContains: `configuration references unknown file service "unknown"`,
		},
		{
			name: "overlapping S3 prefixes",
			storageConfiguration: config.StorageConfiguration{
				InstanceBucket: validBucketConfiguration,
				FileServices: map[string]config.FileServiceConfiguration{
					string(storage.FileServiceIngest):   {Provider: "s3", Prefix: "files"},
					string(storage.FileServiceRetained): {Provider: "s3", Prefix: "files/retained"},
				},
			},
			errContains: "overlap",
		},
		{
			name:                 "duplicate service definition",
			storageConfiguration: config.StorageConfiguration{},
			additionalDefinitions: []api_storage.FileServiceDefinition{
				{Name: storage.FileServiceIngest, LocalPath: t.TempDir()},
			},
			errContains: `duplicate file service definition "ingest"`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			configuration := config.Configuration{Storage: testCase.storageConfiguration}

			fileServices, err := api_storage.NewDefaultFileServices(
				context.Background(),
				configuration,
				testCase.additionalDefinitions...,
			)

			require.ErrorContains(t, err, testCase.errContains)
			require.Nil(t, fileServices)
		})
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
			fileServices, err := api_storage.NewDefaultFileServices(context.Background(), configuration)

			// Assert
			require.Error(t, err)
			require.Nil(t, fileServices)
		})
	}
}
