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

	"github.com/specterops/bloodhound/cmd/api/src/daemons/datapipe"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestDeleteAllData(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", t.Name(), "opengraph")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "base.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	// DELETE ALL
	var (
		deleteRequest = model.AnalysisRequest{DeleteAllGraph: true}
		sourceKinds   = []graph.Kind{graph.StringKind("Base"), graph.StringKind("AZBase"), graph.StringKind("GithubBase")}
	)

	err := datapipe.DeleteCollectedGraphData2(ctx, testSuite.GraphDB, deleteRequest, sourceKinds)
	require.Nil(t, err)

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", t.Name())), "deleteAllExpected.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestDeleteSourcelessData(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", t.Name(), "opengraph")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "base.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	var (
		deleteRequest = model.AnalysisRequest{DeleteSourcelessGraph: true}
		sourceKinds   = []graph.Kind{graph.StringKind("Base"), graph.StringKind("AZBase"), graph.StringKind("GithubBase")}
	)

	err := datapipe.DeleteCollectedGraphData2(ctx, testSuite.GraphDB, deleteRequest, sourceKinds)
	require.Nil(t, err)

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", t.Name())), "deleteSourcelessExpected.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestDeleteSourceKindsData(t *testing.T) {
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", t.Name(), "opengraph")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "base.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	var (
		deleteRequest = model.AnalysisRequest{DeleteSourceKinds: []string{"AZBase", "GithubBase"}}
		sourceKinds   = []graph.Kind{graph.StringKind("Base"), graph.StringKind("AZBase"), graph.StringKind("GithubBase")}
	)

	err := datapipe.DeleteCollectedGraphData2(ctx, testSuite.GraphDB, deleteRequest, sourceKinds)
	require.Nil(t, err)

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", t.Name())), "deleteSourceKindsExpected.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}
