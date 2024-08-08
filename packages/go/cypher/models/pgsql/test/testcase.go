package test

import (
	"cuelang.org/go/pkg/regexp"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/translate"
	"github.com/specterops/bloodhound/cypher/models/walk"
	"github.com/stretchr/testify/require"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
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

func (s *TranslationTestCase) Assert(t *testing.T, expectedSQL string, kindMapper pgsql.KindMapper) {
	if regularQuery, err := frontend.ParseCypher(frontend.NewContext(), s.Cypher); err != nil {
		t.Fatalf("Failed to compile cypher query: %s - %v", s.Cypher, err)
	} else {
		if s.CypherParams != nil {
			if err := walk.WalkCypher(regularQuery, walk.NewSimpleVisitor[cypher.SyntaxNode](func(node cypher.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
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

		if translation, err := translate.Translate(regularQuery, kindMapper); err != nil {
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

type TranslationTestCaseFile struct {
	path    string
	content []byte
}

func (s *TranslationTestCaseFile) Load() ([]*TranslationTestCase, bool, error) {
	const (
		prefixCase          = "case:"
		prefixExclusiveTest = "exclusive:"
		prefixCypherParams  = "cypher_params:"
		prefixPgSQLParams   = "pgsql_params:"
	)

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
