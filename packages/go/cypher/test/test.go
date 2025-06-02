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
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/cypher/analyzer"
	"github.com/specterops/bloodhound/cypher/models/cypher"
	format2 "github.com/specterops/bloodhound/cypher/models/cypher/format"

	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/stretchr/testify/require"
)

//go:embed cases
var testCaseFiles embed.FS

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

type Emitter interface {
	Write(query *cypher.RegularQuery, writer io.Writer) error
	WriteExpression(output io.Writer, expression cypher.Expression) error
}

func CypherToCypher(ctx *frontend.Context, input string) (string, error) {
	if query, err := frontend.ParseCypher(ctx, input); err != nil {
		return "", err
	} else {
		var (
			output  = &bytes.Buffer{}
			emitter = format2.Emitter{
				StripLiterals: false,
			}
		)

		if err := emitter.Write(query, output); err != nil {
			return "", err
		}

		return output.String(), nil
	}
}

type StringMatchTest struct {
	Query           string `json:"query"`
	Matcher         string `json:"matcher,omitempty"`
	ExpectedFitness *int64 `json:"fitness"`
}

func (s StringMatchTest) Run(t *testing.T, testCase Case) {
	var (
		ctx         = frontend.NewContext()
		result, err = CypherToCypher(ctx, s.Query)
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
	Targeted bool            `json:"targeted,omitempty"`
	Ignore   bool            `json:"ignore,omitempty"`
	Details  json.RawMessage `json:"details"`
}

func (s Case) MarshalDetails(target any) error {
	return json.Unmarshal(s.Details, target)
}

func UnmarshallTestCaseDetails[T any](testCase Case) (T, error) {
	var value T

	if err := testCase.MarshalDetails(&value); err != nil {
		return value, fmt.Errorf("failed while unmarshaling details for test case %s: %v", testCase.Name, err)
	}

	return value, nil
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
		file, err = testCaseFiles.Open(filename)
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
			if typedTestCase, err := UnmarshallTestCaseDetails[T](testCase); err != nil {
				t.Fatalf("Error unmarshalling test case details: %v", err)
			} else {
				typedTestCase.Run(t, testCase)
			}
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

func updatedCasesDir() (string, error) {
	if workingDir, err := os.Getwd(); err != nil {
		return "", err
	} else {
		path := filepath.Join(workingDir, "updated_cases")

		if err := os.MkdirAll(path, 0755); err != nil {
			return "", err
		}

		return path, nil
	}
}

func UpdatePositiveTestCasesFitness() error {
	if updatedCasesPath, err := updatedCasesDir(); err != nil {
		return err
	} else if err := fs.WalkDir(testCaseFiles, "cases", func(path string, dir fs.DirEntry, err error) error {
		if dir.IsDir() {
			return nil
		}

		var (
			caseFileName        = filepath.Base(path)
			updatedCaseFilePath = filepath.Join(updatedCasesPath, caseFileName)
		)

		if strings.HasSuffix(path, ".json") {
			if fin, err := testCaseFiles.Open(path); err != nil {
				return err
			} else {
				defer fin.Close()

				var (
					cases        Cases
					updatedCases Cases
				)

				if err := json.NewDecoder(fin).Decode(&cases); err != nil {
					return fmt.Errorf("error decoding fixture: %v", err)
				}

				for _, nextCase := range cases.TestCases {
					if nextCase.Type == TypeStringMatch {
						parseContext := frontend.NewContext()

						if details, err := UnmarshallTestCaseDetails[StringMatchTest](nextCase); err != nil {
							return fmt.Errorf("error unmarshalling test case details: %v", err)
						} else if queryModel, err := frontend.ParseCypher(parseContext, details.Query); err != nil {
							return fmt.Errorf("parser errors: %s", err.Error())
						} else if complexity, err := analyzer.QueryComplexity(queryModel); err != nil {
							return fmt.Errorf("analyzer errors: %s", err.Error())
						} else {
							details.ExpectedFitness = &complexity.RelativeFitness

							if updatedDetails, err := json.Marshal(details); err != nil {
								return fmt.Errorf("error marshalling test case details: %v", err)
							} else {
								nextCase.Details = updatedDetails
							}
						}

						updatedCases.TestCases = append(updatedCases.TestCases, nextCase)
					} else {
						updatedCases.TestCases = append(updatedCases.TestCases, nextCase)
					}
				}

				if output, err := os.OpenFile(updatedCaseFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
					return err
				} else {
					defer output.Close()

					if err := json.NewEncoder(output).Encode(updatedCases); err != nil {
						return err
					}
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
