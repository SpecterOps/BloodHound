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
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/util/size"
	schema "github.com/specterops/bloodhound/graphschema"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/cmd/dawgs-harness/tests"
)

func fatalf(format string, args ...any) {
	fmt.Printf(format, args...)
	os.Exit(1)
}

func RunTestSuite(ctx context.Context, connectionStr, driverName string) tests.TestSuite {
	if connection, err := dawgs.Open(context.TODO(), driverName, dawgs.Config{
		TraversalMemoryLimit: size.Gibibyte,
		DriverCfg:            connectionStr,
	}); err != nil {
		fatalf("Failed opening %s database: %v", driverName, err)
	} else {
		defer connection.Close(ctx)

		if err := connection.AssertSchema(ctx, schema.DefaultGraphSchema()); err != nil {
			fatalf("Failed asserting graph schema on %s database: %v", driverName, err)
		} else if err := connection.WriteTransaction(ctx, func(tx graph.Transaction) error {
			return tx.Nodes().Delete()
		}); err != nil {
			fatalf("Failed to clear %s database: %v", driverName, err)
		} else if testSuite, err := tests.RunSuite(connection, driverName); err != nil {
			fatalf("Test suite error for %s database: %v", driverName, err)
		} else {
			return testSuite
		}
	}

	panic(nil)
}

func newContext() context.Context {
	var (
		ctx, done = context.WithCancel(context.Background())
		sigchnl   = make(chan os.Signal, 1)
	)

	signal.Notify(sigchnl)

	go func() {
		defer done()

		for nextSignal := range sigchnl {
			switch nextSignal {
			case syscall.SIGINT, syscall.SIGTERM:
				return
			}
		}
	}()

	return ctx
}

var enablePprof bool

func execSuite(name string, logic func() tests.TestSuite) tests.TestSuite {
	if enablePprof {
		if cpuProfileFile, err := os.OpenFile(name+".pprof", syscall.O_WRONLY|syscall.O_TRUNC|syscall.O_CREAT, 0644); err != nil {
			fatalf("Unable to open file for CPU profile: %v", err)
		} else {
			defer cpuProfileFile.Close()

			if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
				fatalf("Failed to start CPU profile: %v", err)
			} else {
				defer pprof.StopCPUProfile()
			}
		}
	}

	return logic()
}

func main() {
	var (
		ctx                = newContext()
		neo4jConnectionStr string
		pgConnectionStr    string
		testType           string
	)

	flag.StringVar(&testType, "test", "both", "Test to run. Must be one of: 'postgres', 'neo4j', 'both'")
	flag.BoolVar(&enablePprof, "enable-pprof", true, "Enable the pprof HTTP sampling server.")
	flag.IntVar(&tests.SimpleRelationshipsToCreate, "num-rels", 2000, "Number of simple relationships to create.")
	flag.StringVar(&neo4jConnectionStr, "neo4j", "neo4j://neo4j:neo4jj@localhost:7687", "Neo4j connection string.")
	flag.StringVar(&pgConnectionStr, "pg", "user=bhe dbname=bhe password=bhe4eva host=localhost", "PostgreSQL connection string.")
	flag.Parse()

	log.ConfigureDefaults()

	switch testType {
	case "both":
		n4jTestSuite := execSuite(neo4j.DriverName, func() tests.TestSuite {
			return RunTestSuite(ctx, neo4jConnectionStr, neo4j.DriverName)
		})

		fmt.Println()

		// Sleep between tests
		time.Sleep(time.Second * 3)

		pgTestSuite := execSuite(pg.DriverName, func() tests.TestSuite {
			return RunTestSuite(ctx, pgConnectionStr, pg.DriverName)
		})
		fmt.Println()

		OutputTestSuiteDeltas(pgTestSuite, n4jTestSuite)

	case "postgres":
		execSuite(pg.DriverName, func() tests.TestSuite {
			return RunTestSuite(ctx, pgConnectionStr, pg.DriverName)
		})

	case "neo4j":
		execSuite(neo4j.DriverName, func() tests.TestSuite {
			return RunTestSuite(ctx, neo4jConnectionStr, neo4j.DriverName)
		})
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
