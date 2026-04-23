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
package preferences

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

//go:embed jsonschema
var schemaFile embed.FS

// ValidatePreferences checks that the given model.Preferences conform to the corresponding JSON Schema.
// It rejects unknown preference keys and incorrectly typed values.
func ValidatePreferences(preferences model.Preferences) error {
	loadedSchema, err := loadSchema("preferences_schema.json")
	if err != nil {
		return err
	}

	// transform model.Preferences into map[string]any for jsonschema.Validate
	transformed, err := transformPreferences(preferences)
	if err != nil {
		return err
	}

	// check that preferences conform to the JSON Schema
	err = loadedSchema.Validate(transformed)
	if err != nil {
		return fmt.Errorf("preferences validation failed: %w", err)
	}
	slog.Debug("JSON Schema preferences validation successful")

	return nil
}

// loadSchema reads and compiles a JSON Schema file from the embedded filesystem (embed.FS) for use in model.Preferences validation.
func loadSchema(filename string) (*jsonschema.Schema, error) {
	var (
		schemaDir = "jsonschema"
		compiler  = jsonschema.NewCompiler()
		path      = fmt.Sprintf("%s/%s", schemaDir, filename)
	)

	if data, err := schemaFile.ReadFile(path); err != nil {
		return nil, fmt.Errorf("failed to read schema %q: %w", path, err)
	} else if document, err := jsonschema.UnmarshalJSON(bytes.NewReader(data)); err != nil { // Parse the JSON into a generic in-memory representation
		return nil, fmt.Errorf("failed to unmarshal schema %q: %w", path, err)
	} else if err := compiler.AddResource(filename, document); err != nil {
		return nil, fmt.Errorf("failed to add resource for schema %q: %w", filename, err)
	} else if preferencesSchema, err := compiler.Compile(filename); err != nil {
		return nil, fmt.Errorf("failed to compile schema %q: %w", filename, err)
	} else {
		slog.Info("Preferences schema was successfully loaded") // TODO: remove if loading schema here and not at startup
		return preferencesSchema, nil
	}
}

// transformPreferences transforms a model.Preferences map into map[string]any
func transformPreferences(preferences model.Preferences) (map[string]any, error) {
	var (
		data   []byte
		err    error
		result map[string]any
	)

	// marshaling/unmarshaling (vs. manual conversion) for maintainability, in case new fields are later added to model.PreferenceItem
	data, err = json.Marshal(preferences)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
