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
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractJsonSchemaErrors_SortsByLocation(t *testing.T) {
	var (
		node = map[string]any{
			"id":    float64(1),
			"kinds": []any{"User"},
			"properties": map[string]any{
				"z": map[string]any{},
				"a": map[string]any{},
			},
		}
		expectedDetails = []ValidationErrorDetail{
			{Location: "/id", Error: "got number, want string"},
			{Location: "/properties/a", Error: "invalid type"},
			{Location: "/properties/z", Error: "invalid type"},
		}
		schema, schemaLoadErr = LoadSchema()
		schemaErr             *jsonschema.ValidationError
	)

	require.NoError(t, schemaLoadErr)
	require.ErrorAs(t, schema.NodeSchema.Validate(node), &schemaErr)

	errorDetails, err := extractJsonSchemaErrors(schemaErr)
	require.NoError(t, err)
	assert.Equal(t, expectedDetails, errorDetails)
	assert.Equal(t, "validation error: /id: got number, want string; /properties/a: invalid type; /properties/z: invalid type", ValidationError{Errors: errorDetails}.Error())
}
