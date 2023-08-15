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

package tests

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type TestDelegate func(testCase *TestCase) any

type Sample struct {
	Start time.Time
	End   time.Time
}

func (s Sample) Duration() time.Duration {
	return s.End.Sub(s.Start)
}

type Samples []Sample

func (s Samples) AverageDuration() time.Duration {
	var totalDuration time.Duration

	for _, sample := range s {
		totalDuration += sample.Duration()
	}

	return totalDuration / time.Duration(len(s))
}

func (s Samples) ShortestDuration() time.Duration {
	shortestDuration := s[0].Duration()

	for _, sample := range s[1:] {
		if sampleDuration := sample.Duration(); shortestDuration > sampleDuration {
			shortestDuration = sampleDuration
		}
	}

	return shortestDuration
}

func (s Samples) LongestDuration() time.Duration {
	longestDuration := s[0].Duration()

	for _, sample := range s[1:] {
		if sampleDuration := sample.Duration(); longestDuration < sampleDuration {
			longestDuration = sampleDuration
		}
	}

	return longestDuration
}

type TestSuite struct {
	Name  string
	Cases []*TestCase
}

func (s *TestSuite) GetTestCase(testName string) *TestCase {
	for _, testCase := range s.Cases {
		if testCase.Name == testName {
			return testCase
		}
	}

	return nil
}

func (s *TestSuite) NewTestCase(testName string, delegate TestDelegate) {
	s.Cases = append(s.Cases, &TestCase{
		Name:     testName,
		Delegate: delegate,
	})
}

type TestOutput struct {
	content strings.Builder
}

func (s *TestOutput) Print(value string) {
	s.content.WriteString(value)
}

func (s *TestOutput) Printf(format string, arguments ...any) {
	s.Print(fmt.Sprintf(format, arguments...))
}

func (s *TestOutput) String() string {
	return s.content.String()
}

func (s *TestSuite) Execute(db graph.Database) error {
	fmt.Printf("Database test suite %s executing\n", s.Name)

	var (
		suiteStart  = time.Now()
		tableWriter = table.NewWriter()
	)

	tableWriter.AppendHeader(table.Row{"Test Case", "Samples Taken", "Test Case Duration", "Shortest Sample Duration", "Longest Sample Duration", "Average Sample Duration"})

	for _, testCase := range s.Cases {
		caseStart := time.Now()

		if err := testCase.execute(db); err != nil {
			return fmt.Errorf("test case \"%s\" failed: %w", testCase.Name, err)
		}

		testCase.Duration = time.Since(caseStart)
		tableWriter.AppendRow(table.Row{testCase.Name, len(testCase.Samples), testCase.Duration, testCase.Samples.ShortestDuration(), testCase.Samples.LongestDuration(), testCase.Samples.AverageDuration()})
	}

	fmt.Print(tableWriter.Render())
	fmt.Printf("\nDatabase test suite %s completed in %d ms.\n", s.Name, time.Since(suiteStart).Milliseconds())

	return nil
}

type TestCase struct {
	Name     string
	Delegate TestDelegate
	Samples  Samples
	Duration time.Duration
}

// TODO: Plumb neo4j query analysis?
func (s TestCase) EnableAnalysis() {
	//pg.EnableQueryAnalysis()
}

// TODO: Plumb neo4j query analysis?
func (s TestCase) DisableAnalysis() {
	//pg.DisableQueryAnalysis()
}

func (s *TestCase) Sample(delegate func() error) error {
	// Track a new sample for this logic
	s.Samples = append(s.Samples, Sample{
		Start: time.Now(),
	})

	// Run the test logic
	err := delegate()

	// Set the end time of the last sample and then return the captured error
	s.Samples[len(s.Samples)-1].End = time.Now()
	return err
}

func (s *TestCase) execute(db graph.Database) error {
	switch typedDelegate := s.Delegate(s).(type) {
	case func(tx graph.Transaction) error:
		return db.WriteTransaction(context.Background(), typedDelegate)

	case func(batch graph.Batch) error:
		return db.BatchOperation(context.Background(), typedDelegate)

	default:
		panic(fmt.Sprintf("Bad test case delegate type: %T", s.Delegate))
	}
}
