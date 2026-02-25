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
package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
)

// Error Definitions ------------------------------------------------------------------------------

var (
	ErrMaxValidationErrors         = errors.New("reached maximum validation errors allowed")
	ErrValidationErrors            = errors.New("validator exited with validation errors")
	ErrInvalidFileConfiguration    = errors.New("invalid file configuration")
	ErrOpengraphMetadataValidation = errors.New("opengraph metadata validation error")
	ErrInvalidDataType             = errors.New("invalid data type")
)

// Validator Definitions --------------------------------------------------------------------------

type Validator struct {
	reader  io.Reader
	decoder *json.Decoder
	depth   int

	schema IngestSchema

	originalData  originalData
	opengraphData opengraphData

	maxValidationErrors int
	criticalErrors      []CriticalError
	validationErrors    []ValidationError
}

type originalData struct {
	DataFound     bool
	MetadataFound bool

	Metadata ingest.OriginalMetadata
}

type opengraphData struct {
	GraphFound    bool
	MetadataFound bool
	EdgesFound    bool
	NodesFound    bool

	Metadata ingest.OpengraphMetadata

	NodesValidated int
	EdgesValidated int
}

func NewValidator(reader io.Reader, schema IngestSchema) Validator {
	return Validator{
		reader:  reader,
		decoder: json.NewDecoder(reader),
		depth:   0,

		schema: schema,

		maxValidationErrors: 15,
		criticalErrors:      make([]CriticalError, 0),
		validationErrors:    make([]ValidationError, 0),
	}
}

// Return Definitions -----------------------------------------------------------------------------

type ValidationReport struct {
	CriticalErrors   []CriticalError
	ValidationErrors []ValidationError
}

type CriticalError struct {
	Message string
	Error   error
}

type ValidationError struct {
	Location  string
	RawObject string
	Errors    []ValidationErrorDetail
}

type ValidationErrorDetail struct {
	Location string
	Error    string
}

type ParsedData struct {
	PayloadType ingest.DataType

	LegacyMetadata ingest.OriginalMetadata
	OpengraphData  ParsedOpenGraphData
}

type ParsedOpenGraphData struct {
	Metadata       ingest.OpengraphMetadata
	NodesValidated int
	EdgesValidated int
}

// buildParsedData() aggregates data collected during ParseAndValidate() into the ParsedData struct
func (v *Validator) buildParsedData() ParsedData {
	p := ParsedData{}

	if (v.opengraphData.GraphFound || v.opengraphData.MetadataFound) && (v.originalData.MetadataFound || v.originalData.DataFound) {
		return p
	}

	if v.opengraphData.GraphFound {
		p.PayloadType = ingest.DataTypeOpenGraph
	}

	if v.opengraphData.MetadataFound {
		p.OpengraphData.Metadata = v.opengraphData.Metadata
	}

	p.OpengraphData.NodesValidated = v.opengraphData.NodesValidated
	p.OpengraphData.EdgesValidated = v.opengraphData.EdgesValidated

	if v.originalData.MetadataFound {
		p.PayloadType = v.originalData.Metadata.Type
		p.LegacyMetadata = v.originalData.Metadata
	}

	return p
}

// buildValidationReport() is a simple wrapper that aggregates critical and validation errors into a ValidationReport
func (v *Validator) buildValidationReport() ValidationReport {
	return ValidationReport{
		CriticalErrors:   v.criticalErrors,
		ValidationErrors: v.validationErrors,
	}
}

// Error Helper functions -------------------------------------------------------------------------

// reportCriticalError() is a helper function for adding a critical error
func (v *Validator) reportCriticalError(message string, err error) {
	v.criticalErrors = append(v.criticalErrors, CriticalError{Message: message, Error: err})
}

// reportValidationError() is a helper function for adding a validation error
func (v *Validator) reportValidationError(validationErr ValidationError) {
	v.validationErrors = append(v.validationErrors, validationErr)
}

// Validator state check --------------------------------------------------------------------------

// exceededValidationErrors returns true if the current number of validation errors exceeds maxValidationErrors.
// If maxValidationErrors is set to 0, this function always returns false. It is designed to be run within the
// parseOpenGraphArray function
func (v *Validator) exceededValidationErrors() bool {
	return v.maxValidationErrors != 0 && (len(v.validationErrors) >= v.maxValidationErrors)
}

// recurringFileConfigCheck() returns an error if there is an invalid mix of opengraph payload and original payload tags.
// It checks for obvious file misconfigurations such as having both the original meta and opengraph metadata tags.
// It is designed to be run every cycle of validationLoop
func (v *Validator) recurringFileConfigCheck() error {
	if v.originalData.MetadataFound && v.opengraphData.MetadataFound {
		v.reportCriticalError("cannot have both original meta tag and opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.originalData.MetadataFound && v.opengraphData.GraphFound {
		v.reportCriticalError("cannot have both original meta tag and opengraph graph tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.originalData.DataFound && v.opengraphData.MetadataFound {
		v.reportCriticalError("cannot have both original data tag and opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.originalData.DataFound && v.opengraphData.GraphFound {
		v.reportCriticalError("cannot have both original data tag and opengraph graph tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	return nil
}

// finalFileConfigCheck() returns an error if the final state of the file has an invalid arrangement of tags. This
// includes no tags at all and no graph tag being found to match an opengraph metadata tag. This is designed to be
// run after the validationLoop has completed
func (v *Validator) finalFileConfigCheck() error {
	if !v.originalData.MetadataFound && !v.originalData.DataFound && !v.opengraphData.MetadataFound && !v.opengraphData.GraphFound {
		v.reportCriticalError("no tags found", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.originalData.MetadataFound && !v.originalData.DataFound {
		v.reportCriticalError("no data tag found to match original metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if !v.originalData.MetadataFound && v.originalData.DataFound {
		v.reportCriticalError("no meta tag found to match original data tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.opengraphData.MetadataFound && !v.opengraphData.GraphFound {
		v.reportCriticalError("no graph tag found to match opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.opengraphData.GraphFound && !v.opengraphData.NodesFound && !v.opengraphData.EdgesFound {
		v.reportCriticalError("graph tag requires child nodes or edges tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	return nil
}

// External call function -------------------------------------------------------------------------

// ParseAndValidate() returns an aggregation of parsed data, a report of all errors, and an error if the
// validator didn't succeed. ParseAndValidate() will attempt to extract useful information into ParsedData
// even if there is an error
func (v *Validator) ParseAndValidate() (ParsedData, ValidationReport, error) {
	if err := v.enterObject(); err != nil {
		v.reportCriticalError("failed to enter json object", err)
		return v.buildParsedData(), v.buildValidationReport(), err
	}

	valLoopErr := v.validationLoop()
	// This multireader ensures that bytes included in the json decoder's buffer. This guarantees that ALL bytes are read from the io.Reader
	_, readToEndErr := io.Copy(io.Discard, io.MultiReader(v.decoder.Buffered(), v.reader))
	if valLoopErr != nil && readToEndErr != nil {
		v.reportCriticalError("failed to read file to end", readToEndErr)
		return v.buildParsedData(), v.buildValidationReport(), errors.Join(valLoopErr, readToEndErr)
	} else if valLoopErr == nil && readToEndErr != nil {
		v.reportCriticalError("failed to read file to end", readToEndErr)
		return v.buildParsedData(), v.buildValidationReport(), readToEndErr
	} else if valLoopErr != nil {
		return v.buildParsedData(), v.buildValidationReport(), valLoopErr
	}

	if err := v.finalFileConfigCheck(); err != nil {
		return v.buildParsedData(), v.buildValidationReport(), err
	} else if len(v.validationErrors) > 0 {
		return v.buildParsedData(), v.buildValidationReport(), ErrValidationErrors
	} else {
		return v.buildParsedData(), v.buildValidationReport(), nil
	}
}

// Validation Loop functions ----------------------------------------------------------------------

// validationLoop() is the primary driver behind the file validation. It walks through the file and directs to
// child validation functions
func (v *Validator) validationLoop() error {
	for {
		if err := v.recurringFileConfigCheck(); err != nil {
			return err
		} else if tag, exitedBlock, err := v.nextTagAtDepth(1); err != nil {
			v.reportCriticalError("failed parsing top level tag", err)
			return err
		} else if exitedBlock {
			return nil
		} else {
			switch tag {
			case "meta":
				if v.originalData.MetadataFound {
					v.reportCriticalError("duplicate top level meta tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				v.originalData.MetadataFound = true

				originalMetadata, err := v.handleOriginalMetadata()
				if err != nil {
					return err
				}

				v.originalData.Metadata = originalMetadata
			case "data":
				if v.originalData.DataFound {
					v.reportCriticalError("duplicate top level data tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				v.originalData.DataFound = true

				err := v.handleData()
				if err != nil {
					return err
				}
			case "metadata":
				if v.opengraphData.MetadataFound {
					v.reportCriticalError("duplicate top level metadata tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				v.opengraphData.MetadataFound = true

				opengraphMetadata, err := v.handleOpenGraphMetadata()
				if err != nil {
					return err
				}

				v.opengraphData.Metadata = opengraphMetadata
			case "graph":
				if v.opengraphData.GraphFound {
					v.reportCriticalError("duplicate top level graph tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				v.opengraphData.GraphFound = true

				err := v.handleGraph()
				if err != nil {
					return err
				}
			default:
				v.reportCriticalError(fmt.Sprintf("unrecognized top level tag: %s", tag), ErrInvalidFileConfiguration)
				return ErrInvalidFileConfiguration
			}
		}
	}
}

// handleOriginalMetadata() parses and validates original metadata after a "meta" tag is found at the top level
func (v *Validator) handleOriginalMetadata() (ingest.OriginalMetadata, error) {
	var originalMetadata ingest.OriginalMetadata

	if err := v.decoder.Decode(&originalMetadata); err != nil {
		v.reportCriticalError("failed to decode original metadata", err)
		return ingest.OriginalMetadata{}, err
	} else if !originalMetadata.Type.IsValidOriginalType() {
		v.reportCriticalError("invalid original metadata data type", ErrInvalidDataType)
		return ingest.OriginalMetadata{}, ErrInvalidDataType
	}

	return originalMetadata, nil
}

// handleOpenGraphMetadata() parses and validates opengraph metadata after the "metadata" tag is found at the top level
func (v *Validator) handleOpenGraphMetadata() (ingest.OpengraphMetadata, error) {
	var (
		rawObject         json.RawMessage
		metaValidate      any
		opengraphMetadata ingest.OpengraphMetadata
		schemaErr         *jsonschema.ValidationError
	)

	if err := v.decoder.Decode(&rawObject); err != nil {
		v.reportCriticalError("failed decoding opengraph metadata to raw object", err)
		return ingest.OpengraphMetadata{}, err
	} else if err := json.Unmarshal(rawObject, &metaValidate); err != nil {
		v.reportCriticalError("failed unmarshalling json to any", err)
		return ingest.OpengraphMetadata{}, err
	} else if err := v.schema.MetaSchema.Validate(metaValidate); errors.As(err, &schemaErr) {
		v.reportCriticalError("opengraph metadata failed validation", ErrOpengraphMetadataValidation)

		errorDetails, err := extractJsonSchemaErrors(schemaErr)
		if err != nil {
			v.reportCriticalError("failed extracting json schema errors at /metadata", err)
			return ingest.OpengraphMetadata{}, err
		}

		v.reportValidationError(ValidationError{Location: "/metadata", RawObject: string(rawObject), Errors: errorDetails})
		return ingest.OpengraphMetadata{}, ErrOpengraphMetadataValidation
	} else if err != nil {
		v.reportCriticalError("schema validation returned non validation error for opengraph metadata", err)
		return ingest.OpengraphMetadata{}, err
	} else if err := json.Unmarshal(rawObject, &opengraphMetadata); err != nil {
		v.reportCriticalError("failed unmarshalling json to opengraph Metadata", err)
		return ingest.OpengraphMetadata{}, err
	} else {
		return opengraphMetadata, nil
	}
}

// handleData() is called after the "data" tag is found. Currently this simply checks that the next token
// is an opening array then passes through.
func (v *Validator) handleData() error {
	if err := v.enterArray(); err != nil {
		v.reportCriticalError("failed to enter data array", err)
		return err
	}

	return nil
}

// handleGraph() parses and validates opengraph specific data after the "graph" tag is found at the top level
func (v *Validator) handleGraph() error {
	if err := v.enterObject(); err != nil {
		v.reportCriticalError("failed to enter graph object", err)
		return err
	}

	for {
		if tag, exitedBlock, err := v.nextTagAtDepth(2); err != nil {
			return err
		} else if exitedBlock {
			return nil
		} else {
			switch tag {
			case "nodes":
				if v.opengraphData.NodesFound {
					v.reportCriticalError("duplicate graph nodes tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}
				v.opengraphData.NodesFound = true

				numItems, err := v.handleOpenGraphArray(tag, v.schema.NodeSchema)
				v.opengraphData.NodesValidated = numItems
				if err != nil {
					return err
				}
			case "edges":
				if v.opengraphData.EdgesFound {
					v.reportCriticalError("duplicate graph edges tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}
				v.opengraphData.EdgesFound = true

				numItems, err := v.handleOpenGraphArray(tag, v.schema.EdgeSchema)
				v.opengraphData.EdgesValidated = numItems
				if err != nil {
					return err
				}
			default:
				v.reportCriticalError(fmt.Sprintf("unrecognized graph child tag: %s", tag), ErrInvalidFileConfiguration)
				return ErrInvalidFileConfiguration
			}
		}
	}
}

// decodedArrayObject is a helper object to extract the raw JSON string and the unmarshaled Go object
// for validation
type decodedArrayObject struct {
	RawObject string
	Object    any
}

func (s *decodedArrayObject) UnmarshalJSON(bytes []byte) error {
	s.RawObject = string(bytes)

	return json.Unmarshal(bytes, &s.Object)
}

// handleOpenGraphArray() parses and validates all objects in the "nodes" or "edges" arrays inside the
// "graph" tag. It is the primary driver for OpenGraph payload validation.
func (v *Validator) handleOpenGraphArray(arrayName string, schema *jsonschema.Schema) (int, error) {
	index := 0

	if err := v.enterArray(); err != nil {
		v.reportCriticalError(fmt.Sprintf("failed to enter graph %s array", arrayName), err)
		return index, err
	}

	for v.decoder.More() {
		var item decodedArrayObject

		if err := v.decoder.Decode(&item); err != nil {
			v.reportCriticalError(fmt.Sprintf("failed to decode %s array object", arrayName), err)
			return index, err
		}

		if err := schema.Validate(item.Object); err != nil {
			var (
				location  = fmt.Sprintf("/graph/%s[%d]", arrayName, index)
				schemaErr *jsonschema.ValidationError
			)

			if ok := errors.As(err, &schemaErr); !ok {
				v.reportCriticalError(fmt.Sprintf("schema validation returned non validation error at /graph/%s[%d]", arrayName, index), err)
				return index, err
			} else if errorDetails, err := extractJsonSchemaErrors(schemaErr); err != nil {
				v.reportCriticalError(fmt.Sprintf("failed extracting json schema errors at /graph/%s[%d]", arrayName, index), err)
				return index, err
			} else {
				v.reportValidationError(ValidationError{
					Location:  location,
					RawObject: item.RawObject,
					Errors:    errorDetails,
				})
			}
		}

		index++

		if v.exceededValidationErrors() {
			return index, ErrMaxValidationErrors
		}
	}

	return index, nil
}

// extractJsonSchemaErrors() is a helper function that takes the errors returned by santhosh-tekuri/jsonschema and
// make turn them into a format agreeable with ValidationReport
func extractJsonSchemaErrors(ve *jsonschema.ValidationError) ([]ValidationErrorDetail, error) {
	var (
		errMap       = make(map[string]string, 0)
		errorDetails = make([]ValidationErrorDetail, 0)
	)

	for _, cause := range ve.Causes {
		output := cause.BasicOutput()

		if output == nil {
			return []ValidationErrorDetail{}, fmt.Errorf("failed to extract schema validation error BasicOutput")
		} else if output.Error != nil {
			errMap[output.InstanceLocation] = output.Error.String()
		} else if output.Errors != nil {
			for _, e := range output.Errors {
				if e.Error == nil {
					return []ValidationErrorDetail{}, fmt.Errorf("failed to extract output error from output unit errors")
				}

				if strings.HasPrefix(e.InstanceLocation, "/properties/") {
					locSplit := strings.Split(e.InstanceLocation, "/")

					newLocation := fmt.Sprintf("/%s/%s", locSplit[1], locSplit[2])
					if _, found := errMap[newLocation]; !found {
						errMap[newLocation] = "invalid type"
					}
				} else {
					if _, found := errMap[e.InstanceLocation]; !found {
						errMap[e.InstanceLocation] = e.Error.String()
					}
				}
			}
		} else {
			return []ValidationErrorDetail{}, fmt.Errorf("failed to extract Error or Errors from cause.BasicOutput()")
		}
	}

	for loc, err := range errMap {
		errorDetails = append(errorDetails, ValidationErrorDetail{
			Location: loc,
			Error:    err,
		})
	}

	return errorDetails, nil
}

// Scanner functions ------------------------------------------------------------------------------

// enterObject() consumes the next JSON token. Returns an error if the next token is not {
func (v *Validator) enterObject() error {
	t, err := v.nextToken()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); !ok || delim != ingest.DelimOpenBracket {
		return fmt.Errorf("expected open bracket")
	}

	return nil
}

// enterArray() consumes the next JSON token. Returns an error if the next token is not [
func (v *Validator) enterArray() error {
	t, err := v.nextToken()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); !ok || delim != ingest.DelimOpenSquareBracket {
		return fmt.Errorf("expected open square bracket")
	}

	return nil
}

// nextTagAtDepth() consumes tokens until it finds the next tag at the specified depth, returning that token. If
// nextTagAtDepth exits the specified depth (depth decreases), then the function returns true in the second argument.
func (v *Validator) nextTagAtDepth(depth int) (string, bool, error) {
	for {
		t, err := v.nextToken()
		if err != nil {
			return "", false, err
		}

		if v.depth < depth {
			return "", true, nil
		}

		tag, ok := t.(string)
		if !ok {
			continue
		}

		if v.depth == depth {
			return tag, false, nil
		}
	}
}

// nextToken() consumes the next JSON token, returning it. This function should be the only one used for
// interacting with the underlying JSON file because it keeps track of file depth.
func (v *Validator) nextToken() (json.Token, error) {
	tok, err := v.decoder.Token()
	if err != nil {
		return nil, err
	}

	if d, ok := tok.(json.Delim); ok {
		if d == ingest.DelimOpenBracket || d == ingest.DelimOpenSquareBracket {
			v.depth++
		} else {
			v.depth--
		}
	}

	return tok, nil
}
