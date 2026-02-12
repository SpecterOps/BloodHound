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
	ErrExceededMaxValidationErrors = errors.New("exceeded maximum allowable validation errors")
	ErrInvalidFileConfiguration    = errors.New("invalid file configuration")
	ErrOpengraphMetadataValidation = errors.New("opengraph metadata validation error")
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

func (v *Validator) buildParsedData() (ParsedData, error) {
	p := ParsedData{}

	if v.opengraphData.GraphFound {
		p.PayloadType = ingest.DataTypeOpenGraph
		p.OpengraphData.Metadata = v.opengraphData.Metadata
		p.OpengraphData.NodesValidated = v.opengraphData.NodesValidated
		p.OpengraphData.EdgesValidated = v.opengraphData.EdgesValidated
		return p, nil
	}

	if v.legacyData.MetadataFound && v.legacyData.DataFound {
		p.PayloadType = v.legacyData.Metadata.Type
		p.LegacyMetadata = v.legacyData.Metadata
		return p, nil
	}

	return ParsedData{}, fmt.Errorf("insufficient data to build ParsedData")
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
		v.reportCriticalError("no metadata tag found to match legacy data tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.opengraphData.MetadataFound && !v.opengraphData.GraphFound {
		v.reportCriticalError("no graph tag found to match opengraph metadata tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	if v.opengraphData.MetadataFound && !v.opengraphData.NodesFound {
		v.reportCriticalError("graph tag requires child nodes tag", ErrInvalidFileConfiguration)
		return ErrInvalidFileConfiguration
	}

	return nil
}

// External call function

func (v *Validator) ParseAndValidate() (ParsedData, ValidationReport, error) {
	if err := v.enterObject(); err != nil {
		v.reportCriticalError("failed to enter json object", err)
		return ParsedData{}, v.buildValidationReport(), err
	}

	valLoopErr := v.validationLoop()
	_, readToEndErr := io.Copy(io.Discard, v.reader)
	if valLoopErr != nil && readToEndErr != nil {
		v.reportCriticalError("failed to read file to end", readToEndErr)
		return ParsedData{}, v.buildValidationReport(), errors.Join(valLoopErr, readToEndErr)
	} else if valLoopErr == nil && readToEndErr != nil {
		v.reportCriticalError("failed to read file to end", readToEndErr)
		return ParsedData{}, v.buildValidationReport(), readToEndErr
	} else if valLoopErr != nil {
		return ParsedData{}, v.buildValidationReport(), valLoopErr
	}

	if err := v.finalFileConfigCheck(); err != nil {
		return ParsedData{}, v.buildValidationReport(), err
	}

	parsedData, err := v.buildParsedData()
	if err != nil {
		return ParsedData{}, v.buildValidationReport(), err
	}

	return parsedData, v.buildValidationReport(), nil
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
				v.legacyData.MetadataFound = true

				legacyMetadata, err := v.parseLegacyMetadata()
				if err != nil {
					return err
				}

				v.legacyData.Metadata = legacyMetadata
			case "data":
				v.legacyData.DataFound = true

				err := v.parseData()
				if err != nil {
					return err
				}
			case "metadata":
				v.opengraphData.MetadataFound = true

				opengraphMetadata, err := v.parseOpenGraphMetadata()
				if err != nil {
					return err
				}

				v.opengraphData.Metadata = opengraphMetadata
			case "graph":
				v.opengraphData.GraphFound = true

				err := v.parseGraph()
				if err != nil {
					return err
				}
			}
		}
	}
}

func (v *Validator) parseLegacyMetadata() (ingest.LegacyMetadata, error) {
	var legacyMetadata ingest.LegacyMetadata
	if err := v.decoder.Decode(&legacyMetadata); err != nil {
		v.reportCriticalError("failed to decode legacy metadata", err)
		return legacyMetadata, err
	}

	return ingest.LegacyMetadata{}, nil
}

func (v *Validator) parseOpenGraphMetadata() (ingest.OpengraphMetadata, error) {
	var rawObject json.RawMessage
	if err := v.schema.MetaSchema.Validate(rawObject); err != nil {
		var schemaErr *jsonschema.ValidationError
		if ok := errors.As(err, &schemaErr); !ok {
			v.reportCriticalError("schema validation returned non validation error for opengraph metadata", err)
			return ingest.OpengraphMetadata{}, err
		}

		v.reportCriticalError("opengraph metadata failed validation", ErrOpengraphMetadataValidation)
		v.reportValidationError(ValidationError{Location: "/metadata", RawObject: string(rawObject), Errors: extractJsonSchemaErrors(schemaErr)})
		return ingest.OpengraphMetadata{}, ErrOpengraphMetadataValidation
	}

	var opengraphMetadata ingest.OpengraphMetadata
	if err := json.Unmarshal(rawObject, &opengraphMetadata); err != nil {
		v.reportCriticalError("failed to decode opengraph metadata", err)
		return opengraphMetadata, err
	}

	return opengraphMetadata, nil
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
				v.opengraphData.NodesFound = true

				numItems, err := v.parseOpenGraphArray(tag, v.schema.NodeSchema)
				v.opengraphData.NodesValidated = numItems
				return err
			case "edges":
				v.opengraphData.EdgesFound = true

				numItems, err := v.parseOpenGraphArray(tag, v.schema.EdgeSchema)
				v.opengraphData.EdgesValidated = numItems
				return err
			}
		}
	}
}

func (v *Validator) parseOpenGraphArray(arrayName string, schema *jsonschema.Schema) (int, error) {
	index := 0

	if err := v.enterArray(); err != nil {
		v.reportCriticalError(fmt.Sprintf("failed to enter graph %s array", arrayName), err)
		return index, err
	}

	for v.decoder.More() {
		var rawItem json.RawMessage

		err := v.decoder.Decode(&rawItem)
		if err != nil {
			v.reportCriticalError(fmt.Sprintf("failed to decode raw %s array object", arrayName), err)
			return index, err
		}

		var item map[string]any

		err = json.Unmarshal(rawItem, &item)
		if err != nil {
			v.reportCriticalError(fmt.Sprintf("failed to unmarshal %s array object", arrayName), err)
			return index, err
		}

		err = schema.Validate(item)
		if err != nil {
			location := fmt.Sprintf("/graph/%s[%d]", arrayName, index)

			var schemaErr *jsonschema.ValidationError
			if ok := errors.As(err, &schemaErr); !ok {
				v.reportCriticalError(fmt.Sprintf("schema validation returned non validation error at /graph/%s[%d]", arrayName, index), err)
				return index, err
			}

			v.reportValidationError(ValidationError{
				Location:  location,
				RawObject: string(rawItem),
				Errors:    extractJsonSchemaErrors(schemaErr),
			})
		}

		if v.exceededValidationErrors() {
			return index, ErrExceededMaxValidationErrors
		}

		index++
	}

	return index, nil
}

func extractJsonSchemaErrors(ve *jsonschema.ValidationError) []ValidationErrorDetail {
	errors := make(map[string]string, 0)

	for _, cause := range ve.Causes {
		if cause.BasicOutput() == nil {
			continue
		}

		output := cause.BasicOutput()

		if output.Errors != nil {
			for _, e := range output.Errors {
				if e.Error == nil {
					continue
				}

				locSplit := strings.Split(e.InstanceLocation, "/")
				if len(locSplit) < 3 {
					continue
				}

				if locSplit[1] == "properties" {
					newLocation := fmt.Sprintf("/%s/%s", locSplit[1], locSplit[2])
					if _, ok := errors[newLocation]; !ok {
						errors[newLocation] = e.Error.String()
					}
				} else {
					if _, ok := errors[e.InstanceLocation]; !ok {
						errors[e.InstanceLocation] = e.Error.String()
					}
				}
			}
		} else if output.Error != nil {
			errors[output.InstanceLocation] = output.Error.String()
		}
	}

	errorDetails := make([]ValidationErrorDetail, 0)
	for loc, err := range errors {
		errorDetails = append(errorDetails, ValidationErrorDetail{
			Location: loc,
			Error:    err,
		})
	}

	return errorDetails
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
