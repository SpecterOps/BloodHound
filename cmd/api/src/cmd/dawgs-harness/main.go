// Copyright 2023 Specter Ops, Inc.
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

package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/cmd/dawgs-harness/tests"
)

func fatalf(format string, args ...any) {
	fmt.Printf(format, args...)
	os.Exit(1)
}

func RunNeo4jTestSuite(dbHost string) tests.TestSuite {
	if connection, err := dawgs.Open("neo4j", dawgs.Config{DriverCfg: fmt.Sprintf("neo4j://neo4j:neo4jj@%s:7687/neo4j", dbHost)}); err != nil {
		fatalf("Failed opening neo4j: %v", err)
	} else if err := connection.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		return tx.Nodes().Delete()
	}); err != nil {
		fatalf("Failed to clear neo4j: %v", err)
	} else if err := neo4j.AssertNodePropertyIndex(connection, ad.Entity, common.Name.String(), graph.BTreeIndex); err != nil {
		fatalf("Error creating database schema: %v", err)
	} else if testSuite, err := tests.RunSuite(tests.Neo4j, connection); err != nil {
		fatalf("Test suite error: %v", err)
	} else {
		connection.Close()
		return testSuite
	}

	panic("")
}

func main() {
	var (
		dbHost      string
		testType    string
		enablePprof bool
	)

	flag.StringVar(&testType, "test", "both", "Test to run. Must be one of: 'postgres', 'neo4j', 'both'")
	flag.BoolVar(&enablePprof, "enable-pprof", false, "Enable the pprof HTTP sampling server.")
	flag.IntVar(&tests.SimpleRelationshipsToCreate, "num-rels", 5000, "Number of simple relationships to create.")
	flag.StringVar(&dbHost, "db-host", "192.168.122.170", "Database host.")
	flag.Parse()

	if enablePprof {
		go func() {
			if err := http.ListenAndServe("localhost:8080", nil); err != nil {
				log.Error().Fault(err).Msg("HTTP server caught an error while running.")
			}
		}()
	}

	log.ConfigureDefaults()

	switch testType {
	//case "both":
	//	pgTestSuite := RunPostgresqlTestSuite(dbHost)
	//
	//	// Sleep between tests
	//	time.Sleep(time.Second * 3)
	//	fmt.Println()
	//
	//	n4jTestSuite := RunNeo4jTestSuite(dbHost)
	//	fmt.Println()
	//
	//	OutputTestSuiteDeltas(pgTestSuite, n4jTestSuite)
	//
	//case "postgres":
	//	RunPostgresqlTestSuite(dbHost)

	case "neo4j":
		RunNeo4jTestSuite(dbHost)
	}
}

func OutputTestSuiteDeltas(leftSuite, rightSuite tests.TestSuite) {
	tableWriter := table.NewWriter()
	tableWriter.AppendHeader(table.Row{"Test Case", "Suite", "Samples", "Test Case Duration", "Shortest Sample Duration", "Longest Sample Duration", "Average Sample Duration"})

	for _, leftTestCase := range leftSuite.Cases {
		var (
			testCaseName  = leftTestCase.Name
			rightTestCase = rightSuite.GetTestCase(testCaseName)
		)

		tableWriter.AppendRow(table.Row{
			testCaseName,
		})

		tableWriter.AppendRow(table.Row{
			"",
			leftSuite.Name,
			len(leftTestCase.Samples),
			leftTestCase.Duration,
			leftTestCase.Samples.ShortestDuration(),
			leftTestCase.Samples.LongestDuration(),
			leftTestCase.Samples.AverageDuration(),
		})

		tableWriter.AppendRow(table.Row{
			"",
			rightSuite.Name,
			len(rightTestCase.Samples),
			rightTestCase.Duration,
			rightTestCase.Samples.ShortestDuration(),
			rightTestCase.Samples.LongestDuration(),
			rightTestCase.Samples.AverageDuration(),
		})

		tableWriter.AppendRow(table.Row{
			"",
			"",
			"",
			durationValuesAsPercentageString(leftTestCase.Duration, rightTestCase.Duration),
			durationValuesAsPercentageString(leftTestCase.Samples.ShortestDuration(), rightTestCase.Samples.ShortestDuration()),
			durationValuesAsPercentageString(leftTestCase.Samples.LongestDuration(), rightTestCase.Samples.LongestDuration()),
			durationValuesAsPercentageString(leftTestCase.Samples.AverageDuration(), rightTestCase.Samples.AverageDuration()),
		})
	}

	fmt.Printf("%s\n", tableWriter.Render())
}

func durationValuesAsPercentageString(left, right time.Duration) string {
	return int64ValuesAsPercentageString(int64(left), int64(right))
}

func int64ValuesAsPercentageString(left, right int64) string {
	if percent := 1 - float64(left)/float64(right); percent < 0 {
		return fmt.Sprintf("%.2f%% slower", percent*-100)
	} else {
		return fmt.Sprintf("%.2f%% faster", percent*100)
	}
}
