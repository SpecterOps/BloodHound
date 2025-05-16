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

	"github.com/specterops/bloodhound/src/model/ingest"
)

func SeekToKey(decoder *json.Decoder, key string, targetDepth int) error {
	var (
		depth    = 0
		keyFound = false
	)

	for {
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				return ingest.ErrDataTagNotFound
			}

			return fmt.Errorf("%w: %w", ingest.ErrJSONDecoderInternal, err)
		} else {
			//Break here to allow for one more token read, which should take us to the "[" token, exactly where we need to be
			if keyFound {
				//Do some extra checks
				if typed, ok := token.(json.Delim); !ok {
					return ingest.ErrInvalidDataTag
				} else if typed != ingest.DelimOpenSquareBracket {
					return ingest.ErrInvalidDataTag
				}
				//Break out of our loop if we're in a good spot
				return nil
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
				if !keyFound && depth == targetDepth && typed == key {
					keyFound = true
				}
			}
		}
	}
}

// CreateIngestDecoder returns a JSON decoder that is positioned at the start of the array
// under the specified top-level key (e.g., "nodes", "edges", "data").
// The returned decoder is ready to stream-decode each element of the array sequentially.
func CreateIngestDecoder(reader io.ReadSeeker, key string, targetDepth int) (*json.Decoder, error) {
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("error seeking to start of file: %w", err)
	} else {
		decoder := json.NewDecoder(reader)
		if err := SeekToKey(decoder, key, targetDepth); err != nil {
			return nil, fmt.Errorf("error seeking to %s tag: %w", key, err)
		} else {
			return decoder, nil
		}
	}
}
