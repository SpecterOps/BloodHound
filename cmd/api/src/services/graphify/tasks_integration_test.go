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
//go:build integration

package graphify_test

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/stretchr/testify/require"
)

func TestVersion5IngestJSON(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version5JSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "computers.json"),
			path.Join(testSuite.WorkDir, "containers.json"),
			path.Join(testSuite.WorkDir, "domains.json"),
			path.Join(testSuite.WorkDir, "gpos.json"),
			path.Join(testSuite.WorkDir, "groups.json"),
			path.Join(testSuite.WorkDir, "ous.json"),
			path.Join(testSuite.WorkDir, "sessions.json"),
			path.Join(testSuite.WorkDir, "users.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version5JSON", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion5IngestZIP(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version5ZIP", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "archive.zip"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeZip}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 8, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version5ZIP", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6ADCSJSON(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6ADCSJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "aiacas.json"),
			path.Join(testSuite.WorkDir, "certtemplates.json"),
			path.Join(testSuite.WorkDir, "computers.json"),
			path.Join(testSuite.WorkDir, "containers.json"),
			path.Join(testSuite.WorkDir, "domains.json"),
			path.Join(testSuite.WorkDir, "enterprisecas.json"),
			path.Join(testSuite.WorkDir, "gpos.json"),
			path.Join(testSuite.WorkDir, "groups.json"),
			path.Join(testSuite.WorkDir, "issuancepolicies.json"),
			path.Join(testSuite.WorkDir, "ntauthstores.json"),
			path.Join(testSuite.WorkDir, "ous.json"),
			path.Join(testSuite.WorkDir, "rootcas.json"),
			path.Join(testSuite.WorkDir, "users.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6ADCSJSON", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6ADCSZIP(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6ADCSZIP", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "archive.zip"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeZip}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 13, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6ADCSZIP", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6AllJSON(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6AllJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "aiacas.json"),
			path.Join(testSuite.WorkDir, "certtemplates.json"),
			path.Join(testSuite.WorkDir, "computers.json"),
			path.Join(testSuite.WorkDir, "containers.json"),
			path.Join(testSuite.WorkDir, "domains.json"),
			path.Join(testSuite.WorkDir, "enterprisecas.json"),
			path.Join(testSuite.WorkDir, "gpos.json"),
			path.Join(testSuite.WorkDir, "groups.json"),
			path.Join(testSuite.WorkDir, "issuancepolicies.json"),
			path.Join(testSuite.WorkDir, "ntauthstores.json"),
			path.Join(testSuite.WorkDir, "ous.json"),
			path.Join(testSuite.WorkDir, "rootcas.json"),
			path.Join(testSuite.WorkDir, "users.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6AllJSON", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6AllZIP(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6AllZIP", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "archive.zip"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeZip}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 13, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6AllZIP", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6IngestJSON(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6JSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "computers.json"),
			path.Join(testSuite.WorkDir, "containers.json"),
			path.Join(testSuite.WorkDir, "domains.json"),
			path.Join(testSuite.WorkDir, "gpos.json"),
			path.Join(testSuite.WorkDir, "groups.json"),
			path.Join(testSuite.WorkDir, "ous.json"),
			path.Join(testSuite.WorkDir, "sessions.json"),
			path.Join(testSuite.WorkDir, "users.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6JSON", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6IngestZIP(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "Version6ZIP", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "archive.zip"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeZip}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 8, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6ZIP", "ingest")), "ingested.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}
