// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"archive/zip"
	"bytes"
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const validSchemaBaseJSON = `{
	"schema": {
		"name": "TestExtension",
		"display_name": "Test Extension",
		"version": "v1.0.0",
		"namespace": "test"
	},
	"node_kinds": [],
	"relationship_kinds": [],
	"environments": []`

// validPZRulesJSON is a minimal privilege zone rules component.
const validPZRulesJSON = `{
	"rules": [
		{
			"name": "Tier Zero Admins",
			"description": "seeds for tier zero",
			"seeds": [
				{"type": 1, "value": "match (n:OG_Kind) return n"}
			]
		}
	]
}`

// validSavedQueryJSON is a minimal saved query definition.
const validSavedQueryJSON = `{
	"query_key": "all-domain-admins",
	"name": "All Domain Admins",
	"query": "MATCH (n) RETURN n LIMIT 1",
	"description": "example",
	"category": "administration"
}`

// validSavedQueriesJSON is a minimal saved queries component.
const validSavedQueriesJSON = `{
	"queries": [` + validSavedQueryJSON + `]
}`

// validSchemaJSON is a minimal extension definition schema used to exercise the ZIP ingest path.
const validSchemaJSON = validSchemaBaseJSON + `,
	"relationship_findings": []
}`

const validSchemaWithFindingsJSON = validSchemaBaseJSON + `,
	"relationship_findings": [
		{"name": "T0Example"}
	]
}`

const validSchemaWithEmbeddedPZRulesJSON = validSchemaBaseJSON + `,
	"relationship_findings": [],
	"pz_rules": ` + validPZRulesJSON + `
}`

const validSchemaWithEmbeddedSavedQueriesJSON = validSchemaBaseJSON + `,
	"relationship_findings": [],
	"queries": [` + validSavedQueryJSON + `]
}`

const validSchemaWithEmbeddedOptionalComponentsJSON = validSchemaBaseJSON + `,
	"relationship_findings": [],
	"pz_rules": ` + validPZRulesJSON + `,
	"queries": [` + validSavedQueryJSON + `]
}`

// newExtensionZip builds an in-memory ZIP archive from the given file name -> content map.
func newExtensionZip(t *testing.T, entries map[string]string) []byte {
	t.Helper()

	var buf bytes.Buffer

	zipWriter := zip.NewWriter(&buf)
	for name, content := range entries {
		entryWriter, err := zipWriter.Create(name)
		require.NoError(t, err)
		_, err = entryWriter.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, zipWriter.Close())

	return buf.Bytes()
}

// newExtensionZipWithDuplicate builds an in-memory ZIP archive containing two entries with the same name and content.
func newExtensionZipWithDuplicate(t *testing.T, name, content string) []byte {
	t.Helper()

	var buf bytes.Buffer

	zipWriter := zip.NewWriter(&buf)
	for i := 0; i < 2; i++ {
		entryWriter, err := zipWriter.Create(name)
		require.NoError(t, err)
		_, err = entryWriter.Write([]byte(content))
		require.NoError(t, err)
	}
	require.NoError(t, zipWriter.Close())

	return buf.Bytes()
}

func newOversizedExtensionZip(t *testing.T, name, baseContent string) []byte {
	t.Helper()

	var (
		buf              bytes.Buffer
		whitespaceChunk  = bytes.Repeat([]byte(" "), 4*1024)
		uncompressedSize int
	)

	zipWriter := zip.NewWriter(&buf)
	entryWriter, err := zipWriter.Create(name)
	require.NoError(t, err)

	bytesWritten, err := entryWriter.Write([]byte(baseContent))
	require.NoError(t, err)
	uncompressedSize += bytesWritten

	for uncompressedSize <= bundleComponentDecompressedReadLimitBytes {
		bytesWritten, err = entryWriter.Write(whitespaceChunk)
		require.NoError(t, err)
		uncompressedSize += bytesWritten
	}

	require.NoError(t, zipWriter.Close())

	archive := buf.Bytes()
	require.Less(t, len(archive), api.DefaultAPIPayloadReadLimitBytes)

	return archive
}

func TestExtractBundleFromZip(t *testing.T) {
	type expectedResults struct {
		errorText    string
		schemaName   string
		hasPZRules   bool
		savedQueries *model.SavedQueriesPayload
	}

	var tests = []struct {
		name     string
		archive  []byte
		expected expectedResults
	}{
		{
			name:    "valid zip with only schema.json yields a schema-only bundle",
			archive: newExtensionZip(t, map[string]string{"schema.json": validSchemaJSON}),
			expected: expectedResults{
				schemaName: "TestExtension",
			},
		},
		{
			name:    "schema.json exceeding decompressed component limit is rejected",
			archive: newOversizedExtensionZip(t, bundleFileNameSchema, validSchemaJSON),
			expected: expectedResults{
				errorText: fmt.Sprintf("component %q exceeds decompressed size limit of %d bytes", bundleFileNameSchema, bundleComponentDecompressedReadLimitBytes),
			},
		},
		{
			name:    "zip schema with findings requires pz_rules.json",
			archive: newExtensionZip(t, map[string]string{"schema.json": validSchemaWithFindingsJSON}),
			expected: expectedResults{
				errorText:  "requires a \"pz_rules.json\" component",
				schemaName: "TestExtension",
			},
		},
		{
			name: "zip schema with findings requires at least one pz rule",
			archive: newExtensionZip(t, map[string]string{
				"schema.json":   validSchemaWithFindingsJSON,
				"pz_rules.json": `{"rules": []}`,
			}),
			expected: expectedResults{
				errorText:  "\"pz_rules.json\" must contain at least one rule",
				schemaName: "TestExtension",
				hasPZRules: true,
			},
		},
		{
			name: "zip schema with findings and pz_rules.json is valid",
			archive: newExtensionZip(t, map[string]string{
				"schema.json":   validSchemaWithFindingsJSON,
				"pz_rules.json": validPZRulesJSON,
			}),
			expected: expectedResults{
				schemaName: "TestExtension",
				hasPZRules: true,
			},
		},
		{
			name: "valid zip with all three components populates the full bundle",
			archive: newExtensionZip(t, map[string]string{
				"schema.json":        validSchemaJSON,
				"pz_rules.json":      validPZRulesJSON,
				"saved_queries.json": validSavedQueriesJSON,
			}),
			expected: expectedResults{
				schemaName: "TestExtension",
				hasPZRules: true,
				savedQueries: &model.SavedQueriesPayload{
					{
						QueryKey:    "all-domain-admins",
						Name:        "All Domain Admins",
						Query:       "MATCH (n) RETURN n LIMIT 1",
						Description: "example",
						Category:    "administration",
					},
				},
			},
		},
		{
			name:    "missing schema.json is an extractor error",
			archive: newExtensionZip(t, map[string]string{"pz_rules.json": validPZRulesJSON}),
			expected: expectedResults{
				errorText:  "required component \"schema.json\" not found in extension bundle",
				hasPZRules: true,
			},
		},
		{
			name:    "malformed archive returns an open error",
			archive: []byte("this is not a zip archive"),
			expected: expectedResults{
				errorText: "unable to open zip archive",
			},
		},
		{
			name:    "invalid json in schema.json returns a decode error",
			archive: newExtensionZip(t, map[string]string{"schema.json": "{ not valid json"}),
			expected: expectedResults{
				errorText: "unable to decode graph extension payload",
			},
		},
		{
			name: "unexpected file returns an error but recognized components still populate",
			archive: newExtensionZip(t, map[string]string{
				"schema.json": validSchemaJSON,
				"README.md":   "not a component",
			}),
			expected: expectedResults{
				errorText:  "unexpected file \"README.md\" in zip archive",
				schemaName: "TestExtension",
			},
		},
		{
			name:    "duplicate component (same file name twice) returns an error",
			archive: newExtensionZipWithDuplicate(t, "schema.json", validSchemaJSON),
			expected: expectedResults{
				errorText:  "duplicate component \"schema.json\" in zip archive",
				schemaName: "TestExtension",
			},
		},
		{
			name:    "duplicate null saved query components are rejected",
			archive: newExtensionZipWithDuplicate(t, "saved_queries.json", `{"queries": null}`),
			expected: expectedResults{
				errorText:    "duplicate component \"saved_queries.json\" in zip archive",
				savedQueries: new(model.SavedQueriesPayload),
			},
		},
		{
			name:    "duplicate empty saved query components are rejected",
			archive: newExtensionZipWithDuplicate(t, "saved_queries.json", `{"queries": []}`),
			expected: expectedResults{
				errorText:    "duplicate component \"saved_queries.json\" in zip archive",
				savedQueries: &model.SavedQueriesPayload{},
			},
		},
		{
			name:    "duplicate saved query components with missing queries are rejected",
			archive: newExtensionZipWithDuplicate(t, "saved_queries.json", `{}`),
			expected: expectedResults{
				errorText:    "duplicate component \"saved_queries.json\" in zip archive",
				savedQueries: new(model.SavedQueriesPayload),
			},
		},
		{
			name:    "schema.json embedding pz_rules is rejected in a bundle",
			archive: newExtensionZip(t, map[string]string{"schema.json": validSchemaWithEmbeddedPZRulesJSON}),
			expected: expectedResults{
				errorText: "schema.json must not embed optional components: pz_rules.json",
			},
		},
		{
			name:    "schema.json embedding saved_queries is rejected in a bundle",
			archive: newExtensionZip(t, map[string]string{"schema.json": validSchemaWithEmbeddedSavedQueriesJSON}),
			expected: expectedResults{
				errorText: "schema.json must not embed optional components: saved_queries.json",
			},
		},
		{
			name:    "schema.json embedding both optional components is rejected in a bundle",
			archive: newExtensionZip(t, map[string]string{"schema.json": validSchemaWithEmbeddedOptionalComponentsJSON}),
			expected: expectedResults{
				errorText: "schema.json must not embed optional components: pz_rules.json, saved_queries.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extension, err := extractBundleFromZip(bytes.NewReader(tt.archive))
			if tt.expected.errorText != "" {
				if assert.Error(t, err) {
					assert.Contains(t, err.Error(), tt.expected.errorText)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected.schemaName, extension.GraphSchemaExtension.Name)
			assert.Equal(t, tt.expected.hasPZRules, extension.PZRules != nil)
			assert.Equal(t, tt.expected.savedQueries, extension.SavedQueries)
		})
	}
}

func TestExtractExtensionDataFromJSON(t *testing.T) {
	t.Run("valid schema json yields a schema-only payload", func(t *testing.T) {
		extension, err := extractExtensionDataFromJSON(bytes.NewReader([]byte(validSchemaJSON)))
		require.NoError(t, err)
		assert.Equal(t, "TestExtension", extension.GraphSchemaExtension.Name)
		assert.Nil(t, extension.PZRules)
		assert.Nil(t, extension.SavedQueries)
	})

	t.Run("schema with saved queries preserves every query field", func(t *testing.T) {
		extension, err := extractExtensionDataFromJSON(bytes.NewReader([]byte(validSchemaWithEmbeddedSavedQueriesJSON)))
		require.NoError(t, err)
		require.NotNil(t, extension.SavedQueries)
		require.Len(t, *extension.SavedQueries, 1)
		assert.Equal(t, model.SavedQueryPayload{
			QueryKey:    "all-domain-admins",
			Name:        "All Domain Admins",
			Query:       "MATCH (n) RETURN n LIMIT 1",
			Description: "example",
			Category:    "administration",
		}, (*extension.SavedQueries)[0])
	})

	t.Run("schema with findings is valid without pz_rules", func(t *testing.T) {
		extension, err := extractExtensionDataFromJSON(bytes.NewReader([]byte(validSchemaWithFindingsJSON)))
		require.NoError(t, err)
		assert.Len(t, extension.GraphRelationshipFindings, 1)
		assert.Nil(t, extension.PZRules)
	})

	t.Run("invalid json returns a decode error", func(t *testing.T) {
		_, err := extractExtensionDataFromJSON(bytes.NewReader([]byte("{ not valid json")))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to decode graph extension payload")
	})
}
func TestValidateZipBundle(t *testing.T) {
	var (
		schemaWithFindings = model.GraphExtensionPayload{
			GraphRelationshipFindings: []model.RelationshipFindingsPayload{
				{Name: "T0Example"},
			},
		}
		tests = []struct {
			name        string
			payload     model.GraphExtensionPayload
			wantErrText string
		}{
			{
				name:    "community extension (no findings) is valid without pz_rules",
				payload: model.GraphExtensionPayload{},
			},
			{
				name:        "enterprise extension (findings) without pz_rules is rejected",
				payload:     schemaWithFindings,
				wantErrText: "requires a \"pz_rules.json\" component",
			},
			{
				name: "enterprise extension (findings) with pz_rules is valid",
				payload: func() model.GraphExtensionPayload {
					var payload = schemaWithFindings
					payload.PZRules = &model.PZRulesPayload{Rules: []model.PZRulePayload{{Name: "Test rule"}}}
					return payload
				}(),
			},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateZipBundle(tt.payload)
			if tt.wantErrText != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrText)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
