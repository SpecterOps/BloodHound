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

package ingest

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/specterops/bloodhound/src/model/ingest"
)

var ZipMagicBytes = []byte{0x50, 0x4b, 0x03, 0x04}

//go:embed json_schema
var schemaFiles embed.FS

// ValidateMetaTag ensures that the correct tags are present in a json file for data ingest.
// If readToEnd is set to true, the stream will read to the end of the file (needed for TeeReader)
func ValidateMetaTag(reader io.Reader, readToEnd bool) (ingest.Metadata, error) {
	var (
		depth            = 0
		decoder          = json.NewDecoder(reader)
		dataTagFound     = false
		dataTagValidated = false
		metaTagFound     = false
		meta             ingest.Metadata
	)

	for {
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				if !metaTagFound && !dataTagFound {
					return ingest.Metadata{}, ingest.ErrNoTagFound
				} else if !dataTagFound {
					return ingest.Metadata{}, ingest.ErrDataTagNotFound
				} else {
					return ingest.Metadata{}, ingest.ErrMetaTagNotFound
				}
			} else {
				return ingest.Metadata{}, ErrInvalidJSON
			}
		} else {
			//Validate that our data tag is actually opening correctly
			if dataTagFound && !dataTagValidated {
				if typed, ok := token.(json.Delim); ok && typed == ingest.DelimOpenSquareBracket {
					dataTagValidated = true
				} else {
					dataTagFound = false
				}
			}
			switch typed := token.(type) {
			case json.Delim:
				switch typed {
				case ingest.DelimCloseBracket, ingest.DelimCloseSquareBracket:
					depth--
				case ingest.DelimOpenBracket, ingest.DelimOpenSquareBracket:
					depth++
				}
			case string:
				if !metaTagFound && depth == 1 && typed == "meta" {
					if err := decoder.Decode(&meta); err != nil {
						slog.Warn("Found invalid metatag, skipping")
					} else if meta.Type.IsValid() {
						metaTagFound = true
					}
				}

				if !dataTagFound && depth == 1 && typed == "data" {
					dataTagFound = true
				}

				if typed == "graph" {
					// this is a generic payload
					meta = ingest.Metadata{Type: ingest.DataTypeGeneric}
					if err := ValidateGenericIngest(decoder, readToEnd); err != nil {
						return meta, err
					}
					break
				}
			}
		}

		if dataTagValidated && metaTagFound || meta.Type == ingest.DataTypeGeneric {
			break
		}
	}

	if readToEnd {
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return ingest.Metadata{}, err
		}
	}

	return meta, nil
}

func formatSchemaValidationError(err error) string {
	var sb strings.Builder
	if ve, ok := err.(*jsonschema.ValidationError); ok {
		for i, cause := range ve.Causes {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(cause.Error())
		}
	} else {
		sb.WriteString(err.Error())
	}

	return sb.String()
}

type validationError struct {
	Index   int
	Type    string // "decode" or "validation"
	Message string
}

func formatAggregateErrors(errs []validationError) string {
	var sb strings.Builder
	for _, e := range errs {
		sb.WriteString(fmt.Sprintf("- [%d] %s error: %s\n", e.Index, e.Type, e.Message))
	}
	return sb.String()
}

func DoEmbedStuff() error {
	data, err := schemaFiles.ReadFile("json_schema/node.json")
	if err != nil {
		return err
	}

	nodeSchema, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return err
	}

	c := jsonschema.NewCompiler()
	err = c.AddResource("json_schema/node.json", nodeSchema)
	if err != nil {
		return err
	}

	sch, err := c.Compile("json_schema/schema.json")
	if err != nil {
		return err
	}

	inst, _ := jsonschema.UnmarshalJSON(strings.NewReader(`{"id":1234}`))

	err = sch.Validate(inst)
	if err != nil {
		return err
	}
	return nil
}

/*
payload will be : { nodes: [], edges: [] }
*/
func ValidateGenericIngest(decoder *json.Decoder, readToEnd bool) error {

	var (
		// Initialize schemas
		c          = jsonschema.NewCompiler()
		nodeSchema = c.MustCompile("./json_schema/node.json")
		edgeSchema = c.MustCompile("./json_schema/edge.json")

		nodesFound, edgesFound = false, false
		maxErrors              = 50
		validationErrors       []validationError
	)

	// Validate an array of items (either nodes or edges)
	validateArray := func(arrayName string, validateFunc func(map[string]any) error) error {
		// if err := decoder.EatOpeningBracket(); err != nil {
		// 	return fmt.Errorf("error opening %s array: %w", arrayName, err)
		// }

		_, err := decoder.Token() // [
		if err != nil {
			return err
		}

		index := 0
		for decoder.More() {
			var item map[string]any
			if err := decoder.Decode(&item); err != nil {
				if _, ok := err.(*json.UnmarshalTypeError); ok {
					// json.UnmarshalTypeErrors are recoverable. the stream can continue advancing
					validationErrors = append(validationErrors, validationError{
						Index:   index,
						Type:    "decode",
						Message: fmt.Sprintf("type mismatch for %s[%d]: %s", arrayName, index, err),
					})
				} else {
					// json.SyntaxErrors typically corrupt the stream. abort the parse
					validationErrors = append(validationErrors, validationError{
						Index:   index,
						Type:    "decode",
						Message: fmt.Sprintf("syntax error for %s[%d]: %s. abort parse.", arrayName, index, err),
					})
					break
				}
			}

			// Validate the item using the provided validation function
			if err := validateFunc(item); err != nil {
				validationErrors = append(validationErrors, validationError{
					Index:   index,
					Type:    "validation",
					Message: fmt.Sprintf("validation failed for %s[%d]: %s", arrayName, index, err),
				})

			}

			if len(validationErrors) >= maxErrors {
				break
			}

			index++
		}

		if len(validationErrors) > 0 {
			return fmt.Errorf("validation failed with %d error(s):\n%s", len(validationErrors), formatAggregateErrors(validationErrors))
		}

		// if err := decoder.EatClosingBracket(); err != nil {
		// 	return fmt.Errorf("error closing %s array: %w", arrayName, err)
		// }
		_, err = decoder.Token() // ]
		if err != nil {
			return err
		}

		return nil
	}

	// Generic validation function for nodes and edges
	validateItem := func(item map[string]any, schema *jsonschema.Schema) error {
		if err := schema.Validate(item); err != nil {
			errorStr := formatSchemaValidationError(err)
			return fmt.Errorf("%s", errorStr)
		}
		return nil
	}

	_, err := decoder.Token() // {
	if err != nil {
		return err
	}

	// Loop to read the JSON stream and identify graph elements
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading token: %w", err)
		}

		switch typedToken := token.(type) {
		case string:
			switch typedToken {
			case "nodes":
				nodesFound = true
				if err := validateArray("nodes", func(item map[string]any) error {
					return validateItem(item, nodeSchema)
				}); err != nil {
					return err
				}
			case "edges":
				edgesFound = true
				if err := validateArray("edges", func(item map[string]any) error {
					return validateItem(item, edgeSchema)
				}); err != nil {
					return err
				}
			}
		}
	}

	_, err = decoder.Token() // }
	if err != nil {
		return err
	}

	if !nodesFound && !edgesFound {
		return ingest.ErrEmptyIngest
	}
	return nil
}

func ValidateZipFile(reader io.Reader) error {
	bytes := make([]byte, 4)
	if readBytes, err := reader.Read(bytes); err != nil {
		return err
	} else if readBytes < 4 {
		return ingest.ErrInvalidZipFile
	} else {
		for i := 0; i < 4; i++ {
			if bytes[i] != ZipMagicBytes[i] {
				return ingest.ErrInvalidZipFile
			}
		}

		_, err := io.Copy(io.Discard, reader)

		return err
	}
}
