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

package datapipe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const (
	delimOpenBracket        = json.Delim('{')
	delimCloseBracket       = json.Delim('}')
	delimOpenSquareBracket  = json.Delim('[')
	delimCloseSquareBracket = json.Delim(']')
)

func SeekToDataTag(decoder *json.Decoder) error {
	var (
		depth        = 0
		dataTagFound = false
	)

	for {
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				return ErrDataTagNotFound
			}

			return fmt.Errorf("%w: %w", ErrJSONDecoderInternal, err)
		} else {
			//Break here to allow for one more token read, which should take us to the "[" token, exactly where we need to be
			if dataTagFound {
				//Do some extra checks
				if typed, ok := token.(json.Delim); !ok {
					return ErrInvalidDataTag
				} else if typed != delimOpenSquareBracket {
					return ErrInvalidDataTag
				}
				//Break out of our loop if we're in a good spot
				return nil
			}
			switch typed := token.(type) {
			case json.Delim:
				switch typed {
				case delimCloseBracket, delimCloseSquareBracket:
					depth--
				case delimOpenBracket, delimOpenSquareBracket:
					depth++
				}
			case string:
				if !dataTagFound && depth == 1 && typed == "data" {
					dataTagFound = true
				}
			}
		}
	}
}

func CreateIngestDecoder(reader io.ReadSeeker) (*json.Decoder, error) {
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("error seeking to start of file: %w", err)
	} else {
		decoder := json.NewDecoder(reader)
		if err := SeekToDataTag(decoder); err != nil {
			return nil, fmt.Errorf("error seeking to data tag: %w", err)
		} else {
			return decoder, nil
		}
	}
}
