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

// Error Definitions

var (
	ErrMaxValidationErrors         = errors.New("reached maximum validation errors allowed")
	ErrValidationErrors            = errors.New("validator exited with validation errors")
	ErrInvalidFileConfiguration    = errors.New("invalid file configuration")
	ErrOpengraphMetadataValidation = errors.New("opengraph metadata validation error")
	ErrInvalidDataType             = errors.New("invalid data type")
)

// Validator Definitions

type Validator struct {
	reader  io.Reader
	decoder *json.Decoder
	depth   int

	schema IngestSchema

	legacyData    legacyData
	opengraphData opengraphData

	maxValidationErrors int
	criticalErrors      []CriticalError
	validationErrors    []ValidationError
}

type legacyData struct {
	DataFound     bool
	MetadataFound bool

	Metadata ingest.LegacyMetadata
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

// Return Definitions

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

	LegacyMetadata ingest.LegacyMetadata
	OpengraphData  ParsedOpenGraphData
}

type ParsedOpenGraphData struct {
	Metadata       ingest.OpengraphMetadata
	NodesValidated int
	EdgesValidated int
}

func (v *Validator) buildParsedData() ParsedData {
	p := ParsedData{}

	if (v.opengraphData.GraphFound || v.opengraphData.MetadataFound) && (v.legacyData.MetadataFound || v.legacyData.DataFound) {
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

	if v.legacyData.MetadataFound {
		p.PayloadType = v.legacyData.Metadata.Type
		p.LegacyMetadata = v.legacyData.Metadata
	}

	return p
}

func (v *Validator) buildValidationReport() ValidationReport {
	return ValidationReport{
		CriticalErrors:   v.criticalErrors,
		ValidationErrors: v.validationErrors,
	}
}

// Error Helper functions

func (v *Validator) reportCriticalError(message string, err error) {
	v.criticalErrors = append(v.criticalErrors, CriticalError{Message: message, Error: err})
}

func (v *Validator) reportValidationError(validationErr ValidationError) {
	v.validationErrors = append(v.validationErrors, validationErr)
}

// Validator state check

func (v *Validator) exceededValidationErrors() bool {
	return v.maxValidationErrors != 0 && (len(v.validationErrors) >= v.maxValidationErrors)
}

func (v *Validator) recurringFileConfigCheck() error {
	if v.legacyData.MetadataFound && v.opengraphData.MetadataFound {
		v.reportCriticalError("cannot have both legacy meta tag and opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.legacyData.MetadataFound && v.opengraphData.GraphFound {
		v.reportCriticalError("cannot have both legacy meta tag and opengraph graph tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.legacyData.DataFound && v.opengraphData.MetadataFound {
		v.reportCriticalError("cannot have both legacy data tag and opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.legacyData.DataFound && v.opengraphData.GraphFound {
		v.reportCriticalError("cannot have both legacy data tag and opengraph graph tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	return nil
}

func (v *Validator) finalFileConfigCheck() error {
	if !v.legacyData.MetadataFound && !v.legacyData.DataFound && !v.opengraphData.MetadataFound && !v.opengraphData.GraphFound {
		v.reportCriticalError("no tags found", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.legacyData.MetadataFound && !v.legacyData.DataFound {
		v.reportCriticalError("no data tag found to match legacy metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if !v.legacyData.MetadataFound && v.legacyData.DataFound {
		v.reportCriticalError("no meta tag found to match legacy data tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.opengraphData.MetadataFound && !v.opengraphData.GraphFound {
		v.reportCriticalError("no graph tag found to match opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.opengraphData.GraphFound && !v.opengraphData.NodesFound {
		v.reportCriticalError("graph tag requires child nodes tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	return nil
}

// External call function

func (v *Validator) ParseAndValidate() (ParsedData, ValidationReport, error) {
	if err := v.enterObject(); err != nil {
		v.reportCriticalError("failed to enter json object", err)
		return v.buildParsedData(), v.buildValidationReport(), err
	}

	valLoopErr := v.validationLoop()
	_, readToEndErr := io.Copy(io.Discard, v.reader)
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

// Validation Loop functions

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
				if v.legacyData.MetadataFound {
					v.reportCriticalError("duplicate top level meta tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				v.legacyData.MetadataFound = true

				legacyMetadata, err := v.parseLegacyMetadata()
				if err != nil {
					return err
				}

				v.legacyData.Metadata = legacyMetadata
			case "data":
				if v.legacyData.DataFound {
					v.reportCriticalError("duplicate top level data tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				v.legacyData.DataFound = true

				err := v.parseData()
				if err != nil {
					return err
				}
			case "metadata":
				if v.opengraphData.MetadataFound {
					v.reportCriticalError("duplicate top level metadata tag found", ErrInvalidFileConfiguration)
					return ErrInvalidFileConfiguration
				}

				v.opengraphData.MetadataFound = true

				opengraphMetadata, err := v.parseOpenGraphMetadata()
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

				err := v.parseGraph()
				if err != nil {
					return err
				}
			default:
				v.reportCriticalError("unrecognized top level tag", ErrInvalidFileConfiguration)
				return ErrInvalidFileConfiguration
			}
		}
	}
}

func (v *Validator) parseLegacyMetadata() (ingest.LegacyMetadata, error) {
	var legacyMetadata ingest.LegacyMetadata
	if err := v.decoder.Decode(&legacyMetadata); err != nil {
		v.reportCriticalError("failed to decode legacy metadata", err)
		return ingest.LegacyMetadata{}, err
	}

	if !legacyMetadata.Type.IsValidOriginalType() {
		v.reportCriticalError("invalid legacy metadata data type", ErrInvalidDataType)
		return ingest.LegacyMetadata{}, ErrInvalidDataType
	}

	return legacyMetadata, nil
}

func (v *Validator) parseOpenGraphMetadata() (ingest.OpengraphMetadata, error) {
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

func (v *Validator) parseData() error {
	if err := v.enterArray(); err != nil {
		v.reportCriticalError("failed to enter data array", err)
		return err
	}

	return nil
}

func (v *Validator) parseGraph() error {
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

				numItems, err := v.parseOpenGraphArray(tag, v.schema.NodeSchema)
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

				numItems, err := v.parseOpenGraphArray(tag, v.schema.EdgeSchema)
				v.opengraphData.EdgesValidated = numItems
				if err != nil {
					return err
				}
			}
		}
	}
}

type decodedArrayObject struct {
	RawObject string
	Object    any
}

func (s *decodedArrayObject) UnmarshalJSON(bytes []byte) error {
	s.RawObject = string(bytes)

	return json.Unmarshal(bytes, &s.Object)
}

func (v *Validator) parseOpenGraphArray(arrayName string, schema *jsonschema.Schema) (int, error) {
	index := 0

	if err := v.enterArray(); err != nil {
		v.reportCriticalError(fmt.Sprintf("failed to enter graph %s array", arrayName), err)
		return index, err
	}

	for v.decoder.More() {
		var item decodedArrayObject

		err := v.decoder.Decode(&item)
		if err != nil {
			v.reportCriticalError(fmt.Sprintf("failed to decode %s array object", arrayName), err)
			return index, err
		}

		err = schema.Validate(item.Object)
		if err != nil {
			location := fmt.Sprintf("/graph/%s[%d]", arrayName, index)

			var schemaErr *jsonschema.ValidationError
			if ok := errors.As(err, &schemaErr); !ok {
				v.reportCriticalError(fmt.Sprintf("schema validation returned non validation error at /graph/%s[%d]", arrayName, index), err)
				return index, err
			}

			errorDetails, err := extractJsonSchemaErrors(schemaErr)
			if err != nil {
				v.reportCriticalError(fmt.Sprintf("failed extracting json schema errors at /graph/%s[%d]", arrayName, index), err)
				return index, err
			}

			v.reportValidationError(ValidationError{
				Location:  location,
				RawObject: item.RawObject,
				Errors:    errorDetails,
			})
		}

		index++

		if v.exceededValidationErrors() {
			return index, ErrMaxValidationErrors
		}
	}

	return index, nil
}

func extractJsonSchemaErrors(ve *jsonschema.ValidationError) ([]ValidationErrorDetail, error) {
	errors := make(map[string]string, 0)

	for _, cause := range ve.Causes {
		output := cause.BasicOutput()

		if output == nil {
			return []ValidationErrorDetail{}, fmt.Errorf("failed to extract schema validation error BasicOutput")
		} else if output.Error != nil {
			errors[output.InstanceLocation] = output.Error.String()
		} else if output.Errors != nil {
			for _, e := range output.Errors {
				if e.Error == nil {
					return []ValidationErrorDetail{}, fmt.Errorf("failed to extract output error from output unit errors")
				}

				if strings.HasPrefix(e.InstanceLocation, "/properties/") {
					locSplit := strings.Split(e.InstanceLocation, "/")

					newLocation := fmt.Sprintf("/%s/%s", locSplit[1], locSplit[2])
					if _, found := errors[newLocation]; !found {
						errors[newLocation] = "invalid type"
					}
				} else {
					if _, found := errors[e.InstanceLocation]; !found {
						errors[e.InstanceLocation] = e.Error.String()
					}
				}
			}
		}
	}

	errorDetails := make([]ValidationErrorDetail, 0)
	for loc, err := range errors {
		errorDetails = append(errorDetails, ValidationErrorDetail{
			Location: loc,
			Error:    err,
		})
	}

	return errorDetails, nil
}

// Scanner functions

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
