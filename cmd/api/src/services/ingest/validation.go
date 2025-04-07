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
			}
		}

		if dataTagValidated && metaTagFound {
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

/*
	"graph": {
		"nodes": [],
		"edges": [],
	}

	or

	"graph": {
		"edges": [],
		"nodes": [],
	}
*/

func ValidateNodeSchema() string {

	c := jsonschema.NewCompiler()
	nodeSchema := c.MustCompile("./json_schema/node.json")
	edgeSchema := c.MustCompile("./json_schema/edge.json")

	node, _ := jsonschema.UnmarshalJSON(strings.NewReader(`{"id": "1234","kinds": ["a"]}`))
	err := nodeSchema.Validate(node)
	fmt.Println(err)

	edge, _ := jsonschema.UnmarshalJSON(strings.NewReader(`{"start": 
{"id_value": "234"},
"kind": "a", 
"end": {"id_value": "234"}}`))

	err = edgeSchema.Validate(edge)
	fmt.Println(err)
	return ""
}

func handleValidationError(err error) string {
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

// TODO: a payload can contain just edges or just nodes, or both
func ValidateGenericIngest(reader io.Reader, readToEnd bool) error {

	var (
		decoder    = NewStreamDecoder(reader)
		nodesFound = false
		edgesFound = false
		c          = jsonschema.NewCompiler()
		nodeSchema = c.MustCompile("./json_schema/node.json")
		edgeSchema = c.MustCompile("./json_schema/edge.json")
	)

	var validateNodes = func() error {
		nodesFound = true

		if err := decoder.EatOpeningBracket(); err != nil {
			if err == ingest.ErrNullArray {
				return nil // payload is edge only
			} else {
				return err
			}
		}

		nodeIndex := 0
		// churn through array
		for decoder.More() {
			var node map[string]any
			if err := decoder.DecodeNext(&node); err != nil {
				return err
			}

			// validate against json schema
			if err := nodeSchema.Validate(node); err != nil {
				errorStr := handleValidationError(err)
				// slog.Warn("validation error for node %d: %v", nodeIndex, errorStr)
				return fmt.Errorf("[%d] %w: %s", nodeIndex, ingest.ErrNodeValidation, errorStr)
			}

			nodeIndex++
		}

		if err := decoder.EatClosingBracket(); err != nil {
			return err
		}

		return nil
	}

	var validateEdges = func() error {
		edgesFound = true

		if err := decoder.EatOpeningBracket(); err != nil {
			return err
		}

		edgeIndex := 0
		// churn through array
		for decoder.More() {
			var edge map[string]any
			if err := decoder.DecodeNext(&edge); err != nil {
				return err
			}

			// validate against json schema
			if err := edgeSchema.Validate(edge); err != nil {
				errorStr := handleValidationError(err)
				return fmt.Errorf("[%d] %w: %s", edgeIndex, ingest.ErrEdgeValidation, errorStr)
			}

			edgeIndex++
		}

		if err := decoder.EatClosingBracket(); err != nil {
			return err
		}

		return nil
	}

	// consume JSON until we get to the graph tag
	for {
		if token, err := decoder.dec.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				if !nodesFound && !edgesFound { // if payload empty, reject
					return ingest.ErrEmptyIngest
				}
				// TODO: may need more checking here for closing delims
				break
			}
		} else {
			switch typedToken := token.(type) {
			case string:
				if typedToken == "graph" {
					if err := decoder.EatOpeningCurlyBracket(); err != nil {
						return err
					}
				}

				if typedToken == "nodes" {
					if err := validateNodes(); err != nil {
						return err
					}
				}

				if typedToken == "edges" {
					if err := validateEdges(); err != nil {
						return err
					}
				}
			}
		}
	}

	if readToEnd {
		if _, err := io.Copy(io.Discard, reader); err != nil {
			return err
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
