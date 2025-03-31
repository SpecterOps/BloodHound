// Copyright 2024 Specter Ops, Inc.
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
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/specterops/bloodhound/dawgs/drivers/pg"

	"cuelang.org/go/pkg/regexp"
	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/translate"
	"github.com/specterops/bloodhound/cypher/models/walk"
	"github.com/stretchr/testify/require"
)

const (
	prefixCase          = "case:"
	prefixExclusiveTest = "exclusive:"
	prefixCypherParams  = "cypher_params:"
	prefixPgSQLParams   = "pgsql_params:"
)

//go:embed translation_cases/*
var testCaseFiles embed.FS

// Case is a translation test case
type TranslationTestCase struct {
	Name         string
	Cypher       string
	PgSQL        string
	CypherParams map[string]any
	PgSQLParams  map[string]any
}

func (s *TranslationTestCase) Reset() {
	s.Name = ""
	s.Cypher = ""
	s.PgSQL = ""
	s.CypherParams = nil
	s.PgSQLParams = nil
}

func (s *TranslationTestCase) Copy() *TranslationTestCase {
	return &TranslationTestCase{
		Name:         s.Name,
		Cypher:       s.Cypher,
		PgSQL:        s.PgSQL,
		CypherParams: s.CypherParams,
		PgSQLParams:  s.PgSQLParams,
	}
}

func writeStrings(output io.Writer, strs ...string) error {
	for _, str := range strs {
		if _, err := output.Write([]byte(str)); err != nil {
			return err
		}
	}

	return nil
}

var licenseHeader = `-- Copyright %d Specter Ops, Inc.
--
-- Licensed under the Apache License, Version 2.0
-- you may not use this file except in compliance with the License.
-- You may obtain a copy of the License at
--
--     http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing, software
-- distributed under the License is distributed on an "AS IS" BASIS,
-- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
-- See the License for the specific language governing permissions and
-- limitations under the License.
--
-- SPDX-License-Identifier: Apache-2.0

`

func (s *TranslationTestCase) WriteTo(output io.Writer, kindMapper pgsql.KindMapper) error {
	if regularQuery, err := frontend.ParseCypher(frontend.NewContext(), s.Cypher); err != nil {
		return err
	} else if err := writeStrings(output,
		"-- case: ",
		s.Cypher,
		"\n",
	); err != nil {
		return err
	} else {
		if len(s.CypherParams) > 0 {
			if err := walk.Cypher(regularQuery, walk.NewSimpleVisitor[cypher.SyntaxNode](func(node cypher.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
				switch typedNode := node.(type) {
				case *cypher.Parameter:
					if value, hasValue := s.CypherParams[typedNode.Symbol]; hasValue {
						typedNode.Value = value
					}
				}
			})); err != nil {
				return err
			}

			if encodedJSON, err := json.Marshal(s.CypherParams); err != nil {
				return err
			} else if err := writeStrings(output,
				"-- cypher_params: ",
				string(encodedJSON),
				"\n"); err != nil {
				return err
			}
		}

		if translation, err := translate.Translate(context.Background(), regularQuery, kindMapper, nil); err != nil {
			return err
		} else if formattedQuery, err := translate.Translated(translation); err != nil {
			return err
		} else {
			if len(translation.Parameters) > 0 {
				if encodedJSON, err := json.Marshal(translation.Parameters); err != nil {
					return err
				} else if err := writeStrings(output,
					"-- ", prefixPgSQLParams,
					string(encodedJSON),
					"\n",
				); err != nil {
					return err
				}
			}

			if err := writeStrings(output, formattedQuery, "\n\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *TranslationTestCase) Assert(t *testing.T, expectedSQL string, kindMapper pgsql.KindMapper) {
	if regularQuery, err := frontend.ParseCypher(frontend.NewContext(), s.Cypher); err != nil {
		t.Fatalf("Failed to compile cypher query: %s - %v", s.Cypher, err)
	} else {
		if s.CypherParams != nil {
			if err := walk.Cypher(regularQuery, walk.NewSimpleVisitor[cypher.SyntaxNode](func(node cypher.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
				switch typedNode := node.(type) {
				case *cypher.Parameter:
					if value, hasValue := s.CypherParams[typedNode.Symbol]; hasValue {
						typedNode.Value = value
					}
				}
			})); err != nil {
				t.Fatalf("Error attempting to set parameter values in cypher query: %v", err)
			}
		}

		if translation, err := translate.Translate(context.Background(), regularQuery, kindMapper, nil); err != nil {
			t.Fatalf("Failed to translate cypher query: %s - %v", s.Cypher, err)
		} else if formattedQuery, err := translate.Translated(translation); err != nil {
			t.Fatalf("Failed to format SQL translatedQuery: %v", err)
		} else {
			require.Equalf(t, expectedSQL, formattedQuery, "Test case for cypher query: '%s' failed to match.", s.Cypher)

			if s.PgSQLParams != nil {
				require.Equal(t, s.PgSQLParams, translation.Parameters)
			}
		}
	}
}

func (s *TranslationTestCase) AssertLive(ctx context.Context, t *testing.T, driver *pg.Driver) {
	if regularQuery, err := frontend.ParseCypher(frontend.NewContext(), s.Cypher); err != nil {
		t.Fatalf("Failed to compile cypher query: %s - %v", s.Cypher, err)
	} else {
		if s.CypherParams != nil {
			if err := walk.Cypher(regularQuery, walk.NewSimpleVisitor[cypher.SyntaxNode](func(node cypher.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
				switch typedNode := node.(type) {
				case *cypher.Parameter:
					if value, hasValue := s.CypherParams[typedNode.Symbol]; hasValue {
						typedNode.Value = value
					}
				}
			})); err != nil {
				t.Fatalf("Error attempting to set parameter values in cypher query: %v", err)
			}
		}

		if translation, err := translate.Translate(context.Background(), regularQuery, driver.KindMapper(), s.CypherParams); err != nil {
			t.Fatalf("Failed to translate cypher query: %s - %v", s.Cypher, err)
		} else if formattedQuery, err := translate.Translated(translation); err != nil {
			t.Fatalf("Failed to format SQL translatedQuery: %v", err)
		} else {
			require.Nil(t, driver.Run(ctx, "explain "+formattedQuery, translation.Parameters))
		}
	}
}

type TranslationTestCaseFile struct {
	path    string
	content []byte
}

func (s *TranslationTestCaseFile) Load() ([]*TranslationTestCase, bool, error) {
	var (
		testCases         []*TranslationTestCase
		isExclusive       = false
		hasExclusiveTests = false
		nextTestCase      = &TranslationTestCase{}
		queryBuilder      = strings.Builder{}
	)

	for lineNumber, line := range strings.Split(string(s.content), "\n") {
		// Crush unnecessary whitespace
		formattedLine, err := regexp.ReplaceAll("\\s+", strings.TrimSpace(line), " ")

		if err != nil {
			return nil, false, fmt.Errorf("error while attempting to collapse whitespace in query: %w", err)
		}

		if len(formattedLine) == 0 {
			continue
		}

		if isLineComment := strings.HasPrefix(formattedLine, "--"); isLineComment {
			// Strip the comment header
			formattedLine = strings.Trim(formattedLine, "- ")

			lowerFormattedLine := strings.ToLower(formattedLine)

			if caseIndex := strings.Index(lowerFormattedLine, prefixCase); caseIndex != -1 {
				// This is a new test case - capture the comment as the cypher statement to test
				nextTestCase.Cypher = strings.TrimSpace(formattedLine[caseIndex+len(prefixCase):])
			} else if strings.Contains(lowerFormattedLine, prefixExclusiveTest) {
				if !hasExclusiveTests {
					// Clear the existing test cases
					testCases = testCases[:0]
					hasExclusiveTests = true
				}

				// The current test case as being marked as run-only
				isExclusive = true
			} else if cypherParamsIdx := strings.Index(lowerFormattedLine, prefixCypherParams); cypherParamsIdx != -1 {
				paramsContent := []byte(strings.TrimSpace(formattedLine[cypherParamsIdx+len(prefixCypherParams):]))
				nextTestCase.CypherParams = map[string]any{}

				if err := json.Unmarshal(paramsContent, &nextTestCase.CypherParams); err != nil {
					return nil, false, fmt.Errorf("failed to unmarshal cypher params on line number %d: %w", lineNumber, err)
				}
			} else if pgsqlParamsIdx := strings.Index(lowerFormattedLine, prefixPgSQLParams); pgsqlParamsIdx != -1 {
				paramsContent := []byte(strings.TrimSpace(formattedLine[pgsqlParamsIdx+len(prefixPgSQLParams):]))
				nextTestCase.PgSQLParams = map[string]any{}

				if err := json.Unmarshal(paramsContent, &nextTestCase.PgSQLParams); err != nil {
					return nil, false, fmt.Errorf("failed to unmarshal pgsql params on line number %d: %w", lineNumber, err)
				}
			}
		} else if len(nextTestCase.Cypher) > 0 {
			// Strip any comment fragments for this line. Best effort; probably better done with a regex.
			if inlineCommentIdx := strings.Index(formattedLine, "--"); inlineCommentIdx >= 0 {
				formattedLine = strings.TrimSpace(formattedLine[:inlineCommentIdx])
			}

			// Check to make sure there's translatedQuery content.
			if len(formattedLine) == 0 {
				continue
			}

			// If there's content in the translatedQuery builder, prepend a space to conjoin the lines
			if queryBuilder.Len() > 0 {
				queryBuilder.WriteRune(' ')
			}

			queryBuilder.WriteString(formattedLine)

			// SQL queries must end with a ';' character
			if strings.HasSuffix(formattedLine, ";") {
				nextTestCase.PgSQL = queryBuilder.String()

				// Format the expected SQL translation and create a sub-test
				if isExclusive || !hasExclusiveTests {
					nextTestCase.Name = filepath.Base(s.path) + " " + nextTestCase.Cypher
					testCases = append(testCases, nextTestCase.Copy())
				}

				// Reset the query builder and test case
				queryBuilder.Reset()
				nextTestCase.Reset()

				isExclusive = false
			}
		}
	}

	return testCases, hasExclusiveTests, nil
}

func ReadTranslationTestCaseFile(path string, fin fs.File) (TranslationTestCaseFile, error) {
	content, err := io.ReadAll(fin)

	return TranslationTestCaseFile{
		path:    path,
		content: content,
	}, err
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

func UpdateTranslationTestCases(mapper pgsql.KindMapper) error {
	if updatedCasesPath, err := updatedCasesDir(); err != nil {
		return err
	} else if err := fs.WalkDir(testCaseFiles, "translation_cases", func(path string, dir fs.DirEntry, err error) error {
		if !dir.IsDir() {
			var (
				caseFileName        = filepath.Base(path)
				updatedCaseFilePath = filepath.Join(updatedCasesPath, caseFileName)
			)

			if strings.HasSuffix(path, ".sql") {
				if fin, err := testCaseFiles.Open(path); err != nil {
					return err
				} else {
					defer fin.Close()

					if caseFile, err := ReadTranslationTestCaseFile(path, fin); err != nil {
						return err
					} else if nextCases, _, err := caseFile.Load(); err != nil {
						return err
					} else if output, err := os.OpenFile(updatedCaseFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
						return err
					} else {
						formattedLicenseHeader := fmt.Sprintf(licenseHeader, time.Now().Year())

						if _, err := io.WriteString(output, formattedLicenseHeader); err != nil {
							return err
						}

						for _, nextCase := range nextCases {
							nextCase.WriteTo(output, mapper)
						}

						output.Close()
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

func ReadTranslationTestCases() ([]*TranslationTestCase, error) {
	var caseFiles []TranslationTestCaseFile

	if err := fs.WalkDir(testCaseFiles, "translation_cases", func(path string, dir fs.DirEntry, err error) error {
		if !dir.IsDir() {
			if strings.HasSuffix(path, ".sql") {
				if fin, err := testCaseFiles.Open(path); err != nil {
					return err
				} else {
					defer fin.Close()

					if caseFile, err := ReadTranslationTestCaseFile(path, fin); err != nil {
						return err
					} else {
						caseFiles = append(caseFiles, caseFile)
					}
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	var (
		cases             []*TranslationTestCase
		hasExclusiveTests bool
	)

	for _, caseFile := range caseFiles {
		loadedTestCases, caseFileHasExclusiveTests, err := caseFile.Load()

		if err != nil {
			return nil, err
		}

		if !hasExclusiveTests {
			if caseFileHasExclusiveTests {
				hasExclusiveTests = true
				cases = cases[:0]
			}

			cases = append(cases, loadedTestCases...)
		} else if caseFileHasExclusiveTests {
			cases = append(cases, loadedTestCases...)
		}
	}

	return cases, nil
}
