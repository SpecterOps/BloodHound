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

// ValidateMetaTag ensures that the correct tags are present in a json file for data ingest.
// If readToEnd is set to true, the stream will read to the end of the file (needed for TeeReader)
func ValidateMetaTag(reader io.Reader, ingestSchema IngestSchema, readToEnd bool) (ingest.Metadata, error) {
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
					if err := ValidateGenericIngest(decoder, ingestSchema); err != nil {
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

type validationError struct {
	Index   int
	Message string
}

type ValidationReport struct {
	CriticalErrors   []validationError // things like json syntax errors where the document is un-parseable
	ValidationErrors []validationError // nodes and edges that dont conform to the spec
}

func (s ValidationReport) Error() string {
	var sb strings.Builder
	if len(s.CriticalErrors) > 0 {
		sb.WriteString(fmt.Sprintf("Critical errors (%d):\n%s\n", len(s.CriticalErrors), formatAggregateErrors(s.CriticalErrors)))
	}
	if len(s.ValidationErrors) > 0 {
		sb.WriteString(fmt.Sprintf("Validation errors (%d):\n%s\n", len(s.ValidationErrors), formatAggregateErrors(s.ValidationErrors)))
	}
	return sb.String()
}

func formatSchemaValidationError(err error) string {
	var sb strings.Builder
	sb.WriteString("[")
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

	sb.WriteString("]")

	return sb.String()
}

func formatAggregateErrors(errs []validationError) string {
	var sb strings.Builder
	for _, e := range errs {
		sb.WriteString(fmt.Sprintf("- error: %s\n", e.Message))
	}
	return sb.String()
}

func expectOpeningSquareBracket(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding %s array: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '[' {
		return fmt.Errorf("error opening %s array: expected '[', got %v", name, tok)
	}
	return nil
}

func expectClosingSquareBracket(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding %s array: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != ']' {
		return fmt.Errorf("error closing %s array: expected ']', got %v", name, tok)
	}
	return nil
}

func expectOpeningCurlyBracket(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding %s object: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return fmt.Errorf("error opening %s object: expected '{', got %v", name, tok)
	}
	return nil
}

func expectClosingCurlyBracket(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding %s object: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '}' {
		return fmt.Errorf("error closing %s object: expected '}', got %v", name, tok)
	}
	return nil
}

/*
payload will be : { nodes: [], edges: [] }
*/
func ValidateGenericIngest(decoder *json.Decoder, schema IngestSchema) error {
	var (
		nodesFound, edgesFound = false, false
		maxErrors              = 15

		criticalErrors   []validationError
		validationErrors []validationError

		reportCritical = func(index int, msg string) {
			criticalErrors = append(criticalErrors, validationError{Index: index, Message: msg})
		}
		reportValidation = func(index int, msg string) {
			validationErrors = append(validationErrors, validationError{Index: index, Message: msg})
		}
		hasErrors = func() bool {
			return len(validationErrors) > 0 || len(criticalErrors) > 0
		}
	)

	// Validate an array of items (either nodes or edges)
	validateArray := func(arrayName string, schema *jsonschema.Schema) {

		// eat [
		if err := expectOpeningSquareBracket(decoder, arrayName); err != nil {
			reportCritical(0, err.Error())
			return
		}

		index := 0
		for decoder.More() {
			var item map[string]any
			if err := decoder.Decode(&item); err != nil {
				switch err.(type) {
				case *json.UnmarshalTypeError:
					reportValidation(index, fmt.Sprintf("%s[%d] type mismatch: %s", arrayName, index, err))
				default:
					reportCritical(index, fmt.Sprintf("%s[%d] syntax error: %s", arrayName, index, err))
				}
			} else if err := schema.Validate(item); err != nil {
				reportValidation(index, fmt.Sprintf("%s[%d] schema validation failed: %s", arrayName, index, formatSchemaValidationError(err)))
			}

			if len(validationErrors) >= maxErrors || len(criticalErrors) > 0 {
				return
			}

			index++
		}

		// eat ]
		if err := expectClosingSquareBracket(decoder, arrayName); err != nil {
			reportCritical(0, err.Error())
			return
		}
	}

	// eat {
	if err := expectOpeningCurlyBracket(decoder, "graph"); err != nil {
		reportCritical(0, err.Error())
		return ValidationReport{
			CriticalErrors:   criticalErrors,
			ValidationErrors: validationErrors,
		}
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

		switch key := token.(type) {
		case string:
			switch key {
			case "nodes":
				nodesFound = true
				validateArray("nodes", schema.NodeSchema)
				if len(criticalErrors) > 0 {
					return ValidationReport{
						CriticalErrors:   criticalErrors,
						ValidationErrors: validationErrors,
					}
				}
			case "edges":
				edgesFound = true
				validateArray("edges", schema.EdgeSchema)
				if len(criticalErrors) > 0 {
					return ValidationReport{
						CriticalErrors:   criticalErrors,
						ValidationErrors: validationErrors,
					}
				}
			}
		}

		if len(validationErrors) >= maxErrors {
			break
		}
	}

	if err := expectClosingCurlyBracket(decoder, "graph"); err != nil {
		reportCritical(0, err.Error())
		return ValidationReport{
			CriticalErrors:   criticalErrors,
			ValidationErrors: validationErrors,
		}
	}

	if !nodesFound && !edgesFound {
		return ValidationReport{
			CriticalErrors: []validationError{
				{Message: "graph tag is empty. atleast one of nodes: [] or edges: [] is required"},
			},
		}
	}

	if hasErrors() {
		return ValidationReport{
			CriticalErrors:   criticalErrors,
			ValidationErrors: validationErrors,
		}
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
