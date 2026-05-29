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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

// step 1. ingest a file
// step 2. delete sourceless
// step 3. assert graph
func TestDeleteData_Sourceless(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "OpenGraphJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "base.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)
	ingestCtx := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))

	for _, file := range files {
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestCtx, model.IngestTask{StoredFileName: file, FileType: model.FileTypeJson})
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
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "OpenGraphJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "base.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)
	ingestCtx := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))

	for _, file := range files {
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestCtx, model.IngestTask{StoredFileName: file, FileType: model.FileTypeJson})
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
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "OpenGraphJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "base.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)
	ingestCtx := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))

	for _, file := range files {
		fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestCtx, model.IngestTask{StoredFileName: file, FileType: model.FileTypeJson})
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

// TestPartialIngest verifies that when a batch contains one resolvable edge and one unresolvable
// edge, the resolvable edge is still committed and the resolution failure is surfaced as a
// UserDataErr rather than a fatal error that rolls back the batch.
func TestPartialIngest(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", "OpenGraphJSON", "raw")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)
	)

	defer teardownIntegrationTestSuite(t, &testSuite)
	ingestCtx := graphify.NewIngestContext(ctx, graphify.WithEndpointResolver(endpoint.NewResolver(testSuite.GraphDB)))

	fileData, err := testSuite.GraphifyService.ProcessIngestFile(ingestCtx, model.IngestTask{
		StoredFileName: path.Join(testSuite.WorkDir, "oneGoodOneInvalidRel.json"),
		FileType:       model.FileTypeJson,
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(fileData))

	require.Empty(t, fileData[0].Errors, "no fatal errors expected; resolution failures must not abort the batch")
	require.Len(t, fileData[0].UserDataErrs, 1, "expected one resolution error for the unresolvable property_match")

	// Verify that the edge for the existing endpoints was created despite the failing edge existing in the same batch.
	var edgeCount int
	err = testSuite.GraphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Relationships().Filterf(func() graph.Criteria {
			return query.Kind(query.Relationship(), graph.StringKind("THIS_EDGE_GETS_CREATED"))
		}).Fetch(func(cursor graph.Cursor[*graph.Relationship]) error {
			for range cursor.Chan() {
				edgeCount++
			}
			return cursor.Error()
		})
	})
	require.NoError(t, err)
	require.Equal(t, 1, edgeCount, "THIS_EDGE_GETS_CREATED was not created")
}

func TestAnalyze_LastAnalysisTimestampUpdated(t *testing.T) {
	var (
		ctx              = context.Background()
		ingestedFilePath = path.Join("fixtures", "OpenGraphJSON", "raw")
		testSuite        = setupIntegrationTestSuite(t, ingestedFilePath)
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	datapipeStatus, err := testSuite.BHDatabase.GetDatapipeStatus(ctx)
	require.NoError(t, err)
	require.True(t, datapipeStatus.LastAnalysisRunAt.IsZero())

	// request analysis so that Analyze will run
	err = testSuite.BHDatabase.RequestAnalysis(ctx, "test")
	require.NoError(t, err)

	err = testSuite.Daemon.Analyze(ctx)
	require.NoError(t, err)

	// confirm that the last analysis run timestamp is updated
	updatedDatapipeStatus, err := testSuite.BHDatabase.GetDatapipeStatus(ctx)
	require.NoError(t, err)
	require.False(t, updatedDatapipeStatus.LastAnalysisRunAt.IsZero())
	require.Greater(t, updatedDatapipeStatus.LastAnalysisRunAt, datapipeStatus.LastAnalysisRunAt)
}
