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

package test

import (
	"embed"
	"encoding/json"
	"github.com/specterops/bloodhound/cypher/gen"
	"regexp"
	"testing"

	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/stretchr/testify/require"
)

//go:embed cases
var fixtureFS embed.FS

type Type = string

const (
	TypeStringMatch  Type = "string_match"
	TypeNegativeCase Type = "negative_case"
)

type Runner interface {
	Run(t *testing.T, testCase Case)
}

type NegativeTest struct {
	Queries       []string `json:"queries"`
	ErrorMatchers []string `json:"error_matchers"`
}

func (s NegativeTest) Run(t *testing.T, testCase Case) {
	for _, query := range s.Queries {
		var (
			ctx               = frontend.DefaultCypherContext()
			_, combinedErr    = frontend.ParseCypher(ctx, query)
			foundMatcherError = false
		)

		// This set of queries may not return errors
		if len(s.ErrorMatchers) == 0 {
			require.Nilf(t, combinedErr, "Test case may not return an error but encountered:\n%v", combinedErr)
		} else {
			require.NotNil(t, combinedErr, "Test case did not return an error where expected")

			for _, errorMatcherStr := range s.ErrorMatchers {
				var (
					errorMatcher   = regexp.MustCompile(errorMatcherStr)
					combinedErrStr = combinedErr.Error()
				)

				if errorMatcherStr == combinedErrStr || errorMatcher.MatchString(combinedErrStr) {
					foundMatcherError = true
					break
				}

				for _, queryErr := range ctx.Errors {
					queryErrStr := queryErr.Error()

					if errorMatcherStr == queryErrStr || errorMatcher.MatchString(queryErrStr) {
						foundMatcherError = true
						break
					}
				}

				if foundMatcherError {
					break
				}
			}

			require.Truef(t, foundMatcherError, "Test case did not return an error matching defined expectations.Error: %v", combinedErr)
		}
	}
}

type StringMatchTest struct {
	Query      string   `json:"query"`
	Matcher    string   `json:"matcher"`
	Complexity *float64 `json:"complexity"`
}

func (s StringMatchTest) Run(t *testing.T, testCase Case) {
	var (
		ctx         = frontend.NewContext()
		result, err = gen.CypherToCypher(ctx, s.Query)
	)

	if err != nil {
		t.Fatalf("Unexpected error during parsing: %s", err.Error())
	}

	if s.Matcher != "" {
		// Attempt a regex match against the query output
		if matcher := regexp.MustCompile(s.Matcher); !matcher.MatchString(result) {
			t.Fatalf("Unable to find a match for query\nResult: %s\nExpected: %s", result, s.Query)
		}
	} else if s.Query != result {
		require.Equal(t, s.Query, result)
	}
}

type Case struct {
	Name     string          `json:"name"`
	Type     Type            `json:"type"`
	Targeted bool            `json:"targeted"`
	Ignore   bool            `json:"ignore"`
	Details  json.RawMessage `json:"details"`
}

func (s Case) MarshalDetails(target any) error {
	return json.Unmarshal(s.Details, target)
}

func UnmarshallTestCaseDetails[T any](t *testing.T, testCase Case) T {
	var value T

	if err := testCase.MarshalDetails(&value); err != nil {
		t.Fatalf("Failed while unmarshaling details for test case %s: %v", testCase.Name, err)
	}

	return value
}

type Cases struct {
	TestCases []Case `json:"test_cases"`
}

func (s Cases) RunnableCases() []Case {
	var (
		allTestCases  []Case
		targetedCases []Case
	)

	for _, testCase := range s.TestCases {
		if testCase.Targeted {
			targetedCases = append(targetedCases, testCase)
		}

		if !testCase.Ignore {
			allTestCases = append(allTestCases, testCase)
		}
	}

	if len(targetedCases) > 0 {
		return targetedCases
	}

	return allTestCases
}

func (s Cases) Run(t *testing.T) {
	for _, test := range s.RunnableCases() {
		t.Run(test.Name, testCase(test))
	}
}

func LoadFixture(t *testing.T, filename string) Cases {
	var (
		fixture   Cases
		file, err = fixtureFS.Open(filename)
	)

	if err != nil {
		t.Fatalf("Error loading fixture: %v", err)
	} else {
		defer file.Close()
	}

	if err := json.NewDecoder(file).Decode(&fixture); err != nil {
		t.Fatalf("Error decoding fixture: %v", err)
	}

	return fixture
}

func testRunner[T Runner](testCase Case) func(t *testing.T) {
	return func(t *testing.T) {
		// Run the test case if it isn't ignored
		if !testCase.Ignore {
			UnmarshallTestCaseDetails[T](t, testCase).Run(t, testCase)
		}
	}
}

func testCase(test Case) func(t *testing.T) {
	switch test.Type {
	case TypeStringMatch:
		return testRunner[StringMatchTest](test)

	case TypeNegativeCase:
		return testRunner[NegativeTest](test)

	default:
		return func(t *testing.T) {
			t.Fatalf("Unknown test type: %s", test.Type)
		}
	}
}
