// Copyright 2025 Specter Ops, Inc.
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
	"reflect"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/specterops/bloodhound/src/model/ingest"
)

var ZipMagicBytes = []byte{0x50, 0x4b, 0x03, 0x04}

type tagEvent struct {
	Name  string
	Depth int // (1 for top-level keys)
}

type TagScanner struct {
	decoder *json.Decoder
	depth   int
}

func NewTagScanner(decoder *json.Decoder) *TagScanner {
	return &TagScanner{
		decoder: decoder,
		depth:   0,
	}
}

func (s *TagScanner) Next() (tagEvent, error) {
	for {
		if tok, err := s.decoder.Token(); err != nil {
			return tagEvent{}, err
		} else {
			switch t := tok.(type) {
			case json.Delim:
				if t == '{' || t == '[' {
					s.depth++
				} else { // ']','}'
					s.depth--
				}
			case string:
				if s.depth == 1 {
					return tagEvent{Name: t, Depth: 1}, nil
				}
			}
		}
	}
}

// Token reads the next JSON token and updates depth internally.
func (s *TagScanner) Token() (json.Token, error) {
	tok, err := s.decoder.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); ok {
		if d == '{' || d == '[' {
			s.depth++
		} else {
			s.depth--
		}
	}
	return tok, nil
}

func decodeMetaTag(decoder *json.Decoder) (ingest.Metadata, error) {
	var m ingest.Metadata
	if err := decoder.Decode(&m); err != nil {
		slog.Warn("Found invalid metatag, skipping", slog.String("err", err.Error()))
		return ingest.Metadata{}, nil
	}
	if !m.Type.IsValid() {
		return ingest.Metadata{}, ingest.ErrMetaTagNotFound
	}
	return m, nil
}

func scanAndDetectMetaOrGraph(scanner *TagScanner, schema IngestSchema) (ingest.Metadata, error) {
	var (
		dataFound bool
		metaFound bool
		meta      ingest.Metadata
	)

	for {
		if tag, err := scanner.Next(); err != nil {
			return handleScannerError(err, dataFound, metaFound)
		} else {
			switch tag.Name {
			case "meta":
				if m, err := decodeMetaTag(scanner.decoder); err != nil {
					return m, err
				} else if m.Type.IsValid() {
					meta = m
					metaFound = true
				}
			case "data":
				// Validate that the data key is followed by an opening '[' array delimiter
				if tok, err := scanner.Token(); err != nil {
					return ingest.Metadata{}, ErrInvalidJSON
				} else if delim, ok := tok.(json.Delim); !ok || delim != '[' {
					return ingest.Metadata{}, ingest.ErrDataTagNotFound
				}
				dataFound = true
			case "graph":
				// generic ingest path
				meta = ingest.Metadata{Type: ingest.DataTypeGeneric}
				if err := ValidateGenericIngest(scanner.decoder, schema); err != nil {
					if report, ok := err.(ValidationReport); ok {
						slog.With("validation", report).Warn("generic ingest failed")
					}
					return meta, err
				}
				return meta, nil
			}

			if metaFound && dataFound {
				return meta, nil
			}
		}
	}
}

func handleScannerError(err error, dataFound, metaFound bool) (ingest.Metadata, error) {
	var m ingest.Metadata
	if errors.Is(err, io.EOF) {
		if !dataFound && !metaFound {
			return m, ingest.ErrNoTagFound
		} else if !dataFound {
			return m, ingest.ErrDataTagNotFound
		} else {
			return m, ingest.ErrMetaTagNotFound
		}
	}
	return m, ErrInvalidJSON
}

// ValidateMetaTag ensures that the correct tags are present in a json file for data ingest.
// If readToEnd is set to true, the stream will read to the end of the file (needed for TeeReader)
func ValidateMetaTag(reader io.Reader, schema IngestSchema, readToEnd bool) (ingest.Metadata, error) {
	decoder := json.NewDecoder(reader)
	scanner := NewTagScanner(decoder)

	meta, err := scanAndDetectMetaOrGraph(scanner, schema)
	if err != nil {
		return ingest.Metadata{}, err
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

func (s ValidationReport) BuildAPIError() []string {
	msgs := []string{"Error saving ingest file. File failed schema validation."}

	for _, criticalErr := range s.CriticalErrors {
		msgs = append(msgs, criticalErr.Message)
	}

	for _, valErr := range s.ValidationErrors {
		msgs = append(msgs, valErr.Message)
	}
	return msgs
}

func (s ValidationReport) Error() string {
	var sb strings.Builder
	if len(s.CriticalErrors) > 0 {
		sb.WriteString(fmt.Sprintf("(%d) critical error(s): [%s]", len(s.CriticalErrors), formatAggregateErrors(s.CriticalErrors)))
		if len(s.ValidationErrors) > 0 {
			sb.WriteString(", ")
		}
	}
	if len(s.ValidationErrors) > 0 {
		sb.WriteString(fmt.Sprintf("(%d) validation error(s): [%s]", len(s.ValidationErrors), formatAggregateErrors(s.ValidationErrors)))
	}
	return sb.String()
}

func formatSchemaValidationError(arrayName string, index int, err error) string {
	var sb strings.Builder
	if ve, ok := err.(*jsonschema.ValidationError); ok {
		numberOfViolations := len(ve.Causes)
		sb.WriteString(fmt.Sprintf("%s[%d] schema validation failed with %d error(s): ", arrayName, index, numberOfViolations))

		sb.WriteString("[")

		for i, cause := range ve.Causes {
			if i > 0 {
				sb.WriteString(", ")
			}

			// this rule fails when there is a nested object in the property bag
			if len(cause.InstanceLocation) > 0 && cause.InstanceLocation[0] == "properties" && cause.BasicOutput().KeywordLocation == "/anyOf" {
				if len(cause.InstanceLocation) > 1 {
					badPropertyName := cause.InstanceLocation[1]
					sb.WriteString(fmt.Sprintf("nested object cannot be stored as property. remove \"%s\" from properties.", badPropertyName))
				}
			} else {
				sb.WriteString(cause.Error())
			}
		}

		sb.WriteString("]")
	} else {
		sb.WriteString(err.Error())
	}
	return sb.String()
}

func formatAggregateErrors(errs []validationError) string {
	var sb strings.Builder
	for i, e := range errs {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(e.Message)
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

func isHomogeneousArray(arr []any) bool {
	if len(arr) == 0 {
		return true
	}

	firstType := reflect.TypeOf(arr[0])
	for _, v := range arr[1:] {
		if reflect.TypeOf(v) != firstType {
			return false
		}
	}
	return true
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
				reportValidation(index, formatSchemaValidationError(arrayName, index, err))
			}

			// ensure array homogeneity in "properties". unable to enforce w/ JSON Schema
			if props, ok := item["properties"].(map[string]any); ok {
				for key, val := range props {
					arr, ok := val.([]any)
					if !ok {
						continue // skip non-arrays
					}

					if !isHomogeneousArray(arr) {
						reportValidation(index, fmt.Sprintf("%s[%d] schema validation error. properties[\"%s\"] contains a mixed-type array", arrayName, index, key))
					}
				}
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
