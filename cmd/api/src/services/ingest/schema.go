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

package ingest

import (
	"bytes"
	"embed"
	"fmt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

//go:embed jsonschema
var schemaFiles embed.FS

// IngestSchema holds compiled JSON schemas used to validate
// generic-ingested graph data. It includes separate schemas for nodes
// and edges, which are reused across multiple ingestion requests
// to avoid recompiling on every request.
type IngestSchema struct {
	NodeSchema *jsonschema.Schema
	EdgeSchema *jsonschema.Schema
}

func LoadIngestSchema() (IngestSchema, error) {
	var schema IngestSchema
	if nodeSchema, err := loadSchema("node.json"); err != nil {
		return schema, err
	} else if edgeSchema, err := loadSchema("edge.json"); err != nil {
		return schema, err
	} else {
		schema.NodeSchema = nodeSchema
		schema.EdgeSchema = edgeSchema
		return schema, nil
	}
}

func loadSchema(filename string) (*jsonschema.Schema, error) {
	const schemaDir = "jsonschema"

	// Read the raw JSON schema file from embed.FS
	path := fmt.Sprintf("%s/%s", schemaDir, filename)
	data, err := schemaFiles.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema %q: %w", path, err)
	}

	// Parse the JSON into a generic in-memory representation
	document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema %q: %w", path, err)
	}

	// Create a new compiler and register the document
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource(filename, document); err != nil {
		return nil, fmt.Errorf("failed to add resource for schema %q: %w", filename, err)
	}

	// Compile the schema for validation use
	schema, err := compiler.Compile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema %q: %w", filename, err)
	}

	return schema, nil
}
