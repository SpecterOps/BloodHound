//go:build slow_integration

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

func TestVersion6Analysis(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		ingestFilePath = path.Join("fixtures", "TestVersion6IngestJSON")

		testSuite = setupIntegrationTestSuite(t, ingestFilePath)
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	expected, err := generic.LoadGraphFromFile(os.DirFS(testSuite.WorkDir), "v6expected.json")
	require.NoError(t, err)

	err = generic.WriteGraphToDatabase(testSuite.GraphDB, &expected)
	require.NoError(t, err)

	err = datapipe.RunAnalysisOperations(ctx, testSuite.BHDatabase, testSuite.GraphDB, config.Configuration{})
	require.NoError(t, err)

	expected, err = generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", "TestVersion6IngestJSON")), "analyzed.json")
	require.NoError(t, err)

	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}
