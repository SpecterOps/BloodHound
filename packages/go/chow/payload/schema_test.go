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
package payload

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadSchema(t *testing.T) {
	schema, err := LoadSchema()
	require.NoError(t, err)

	assert.NotNil(t, schema.NodeSchema)
	assert.NotNil(t, schema.EdgeSchema)
	assert.NotNil(t, schema.MetaSchema)
}

func TestLoadSchemaFromFS(t *testing.T) {
	assertions := []struct {
		name        string
		schemaFS    fstest.MapFS
		filename    string
		errContains string
	}{
		{
			name:        "missing file",
			schemaFS:    fstest.MapFS{},
			filename:    "missing.json",
			errContains: `failed to read schema "jsonschema/missing.json"`,
		},
		{
			name: "invalid JSON",
			schemaFS: fstest.MapFS{
				"jsonschema/schema.json": {Data: []byte(`{`)},
			},
			filename:    "schema.json",
			errContains: `failed to unmarshal schema "jsonschema/schema.json"`,
		},
		{
			name: "invalid resource URL",
			schemaFS: fstest.MapFS{
				"jsonschema/%zz.json": {Data: []byte(`{"type":"object"}`)},
			},
			filename:    "%zz.json",
			errContains: `failed to add resource for schema "%zz.json"`,
		},
		{
			name: "invalid schema definition",
			schemaFS: fstest.MapFS{
				"jsonschema/schema.json": {Data: []byte(`{"type":"definitely_not_a_json_schema_type"}`)},
			},
			filename:    "schema.json",
			errContains: `failed to compile schema "schema.json"`,
		},
	}

	for _, assertion := range assertions {
		t.Run(assertion.name, func(t *testing.T) {
			compiledSchema, err := loadSchemaFromFS(assertion.schemaFS, assertion.filename)

			assert.Nil(t, compiledSchema)
			assert.ErrorContains(t, err, assertion.errContains)
		})
	}

	t.Run("valid schema", func(t *testing.T) {
		compiledSchema, err := loadSchemaFromFS(fstest.MapFS{
			"jsonschema/schema.json": {Data: []byte(`{"type":"object"}`)},
		}, "schema.json")

		require.NoError(t, err)
		assert.NotNil(t, compiledSchema)
	})
}

func TestEmbeddedTopLevelJSONSchemaCompilesWithReferences(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	for _, filename := range []string{"schema.json", "metadata.json", "node.json", "edge.json"} {
		data, err := schemaFiles.ReadFile("jsonschema/" + filename)
		require.NoError(t, err)

		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
		require.NoError(t, err)
		require.NoError(t, compiler.AddResource(filename, document))
	}

	schema, err := compiler.Compile("schema.json")
	require.NoError(t, err)
	require.NoError(t, schema.Validate(map[string]any{
		"metadata": map[string]any{"source_kind": "hellobase"},
		"graph": map[string]any{
			"nodes": []any{
				map[string]any{"id": "node-1", "kinds": []any{"User"}},
			},
			"edges": []any{
				map[string]any{
					"start": map[string]any{"value": "node-1"},
					"end":   map[string]any{"value": "node-2"},
					"kind":  "RELATED",
				},
			},
		},
	}))
	require.Error(t, schema.Validate(map[string]any{
		"graph": map[string]any{
			"nodes": []any{
				map[string]any{"kinds": []any{"User"}},
			},
		},
	}))
	require.NoError(t, schema.Validate(map[string]any{
		"graph": map[string]any{"nodes": []any{}},
		"extra": true,
	}))
	require.Error(t, schema.Validate(map[string]any{
		"graph": map[string]any{
			"nodes":  []any{},
			"strays": []any{},
		},
	}))
}
