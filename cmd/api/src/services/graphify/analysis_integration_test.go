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

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/datapipe"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/stretchr/testify/require"
)

func TestVersion5Analysis(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		ingestFilePath = path.Join("fixtures", "Version5JSON", "ingest")

		testSuite = setupIntegrationTestSuite(t, ingestFilePath)
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	expected, err := generic.LoadGraphFromFile(os.DirFS(testSuite.WorkDir), "ingested.json")
	require.NoError(t, err)

	err = generic.WriteGraphToDatabase(testSuite.GraphDB, &expected)
	require.NoError(t, err)

	err = datapipe.RunAnalysisOperations(ctx, testSuite.BHDatabase, testSuite.GraphDB, config.Configuration{})
	require.NoError(t, err)

	expected, err = generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version5JSON", "analysis")), "analyzed.json")
	require.NoError(t, err)

	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6ADCSAnalysis(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		ingestFilePath = path.Join("fixtures", "Version6ADCSJSON", "ingest")

		testSuite = setupIntegrationTestSuite(t, ingestFilePath)
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	expected, err := generic.LoadGraphFromFile(os.DirFS(testSuite.WorkDir), "ingested.json")
	require.NoError(t, err)

	err = generic.WriteGraphToDatabase(testSuite.GraphDB, &expected)
	require.NoError(t, err)

	err = datapipe.RunAnalysisOperations(ctx, testSuite.BHDatabase, testSuite.GraphDB, config.Configuration{})
	require.NoError(t, err)

	expected, err = generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6ADCSJSON", "analysis")), "analyzed.json")
	require.NoError(t, err)

	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6AllAnalysis(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		ingestFilePath = path.Join("fixtures", "Version6AllJSON", "ingest")

		testSuite = setupIntegrationTestSuite(t, ingestFilePath)
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	expected, err := generic.LoadGraphFromFile(os.DirFS(testSuite.WorkDir), "ingested.json")
	require.NoError(t, err)

	err = generic.WriteGraphToDatabase(testSuite.GraphDB, &expected)
	require.NoError(t, err)

	err = datapipe.RunAnalysisOperations(ctx, testSuite.BHDatabase, testSuite.GraphDB, config.Configuration{})
	require.NoError(t, err)

	expected, err = generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6AllJSON", "analysis")), "analyzed.json")
	require.NoError(t, err)

	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6Analysis(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		ingestFilePath = path.Join("fixtures", "Version6JSON", "ingest")

		testSuite = setupIntegrationTestSuite(t, ingestFilePath)
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	expected, err := generic.LoadGraphFromFile(os.DirFS(testSuite.WorkDir), "ingested.json")
	require.NoError(t, err)

	err = generic.WriteGraphToDatabase(testSuite.GraphDB, &expected)
	require.NoError(t, err)

	err = datapipe.RunAnalysisOperations(ctx, testSuite.BHDatabase, testSuite.GraphDB, config.Configuration{})
	require.NoError(t, err)

	expected, err = generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "Version6JSON", "analysis")), "analyzed.json")
	require.NoError(t, err)

	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}
