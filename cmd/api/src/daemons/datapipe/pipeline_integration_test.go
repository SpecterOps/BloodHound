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

package datapipe_test

import (
	"context"
	"os"
	"path"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

// step 1. ingest a file
// step 2. delete sourceless
// step 3. assert graph
func TestDeleteData_Sourceless(t *testing.T) {

	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "OpenGraphJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "base.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)
	for _, file := range files {
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
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

	// simulate requesting deletion
	err := testSuite.BHDatabase.RegisterSourceKind(ctx)(graph.StringKind("GithubBase"))
	require.Nil(t, err)
	testSuite.BHDatabase.RequestCollectedGraphDataDeletion(ctx, model.AnalysisRequest{DeleteSourcelessGraph: true, RequestType: model.AnalysisRequestDeletion})
	require.Nil(t, err)
	testSuite.Daemon.DeleteData(ctx)

	// verify sourceless nodes are deleted
	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "OpenGraphJSON", "deleted")), "expected_sourceless.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)

	// because no source_kind was specified for deletion, ensure integrity of records
	sourceKinds, err := testSuite.BHDatabase.GetSourceKinds(ctx)
	require.NoError(t, err)
	require.Len(t, sourceKinds, 3)
}

func TestDeleteData_SourceKinds(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "OpenGraphJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "base.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
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

	// simulate requesting deletion
	err := testSuite.BHDatabase.RegisterSourceKind(ctx)(graph.StringKind("GithubBase"))
	require.Nil(t, err)
	testSuite.BHDatabase.RequestCollectedGraphDataDeletion(ctx, model.AnalysisRequest{DeleteSourceKinds: []string{"GithubBase", "AZBase"}, RequestType: model.AnalysisRequestDeletion})
	require.Nil(t, err)
	testSuite.Daemon.DeleteData(ctx)

	// verify nodes are deleted by source_kind
	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "OpenGraphJSON", "deleted")), "expected_sourcekinds.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)

	// ensure GithubBase source_kind is removed from table
	sourceKinds, err := testSuite.BHDatabase.GetSourceKinds(ctx)
	require.NoError(t, err)
	require.Len(t, sourceKinds, 2)
}

func TestDeleteData_All(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "OpenGraphJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "base.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
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

	// simulate requesting deletion
	err := testSuite.BHDatabase.RegisterSourceKind(ctx)(graph.StringKind("GithubBase"))
	require.Nil(t, err)
	testSuite.BHDatabase.RequestCollectedGraphDataDeletion(ctx, model.AnalysisRequest{DeleteSourceKinds: []string{"GithubBase", "AZBase", "Base"}, RequestType: model.AnalysisRequestDeletion})
	require.Nil(t, err)
	testSuite.Daemon.DeleteData(ctx)

	// verify all nodes are deleted
	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "OpenGraphJSON", "deleted")), "expected_all.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)

	// ensure GithubBase source_kind is removed from table
	sourceKinds, err := testSuite.BHDatabase.GetSourceKinds(ctx)
	require.NoError(t, err)
	require.Len(t, sourceKinds, 2)
}
