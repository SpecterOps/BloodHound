package translate_test

import (
	"cuelang.org/go/pkg/regexp"
	"embed"
	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/cypher/model/pgsql/format"
	"github.com/specterops/bloodhound/cypher/model/pgsql/translate"
	"github.com/stretchr/testify/require"
	"io"
	"io/fs"
	"strings"
	"testing"
)

//go:embed translation_cases/*
var testCaseFiles embed.FS

type Case struct {
	Cypher string
}

func (s *Case) Reset() {
	s.Cypher = ""
}

func (s *Case) Assert(t *testing.T, expectedSQL string) {
	if regularQuery, err := frontend.ParseCypher(frontend.NewContext(), s.Cypher); err != nil {
		t.Fatalf("Failed to compile cypher translatedQuery: %s - %v", s.Cypher, err)
	} else if sqlStatement, err := translate.Translate(regularQuery); err != nil {
		t.Fatalf("Failed to translate cypher translatedQuery: %s - %v", s.Cypher, err)
	} else if formattedQuery, err := format.Statement(sqlStatement); err != nil {
		t.Fatalf("Failed to format SQL translatedQuery: %v", err)
	} else {
		require.Equalf(t, expectedSQL, formattedQuery.Value, "Test case for cypher translatedQuery: '%s' failed to match.", s.Cypher)
	}
}

type CaseFile struct {
	path    string
	content []byte
}

func (s *CaseFile) Run(t *testing.T) {
	const (
		casePrefix = "case:"
	)

	var (
		nextTestCase = &Case{}
		queryBuilder = strings.Builder{}
	)

	for _, line := range strings.Split(string(s.content), "\n") {
		// Crush unnecessary whitespace
		formattedLine, err := regexp.ReplaceAll("\\s+", strings.TrimSpace(line), " ")
		require.Nilf(t, err, "error while attempting to collapse whitespace in query")

		if len(formattedLine) == 0 {
			continue
		}

		if isLineComment := strings.HasPrefix(formattedLine, "--"); isLineComment {
			// Strip the comment header
			formattedLine = strings.Trim(formattedLine, "- ")

			if caseIndex := strings.Index(strings.ToLower(formattedLine), casePrefix); caseIndex != -1 {
				// This is a new test case - capture the comment as the cypher statement to test
				nextTestCase.Cypher = strings.TrimSpace(formattedLine[caseIndex+len(casePrefix):])
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
				// Format the expected SQL translation and create a sub-test
				t.Run(nextTestCase.Cypher, func(t *testing.T) {
					nextTestCase.Assert(t, queryBuilder.String())
				})

				// Reset the translatedQuery builder and test case
				queryBuilder.Reset()
				nextTestCase.Reset()
			}
		}
	}
}

func ReadCaseFile(path string, fin fs.File) (CaseFile, error) {
	content, err := io.ReadAll(fin)

	return CaseFile{
		path:    path,
		content: content,
	}, err
}

func TestTranslate(t *testing.T) {
	var caseFiles []CaseFile

	require.Nil(t, fs.WalkDir(testCaseFiles, "translation_cases", func(path string, dir fs.DirEntry, err error) error {
		if !dir.IsDir() {
			if strings.HasSuffix(path, ".sql") {
				if fin, err := testCaseFiles.Open(path); err != nil {
					return err
				} else {
					defer fin.Close()

					if caseFile, err := ReadCaseFile(path, fin); err != nil {
						return err
					} else {
						caseFiles = append(caseFiles, caseFile)
					}
				}
			}
		}

		return nil
	}))

	for _, caseFile := range caseFiles {
		caseFile.Run(t)
	}
}
