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

package upload

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"reflect"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"

	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
)

var ZipMagicBytes = []byte{0x50, 0x4b, 0x03, 0x04}

// ParseAndValidatePayload scans a JSON stream to detect and validate the metadata tag
// required for ingesting graph data. It ensures that either top-level "meta" and "data" tags
// or a "graph" tag is present. "meta"/"data" are for existing hound collections (ad and azure).
// The "graph" tag supports generic ingest.
//
// If shouldValidateGraph is true, the function will also attempt to validate the
// presence and structure of a "graph" tag alongside the metadata.
//
// If readToEnd is set to true, the stream will read to the end of the file (needed for TeeReader)
func ParseAndValidatePayload(reader io.Reader, schema IngestSchema, shouldValidateGraph, readToEnd bool) (ingest.Metadata, error) {
	decoder := json.NewDecoder(reader)
	scanner := newTagScanner(decoder)

	meta, err := scanAndDetectMetaOrGraph(scanner, shouldValidateGraph, schema)
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

// ValidateGraph validates a generic ingest graph payload from a JSON stream.
// The input is expected to be a JSON object containing one or both of the keys
// "nodes" and "edges", each mapping to an array of graph elements.
// Each element is validated against the corresponding JSON Schema provided in the
// IngestSchema struct. In addition to schema validation, this function enforces
// constraints not expressible in JSON Schema, such as nested objects and type homogeneity in
// array-valued properties.
//
// If critical errors (e.g., malformed JSON, missing brackets) or a sufficient number
// of validation errors are encountered, a ValidationReport is returned as an error.
// If no errors are found, the function returns nil.
func ValidateGraph(decoder *json.Decoder, schema IngestSchema) error {
	v := &validator{
		decoder:    decoder,
		nodeSchema: schema.NodeSchema,
		edgeSchema: schema.EdgeSchema,
		metaSchema: schema.MetaSchema,
		maxErrors:  15,
	}

	if err := expectOpenObject(decoder, "graph"); err != nil {
		v.reportCritical(0, err.Error())
		return v.report()
	}

	for decoder.More() {
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("error reading token: %w", err)
		} else {
			key, ok := token.(string)
			if !ok {
				continue // ignore non-string keys
			}

			switch key {
			case "nodes":
				v.nodesFound = true
				v.validateArray("nodes", v.nodeSchema)
				if len(v.criticalErrors) > 0 {
					return v.report()
				}
			case "edges":
				v.edgesFound = true
				v.validateArray("edges", v.edgeSchema)
				if len(v.criticalErrors) > 0 {
					return v.report()
				}
			case "metadata":
				v.validateMetadata(v.metaSchema)
				if len(v.criticalErrors) > 0 {
					return v.report()
				}
			}

			if len(v.validationErrors) >= v.maxErrors {
				break
			}
		}
	}

	if err := expectClosingObject(decoder, "graph"); err != nil {
		v.reportCritical(0, err.Error())
		return v.report()
	}

	if !v.nodesFound && !v.edgesFound {
		v.reportCritical(0, "graph tag is empty. at least one of nodes: [] or edges: [] is required")
	}

	return v.report()
}

type tagScanner struct {
	decoder *json.Decoder
	depth   int
}

func newTagScanner(decoder *json.Decoder) *tagScanner {
	return &tagScanner{
		decoder: decoder,
		depth:   0,
	}
}

// nextTopLevelTag only emits string keys at depth 1
func (s *tagScanner) nextTopLevelTag() (string, error) {
	for {
		if tok, err := s.decoder.Token(); err != nil {
			return "", err
		} else {
			switch t := tok.(type) {
			case json.Delim:
				if t == ingest.DelimOpenBracket || t == ingest.DelimOpenSquareBracket {
					s.depth++
				} else { // ']','}'
					s.depth--
				}
			case string:
				if s.depth == 1 {
					return t, nil
				}
			}
		}
	}
}

// nextToken reads the next JSON nextToken and updates depth internally.
func (s *tagScanner) nextToken() (json.Token, error) {
	tok, err := s.decoder.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); ok {
		if d == ingest.DelimOpenBracket || d == ingest.DelimOpenSquareBracket {
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

func scanAndDetectMetaOrGraph(scanner *tagScanner, shouldValidateGraph bool, schema IngestSchema) (ingest.Metadata, error) {
	var (
		dataFound bool
		metaFound bool
		meta      ingest.Metadata
	)

	for {
		if tag, err := scanner.nextTopLevelTag(); err != nil {
			return handleScannerError(err, dataFound, metaFound)
		} else {
			switch tag {
			case "meta":
				if m, err := decodeMetaTag(scanner.decoder); err != nil {
					return m, err
				} else if m.Type.IsValid() {
					meta = m
					metaFound = true
				}
			case "data":
				// Validate that the data key is followed by an opening '[' array delimiter
				if tok, err := scanner.nextToken(); err != nil {
					return ingest.Metadata{}, ErrInvalidJSON
				} else if delim, ok := tok.(json.Delim); !ok || delim != ingest.DelimOpenSquareBracket {
					slog.Warn("expected '[' after data key", slog.Any("got", tok))
					return ingest.Metadata{}, ingest.ErrDataTagNotFound
				}
				dataFound = true
			case "graph":
				// enforce mutual exclusivity
				if dataFound || metaFound {
					return ingest.Metadata{}, ingest.ErrMixedIngestFormat
				}
				// opengraph ingest path
				meta = ingest.Metadata{Type: ingest.DataTypeOpenGraph}
				if shouldValidateGraph {
					if err := ValidateGraph(scanner.decoder, schema); err != nil {
						if report, ok := err.(ValidationReport); ok {
							slog.With("validation", report).Warn("opengraph ingest failed")
						}
						return meta, err
					}
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

			isPropertyError := len(cause.InstanceLocation) > 1 && cause.InstanceLocation[0] == "properties"
			propertyName := ""
			if isPropertyError {
				propertyName = cause.InstanceLocation[1]
			}

			switch {
			// Case: property value is an object (not allowed)
			case isPropertyError && isTypeError(cause, "object"):
				sb.WriteString(fmt.Sprintf(
					"Invalid property '%s': objects are not allowed in the property bag. Use only strings, numbers, booleans, nulls, or arrays of these types.",
					propertyName,
				))

			// Case: array contains a nested object (also not allowed)
			case isPropertyError && isNotError(cause):
				sb.WriteString(fmt.Sprintf(
					"Invalid property '%s': array contains an object. Arrays must contain only primitive values (string, number, boolean, or null).",
					propertyName,
				))

			default:
				sb.WriteString(cause.Error())
			}
		}

		sb.WriteString("]")
	} else {
		sb.WriteString(err.Error())
	}
	return sb.String()
}

func isTypeError(cause *jsonschema.ValidationError, got string) bool {
	typeErr, ok := cause.ErrorKind.(*kind.Type)
	return ok && typeErr.Got == got
}

func isNotError(cause *jsonschema.ValidationError) bool {
	_, ok := cause.ErrorKind.(*kind.Not)
	return ok
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

// ReadZippedFile - Util Function to help read zipped files
func ReadZippedFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
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

type validator struct {
	decoder          *json.Decoder
	nodeSchema       *jsonschema.Schema
	edgeSchema       *jsonschema.Schema
	metaSchema       *jsonschema.Schema
	maxErrors        int
	nodesFound       bool
	edgesFound       bool
	criticalErrors   []validationError
	validationErrors []validationError
}

func (v *validator) reportCritical(index int, msg string) {
	v.criticalErrors = append(v.criticalErrors, validationError{Index: index, Message: msg})
}

func (v *validator) reportValidation(index int, msg string) {
	v.validationErrors = append(v.validationErrors, validationError{Index: index, Message: msg})
}

func (v *validator) hasErrors() bool {
	return len(v.criticalErrors) > 0 || len(v.validationErrors) > 0
}

func (v *validator) validateArray(arrayName string, schema *jsonschema.Schema) {
	if err := expectOpenArray(v.decoder, arrayName); err != nil {
		v.reportCritical(0, err.Error())
		return
	}

	index := 0
	for v.decoder.More() {
		var item map[string]any
		if err := v.decoder.Decode(&item); err != nil {
			switch err.(type) {
			case *json.UnmarshalTypeError:
				v.reportValidation(index, fmt.Sprintf("%s[%d] type mismatch: %s", arrayName, index, err))
			default:
				v.reportCritical(index, fmt.Sprintf("%s[%d] syntax error: %s", arrayName, index, err))
			}
		} else if err := schema.Validate(item); err != nil {
			v.reportValidation(index, formatSchemaValidationError(arrayName, index, err))
		}

		if props, ok := item["properties"].(map[string]any); ok {
			for key, val := range props {
				if arr, ok := val.([]any); ok && !isHomogeneousArray(arr) {
					v.reportValidation(index, fmt.Sprintf("%s[%d] schema validation error. properties[\"%s\"] contains a mixed-type array", arrayName, index, key))
				}
			}
		}

		if len(v.validationErrors) >= v.maxErrors || len(v.criticalErrors) > 0 {
			return
		}
		index++
	}

	if err := expectClosingArray(v.decoder, arrayName); err != nil {
		v.reportCritical(0, err.Error())
	}
}

func (v *validator) validateMetadata(metadataSchema *jsonschema.Schema) {
	var item map[string]any

	if err := v.decoder.Decode(&item); err != nil {
		switch err.(type) {
		case *json.UnmarshalTypeError:
			v.reportValidation(0, fmt.Sprintf("%s[%d] type mismatch: %s", "metadata", 0, err))
		default:
			v.reportCritical(0, fmt.Sprintf("%s[%d] syntax error: %s", "metadata", 0, err))
		}
	} else if err := metadataSchema.Validate(item); err != nil {
		v.reportValidation(0, formatSchemaValidationError("metadata", 0, err))
	}
}

func (v *validator) report() error {
	if v.hasErrors() {
		return ValidationReport{
			CriticalErrors:   v.criticalErrors,
			ValidationErrors: v.validationErrors,
		}
	}
	return nil
}
