// Copyright 2025 Specter Ops, Inc.
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
//go:build slow_integration

package graphify_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	"github.com/specterops/bloodhound/cmd/api/src/services/storage"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/stretchr/testify/require"
)

func TestVersion5IngestJSON(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version5JSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			"computers.json",
			"containers.json",
			"domains.json",
			"gpos.json",
			"groups.json",
			"ous.json",
			"sessions.json",
			"users.json",
		}
	)
	ingestLocalStore, err := storage.NewLocalStore(testSuite.WorkDir)
	require.NoError(t, err, "error creating ingest local store")
	ingestFileService := storage.NewFileService(ingestLocalStore)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		ingestCtx := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestCtx, ingestFileService, model.IngestTask{StoredFileName: file, FileType: model.FileTypeJson})
		require.NoError(t, err)

		failed := 0
		for _, data := range fileData {
			if len(data.Errors) > 0 {
				failed++
			}
		}

		require.Zero(t, failed)
		require.Equal(t, 1, len(fileData))
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version5JSON", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion5IngestZIP(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version5ZIP", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			"archive.zip",
		}
	)
	ingestLocalStore, err := storage.NewLocalStore(testSuite.WorkDir)
	require.NoError(t, err, "error creating ingest local store")
	ingestFileService := storage.NewFileService(ingestLocalStore)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		ingestCtx := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestCtx, ingestFileService, model.IngestTask{StoredFileName: file, FileType: model.FileTypeZip})
		require.NoError(t, err)

		failed := 0
		for _, data := range fileData {
			if len(data.Errors) > 0 {
				failed++
			}
		}

		require.Zero(t, failed)
		require.Equal(t, 8, len(fileData))
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version5ZIP", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6ADCSJSON(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6ADCSJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			"aiacas.json",
			"certtemplates.json",
			"computers.json",
			"containers.json",
			"domains.json",
			"enterprisecas.json",
			"gpos.json",
			"groups.json",
			"issuancepolicies.json",
			"ntauthstores.json",
			"ous.json",
			"rootcas.json",
			"users.json",
		}
	)
	ingestLocalStore, err := storage.NewLocalStore(testSuite.WorkDir)
	require.NoError(t, err, "error creating ingest local store")
	ingestFileService := storage.NewFileService(ingestLocalStore)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		ingestCtx := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestCtx, ingestFileService, model.IngestTask{StoredFileName: file, FileType: model.FileTypeJson})
		require.NoError(t, err)

		failed := 0
		for _, data := range fileData {
			if len(data.Errors) > 0 {
				failed++
			}
		}

		require.Zero(t, failed)
		require.Equal(t, 1, len(fileData))
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6ADCSJSON", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6ADCSZIP(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6ADCSZIP", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			"archive.zip",
		}
	)
	ingestLocalStore, err := storage.NewLocalStore(testSuite.WorkDir)
	require.NoError(t, err, "error creating ingest local store")
	ingestFileService := storage.NewFileService(ingestLocalStore)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		ingestCtx := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestCtx, ingestFileService, model.IngestTask{StoredFileName: file, FileType: model.FileTypeZip})
		require.NoError(t, err)

		failed := 0
		for _, data := range fileData {
			if len(data.Errors) > 0 {
				failed++
			}
		}

		require.Zero(t, failed)
		require.Equal(t, 13, len(fileData))
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6ADCSZIP", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6AllJSON(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6AllJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			"aiacas.json",
			"certtemplates.json",
			"computers.json",
			"containers.json",
			"domains.json",
			"enterprisecas.json",
			"gpos.json",
			"groups.json",
			"issuancepolicies.json",
			"ntauthstores.json",
			"ous.json",
			"rootcas.json",
			"users.json",
		}
	)
	ingestLocalStore, err := storage.NewLocalStore(testSuite.WorkDir)
	require.NoError(t, err, "error creating ingest local store")
	ingestFileService := storage.NewFileService(ingestLocalStore)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		ingestContext := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestContext, ingestFileService, model.IngestTask{StoredFileName: file, FileType: model.FileTypeJson})
		require.NoError(t, err)

		failed := 0
		for _, data := range fileData {
			if len(data.Errors) > 0 {
				failed++
			}
		}

		require.Zero(t, failed)
		require.Equal(t, 1, len(fileData))
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6AllJSON", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6AllZIP(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6AllZIP", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			"archive.zip",
		}
	)
	ingestLocalStore, err := storage.NewLocalStore(testSuite.WorkDir)
	require.NoError(t, err, "error creating ingest local store")
	ingestFileService := storage.NewFileService(ingestLocalStore)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		ingestContext := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestContext, ingestFileService, model.IngestTask{StoredFileName: file, FileType: model.FileTypeZip})
		require.NoError(t, err)

		failed := 0
		for _, data := range fileData {
			if len(data.Errors) > 0 {
				failed++
			}
		}

		require.Zero(t, failed)
		require.Equal(t, 13, len(fileData))
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6AllZIP", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6IngestJSON(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6JSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			"computers.json",
			"containers.json",
			"domains.json",
			"gpos.json",
			"groups.json",
			"ous.json",
			"sessions.json",
			"users.json",
		}
	)
	ingestLocalStore, err := storage.NewLocalStore(testSuite.WorkDir)
	require.NoError(t, err, "error creating ingest local store")
	ingestFileService := storage.NewFileService(ingestLocalStore)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		ingestContext := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestContext, ingestFileService, model.IngestTask{StoredFileName: file, FileType: model.FileTypeJson})
		require.NoError(t, err)

		failed := 0
		for _, data := range fileData {
			if len(data.Errors) > 0 {
				failed++
			}
		}

		require.Zero(t, failed)
		require.Equal(t, 1, len(fileData))
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6JSON", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6IngestZIP(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6ZIP", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			"archive.zip",
		}
	)
	ingestLocalStore, err := storage.NewLocalStore(testSuite.WorkDir)
	require.NoError(t, err, "error creating ingest local store")
	ingestFileService := storage.NewFileService(ingestLocalStore)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		ingestContext := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestContext, ingestFileService, model.IngestTask{StoredFileName: file, FileType: model.FileTypeZip})
		require.NoError(t, err)

		failed := 0
		for _, data := range fileData {
			if len(data.Errors) > 0 {
				failed++
			}
		}

		require.Zero(t, failed)
		require.Equal(t, 8, len(fileData))
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6ZIP", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}
