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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/specterops/bloodhound/packages/go/chow/ingest"
)

// Error Definitions ------------------------------------------------------------------------------

var (
	ErrMaxValidationErrors         = errors.New("reached maximum validation errors allowed")
	ErrValidationErrors            = errors.New("validator exited with validation errors")
	ErrInvalidFileConfiguration    = errors.New("invalid file configuration")
	ErrOpengraphMetadataValidation = errors.New("opengraph metadata validation error")
	ErrInvalidDataType             = errors.New("invalid data type")
)

const (
	delimOpenObject  = json.Delim('{')
	delimCloseObject = json.Delim('}')
	delimOpenArray   = json.Delim('[')
	delimCloseArray  = json.Delim(']')
)

// Validator Definitions --------------------------------------------------------------------------

type Validator struct {
	reader  io.Reader
	decoder *json.Decoder
	depth   int

	schema Schema

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

func NewValidator(reader io.Reader, schema Schema) Validator {
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

func (s ValidationError) Error() string {
	var (
		details = make([]string, 0, len(s.Errors))
		message = "validation error"
	)

	if s.Location != "" {
		message = fmt.Sprintf("%s at %s", message, s.Location)
	}

	for _, validationErrorDetail := range s.Errors {
		if validationErrorDetail.Location != "" {
			details = append(details, fmt.Sprintf("%s: %s", validationErrorDetail.Location, validationErrorDetail.Error))
		} else if validationErrorDetail.Error != "" {
			details = append(details, validationErrorDetail.Error)
		}
	}

	if len(details) > 0 {
		message = fmt.Sprintf("%s: %s", message, strings.Join(details, "; "))
	}

	return message
}

type ValidationErrorDetail struct {
	Location string
	Error    string
}

type ParsedData struct {
	PayloadType ingest.DataType

	// SharpHound and AzureHound Style Metadata
	OriginalData ParsedOriginalData
	// OpenGraph Style Metadata
	OpengraphData ParsedOpenGraphData
}

type ParsedOriginalData struct {
	MetadataFound bool
	Metadata      ingest.OriginalMetadata
}

type ParsedOpenGraphData struct {
	MetadataFound  bool
	Metadata       ingest.OpengraphMetadata
	NodesValidated int
	EdgesValidated int
}

// buildValidatedData() aggregates data collected during ParseAndValidate() into the ParsedData struct.
// It is specific to the validation path and relies on signals (GraphFound, NodesValidated, etc.)
// that are only populated by validationLoop. ParseMetadata() builds its result inline rather than
// using this helper.
func (s *Validator) buildValidatedData() ParsedData {
	p := ParsedData{}

	if (s.opengraphData.GraphFound || s.opengraphData.MetadataFound) && (s.originalData.MetadataFound || s.originalData.DataFound) {
		return p
	}

	if s.opengraphData.GraphFound {
		p.PayloadType = ingest.DataTypeOpenGraph
	}

	if s.opengraphData.MetadataFound {
		p.OpengraphData.MetadataFound = true
		p.OpengraphData.Metadata = s.opengraphData.Metadata
	}

	p.OpengraphData.NodesValidated = s.opengraphData.NodesValidated
	p.OpengraphData.EdgesValidated = s.opengraphData.EdgesValidated

	if s.originalData.MetadataFound {
		p.PayloadType = s.originalData.Metadata.Type
		p.OriginalData.MetadataFound = true
		p.OriginalData.Metadata = s.originalData.Metadata
	}

	return p
}

// buildValidationReport() is a simple wrapper that aggregates critical and validation errors into a ValidationReport
func (s *Validator) buildValidationReport() ValidationReport {
	return ValidationReport{
		CriticalErrors:   s.criticalErrors,
		ValidationErrors: s.validationErrors,
	}
}

// result() is a helper for returning the current parsed data, validation report, and provided error.
func (s *Validator) result(err error) (ParsedData, ValidationReport, error) {
	return s.buildValidatedData(), s.buildValidationReport(), err
}

// Error Helper functions -------------------------------------------------------------------------

// reportCriticalError() is a helper function for adding a critical error
func (s *Validator) reportCriticalError(message string, err error) {
	s.criticalErrors = append(s.criticalErrors, CriticalError{Message: message, Error: err})
}

// reportValidationError() is a helper function for adding a validation error
func (s *Validator) reportValidationError(validationErr ValidationError) {
	s.validationErrors = append(s.validationErrors, validationErr)
}

// Validator state check --------------------------------------------------------------------------

// exceededValidationErrors returns true if the current number of validation errors exceeds maxValidationErrors.
// If maxValidationErrors is set to 0, this function always returns false. It is designed to be run within the
// parseOpenGraphArray function
func (s *Validator) exceededValidationErrors() bool {
	return s.maxValidationErrors != 0 && (len(s.validationErrors) >= s.maxValidationErrors)
}

// recurringFileConfigCheck() returns an error if there is an invalid mix of opengraph payload and original payload tags.
// It checks for obvious file misconfigurations such as having both the original meta and opengraph metadata tags.
// It is designed to be run every cycle of validationLoop
func (s *Validator) recurringFileConfigCheck() error {
	if s.originalData.MetadataFound && s.opengraphData.MetadataFound {
		s.reportCriticalError("cannot have both original meta tag and opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if s.originalData.MetadataFound && s.opengraphData.GraphFound {
		s.reportCriticalError("cannot have both original meta tag and opengraph graph tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if s.originalData.DataFound && s.opengraphData.MetadataFound {
		s.reportCriticalError("cannot have both original data tag and opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if s.originalData.DataFound && s.opengraphData.GraphFound {
		s.reportCriticalError("cannot have both original data tag and opengraph graph tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	return nil
}

// finalFileConfigCheck() returns an error if the final state of the file has an invalid arrangement of tags. This
// includes no recognized payload tags and no graph tag being found to match an opengraph metadata tag. This is
// designed to be run after the validationLoop has completed.
func (s *Validator) finalFileConfigCheck() error {
	if !s.originalData.MetadataFound && !s.originalData.DataFound && !s.opengraphData.MetadataFound && !s.opengraphData.GraphFound {
		s.reportCriticalError("no valid payload tags found", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if s.originalData.MetadataFound && !s.originalData.DataFound {
		s.reportCriticalError("no data tag found to match original metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if !s.originalData.MetadataFound && s.originalData.DataFound {
		s.reportCriticalError("no meta tag found to match original data tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if s.opengraphData.MetadataFound && !s.opengraphData.GraphFound {
		s.reportCriticalError("no graph tag found to match opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if s.opengraphData.GraphFound && !s.opengraphData.NodesFound && !s.opengraphData.EdgesFound {
		s.reportCriticalError("graph tag requires child nodes or edges tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	return nil
}

// External call function -------------------------------------------------------------------------

// ParseAndValidate() returns an aggregation of parsed data, a report of all errors, and an error if the
// validator didn't succeed. ParseAndValidate() will attempt to extract useful information into ParsedData
// even if there is an error
func (s *Validator) ParseAndValidate() (ParsedData, ValidationReport, error) {
	if err := s.enterObject(); err != nil {
		s.reportCriticalError("failed to enter json object", err)
		return s.result(err)
	}

	valLoopErr := s.validationLoop()

	if err := s.readToEnd(valLoopErr); err != nil {
		return s.result(err)
	}

	return s.result(s.finalizeParse())
}

// ParseMetadata() walks the top-level JSON object and extracts metadata (either legacy "meta" or
// opengraph "metadata") without performing schema validation of the payload body. It returns as soon
// as a metadata tag is successfully decoded; the remainder of the reader is not consumed.
func (s *Validator) ParseMetadata() (ParsedData, error) {
	if err := s.enterObject(); err != nil {
		s.reportCriticalError("failed to enter json object", err)
		return ParsedData{}, err
	}

	err := s.parseLoop()

	p := ParsedData{}
	switch {
	case s.originalData.MetadataFound:
		p.PayloadType = s.originalData.Metadata.Type
		p.OriginalData.MetadataFound = true
		p.OriginalData.Metadata = s.originalData.Metadata
	case s.opengraphData.MetadataFound:
		p.PayloadType = ingest.DataTypeOpenGraph
		p.OpengraphData.MetadataFound = true
		p.OpengraphData.Metadata = s.opengraphData.Metadata
	case s.opengraphData.GraphFound:
		p.PayloadType = ingest.DataTypeOpenGraph
	}

	return p, err
}

// readToEnd() checks for trailing input if validation succeeded, then consumes all remaining bytes from the decoder
// buffer and reader while preserving any existing loop error.
func (s *Validator) readToEnd(loopErr error) error {
	errToReturn := loopErr
	if errToReturn == nil {
		if err := s.expectEOF(); err != nil {
			s.reportCriticalError("expected to hit the end of the file", err)
			errToReturn = err
		}
	}

	// This multireader ensures that bytes included in the json decoder's buffer. This guarantees that ALL bytes are read from the io.Reader
	_, readToEndErr := io.Copy(io.Discard, io.MultiReader(s.decoder.Buffered(), s.reader))
	if readToEndErr != nil {
		s.reportCriticalError("failed to read file to end", readToEndErr)
	}

	if errToReturn != nil && readToEndErr != nil {
		return errors.Join(errToReturn, readToEndErr)
	}

	if readToEndErr != nil {
		return readToEndErr
	}

	return errToReturn
}

// finalizeParse() performs the final post-parse validation checks and collapses validation errors into a single error.
func (s *Validator) finalizeParse() error {
	if err := s.finalFileConfigCheck(); err != nil {
		return err
	}

	if len(s.validationErrors) > 0 {
		return ErrValidationErrors
	}

	return nil
}

// Loop functions ----------------------------------------------------------------------

// validationLoop() is the primary driver behind the file validation. It walks through the file and directs to
// child validation functions
func (s *Validator) validationLoop() error {
	for {
		if err := s.recurringFileConfigCheck(); err != nil {
			return err
		} else if tag, exitedBlock, err := s.nextTagAtDepth(1); err != nil {
			s.reportCriticalError("failed parsing top level tag", err)
			return err
		} else if exitedBlock {
			return nil
		} else {
			switch tag {
			case "meta":
				if s.originalData.MetadataFound {
					s.reportCriticalError("duplicate top level meta tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				s.originalData.MetadataFound = true

				originalMetadata, err := s.handleOriginalMetadata()
				if err != nil {
					return err
				}

				s.originalData.Metadata = originalMetadata
			case "data":
				if s.originalData.DataFound {
					s.reportCriticalError("duplicate top level data tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				s.originalData.DataFound = true

				err := s.handleData()
				if err != nil {
					return err
				}
			case "metadata":
				if s.opengraphData.MetadataFound {
					s.reportCriticalError("duplicate top level metadata tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				s.opengraphData.MetadataFound = true

				opengraphMetadata, err := s.handleOpenGraphMetadata()
				if err != nil {
					return err
				}

				s.opengraphData.Metadata = opengraphMetadata
			case "graph":
				if s.opengraphData.GraphFound {
					s.reportCriticalError("duplicate top level graph tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				s.opengraphData.GraphFound = true

				err := s.handleGraph()
				if err != nil {
					return err
				}
			default:
				if err := s.skipValue(); err != nil {
					s.reportCriticalError(fmt.Sprintf("failed to skip unrecognized top level tag: %s", tag), err)
					return err
				}
			}
		}
	}
}

// parseLoop() walks the top-level object looking for tags that identify the payload shape
// ("meta", "metadata"), decoding any metadata tag into the Validator's internal state and
// returning as soon as a tag that uniquely identifies the payload type is found or the
// top-level object is exited.
func (s *Validator) parseLoop() error {
	for {
		if tag, exitedBlock, err := s.nextTagAtDepth(1); err != nil {
			s.reportCriticalError("failed parsing top level tag", err)
			return err
		} else if exitedBlock {
			return nil
		} else {
			switch tag {
			case "meta":
				s.originalData.MetadataFound = true

				var metadata ingest.OriginalMetadata
				if err := s.decoder.Decode(&metadata); err != nil {
					s.reportCriticalError("failed to decode original metadata", err)
					return err
				}

				s.originalData.Metadata = metadata
				return nil
			case "metadata":
				s.opengraphData.MetadataFound = true

				var metadata ingest.OpengraphMetadata
				if err := s.decoder.Decode(&metadata); err != nil {
					s.reportCriticalError("failed to decode opengraph metadata", err)
					return err
				}

				s.opengraphData.Metadata = metadata
				return nil
			case "graph":
				s.opengraphData.GraphFound = true
			default:
			}
		}
	}
}

// handleOriginalMetadata() parses and validates original metadata after a "meta" tag is found at the top level
func (s *Validator) handleOriginalMetadata() (ingest.OriginalMetadata, error) {
	var originalMetadata ingest.OriginalMetadata

	if err := s.decoder.Decode(&originalMetadata); err != nil {
		s.reportCriticalError("failed to decode original metadata", err)
		return ingest.OriginalMetadata{}, err
	} else if !originalMetadata.Type.IsValidOriginalType() {
		s.reportCriticalError("invalid original metadata data type", ErrInvalidDataType)
		return ingest.OriginalMetadata{}, ErrInvalidDataType
	}

	return originalMetadata, nil
}

// handleOpenGraphMetadata() parses and validates opengraph metadata after the "metadata" tag is found at the top level
func (s *Validator) handleOpenGraphMetadata() (ingest.OpengraphMetadata, error) {
	var (
		rawObject         json.RawMessage
		metaValidate      any
		opengraphMetadata ingest.OpengraphMetadata
		schemaErr         *jsonschema.ValidationError
	)

	if err := s.decoder.Decode(&rawObject); err != nil {
		s.reportCriticalError("failed decoding opengraph metadata to raw object", err)
		return ingest.OpengraphMetadata{}, err
	} else if err := json.Unmarshal(rawObject, &metaValidate); err != nil {
		s.reportCriticalError("failed unmarshalling json to any", err)
		return ingest.OpengraphMetadata{}, err
	} else if err := s.schema.MetaSchema.Validate(metaValidate); errors.As(err, &schemaErr) {
		s.reportCriticalError("opengraph metadata failed validation", ErrOpengraphMetadataValidation)

		errorDetails, err := extractJsonSchemaErrors(schemaErr)
		if err != nil {
			s.reportCriticalError("failed extracting json schema errors at /metadata", err)
			return ingest.OpengraphMetadata{}, err
		}

		s.reportValidationError(ValidationError{Location: "/metadata", RawObject: string(rawObject), Errors: errorDetails})
		return ingest.OpengraphMetadata{}, ErrOpengraphMetadataValidation
	} else if err != nil {
		s.reportCriticalError("schema validation returned non validation error for opengraph metadata", err)
		return ingest.OpengraphMetadata{}, err
	} else if err := json.Unmarshal(rawObject, &opengraphMetadata); err != nil {
		s.reportCriticalError("failed unmarshalling json to opengraph Metadata", err)
		return ingest.OpengraphMetadata{}, err
	} else {
		return opengraphMetadata, nil
	}
}

// handleData() is called after the "data" tag is found. Currently this simply checks that the next token
// is an opening array then passes through.
func (s *Validator) handleData() error {
	if err := s.enterArray(); err != nil {
		s.reportCriticalError("failed to enter data array", err)
		return err
	}

	return nil
}

// handleGraph() parses and validates opengraph specific data after the "graph" tag is found at the top level
func (s *Validator) handleGraph() error {
	if err := s.enterObject(); err != nil {
		s.reportCriticalError("failed to enter graph object", err)
		return err
	}

	for {
		if tag, exitedBlock, err := s.nextTagAtDepth(2); err != nil {
			s.reportCriticalError("failed parsing graph child tag", err)
			return err
		} else if exitedBlock {
			return nil
		} else {
			switch tag {
			case "nodes":
				if s.opengraphData.NodesFound {
					s.reportCriticalError("duplicate graph nodes tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}
				s.opengraphData.NodesFound = true

				numItems, err := s.handleOpenGraphArray(tag, s.schema.NodeSchema)
				s.opengraphData.NodesValidated = numItems
				if err != nil {
					return err
				}
			case "edges":
				if s.opengraphData.EdgesFound {
					s.reportCriticalError("duplicate graph edges tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}
				s.opengraphData.EdgesFound = true

				numItems, err := s.handleOpenGraphArray(tag, s.schema.EdgeSchema)
				s.opengraphData.EdgesValidated = numItems
				if err != nil {
					return err
				}
			default:
				s.reportCriticalError(fmt.Sprintf("unrecognized graph child tag: %s", tag), ErrInvalidFileConfiguration)
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
func (s *Validator) handleOpenGraphArray(arrayName string, schema *jsonschema.Schema) (int, error) {
	index := 0

	if err := s.enterArray(); err != nil {
		s.reportCriticalError(fmt.Sprintf("failed to enter graph %s array", arrayName), err)
		return index, err
	}

	for s.decoder.More() {
		var item decodedArrayObject

		if err := s.decoder.Decode(&item); err != nil {
			s.reportCriticalError(fmt.Sprintf("failed to decode %s array object", arrayName), err)
			return index, err
		}

		if err := schema.Validate(item.Object); err != nil {
			var (
				location  = fmt.Sprintf("/graph/%s[%d]", arrayName, index)
				schemaErr *jsonschema.ValidationError
			)

			if ok := errors.As(err, &schemaErr); !ok {
				s.reportCriticalError(fmt.Sprintf("schema validation returned non validation error at /graph/%s[%d]", arrayName, index), err)
				return index, err
			} else if errorDetails, err := extractJsonSchemaErrors(schemaErr); err != nil {
				s.reportCriticalError(fmt.Sprintf("failed extracting json schema errors at /graph/%s[%d]", arrayName, index), err)
				return index, err
			} else {
				s.reportValidationError(ValidationError{
					Location:  location,
					RawObject: item.RawObject,
					Errors:    errorDetails,
				})
			}
		}

		index++

		if s.exceededValidationErrors() {
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

// skipValue consumes one complete JSON value and returns with the same depth
// after a nested object or array has been consumed.
func (s *Validator) skipValue() error {
	initialDepth := s.depth

	token, err := s.nextToken()
	if err != nil {
		return err
	}

	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter || (delimiter != delimOpenObject && delimiter != delimOpenArray) {
		return nil
	}

	for s.depth > initialDepth {
		if _, err := s.nextToken(); err != nil {
			return err
		}
	}

	return nil
}

// enterObject() consumes the next JSON token. Returns an error if the next token is not {
func (s *Validator) enterObject() error {
	t, err := s.nextToken()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); !ok || delim != delimOpenObject {
		return fmt.Errorf("expected open bracket")
	}

	return nil
}

// enterArray() consumes the next JSON token. Returns an error if the next token is not [
func (s *Validator) enterArray() error {
	t, err := s.nextToken()
	if err != nil {
		return err
	}

	if delim, ok := t.(json.Delim); !ok || delim != delimOpenArray {
		return fmt.Errorf("expected open square bracket")
	}

	return nil
}

// nextTagAtDepth() consumes tokens until it finds the next tag at the specified depth, returning that token. If
// nextTagAtDepth exits the specified depth (depth decreases), then the function returns true in the second argument.
func (s *Validator) nextTagAtDepth(depth int) (string, bool, error) {
	for {
		t, err := s.nextToken()
		if err != nil {
			return "", false, err
		}

		if s.depth < depth {
			return "", true, nil
		}

		tag, ok := t.(string)
		if !ok {
			continue
		}

		if s.depth == depth {
			return tag, false, nil
		}
	}
}

// nextToken() consumes the next JSON token, returning it. This function should be the only one used for
// interacting with the underlying JSON file because it keeps track of file depth.
func (s *Validator) nextToken() (json.Token, error) {
	tok, err := s.decoder.Token()
	if err != nil {
		return nil, err
	}

	if d, ok := tok.(json.Delim); ok {
		switch d {
		case delimOpenObject, delimOpenArray:
			s.depth++
		case delimCloseObject, delimCloseArray:
			s.depth--
		}
	}

	return tok, nil
}

// expectEOF() reads the next JSON token and expects to hit the end of the file. Returns an error otherwise
func (s *Validator) expectEOF() error {
	tok, err := s.nextToken()

	if err == io.EOF {
		return nil
	}

	if err != nil {
		return err
	}

	return fmt.Errorf("expected EOF, instead got token: %v", tok)
}
