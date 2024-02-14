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
	"github.com/specterops/bloodhound/log"
	"io"
)

func ValidateMetaTag(reader io.ReadSeeker) (Metadata, error) {
	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return Metadata{}, fmt.Errorf("error seeking to start of file: %w", err)
	}
	var (
		depth            = 0
		decoder          = json.NewDecoder(reader)
		dataTagFound     = false
		dataTagValidated = false
		metaTagFound     = false
		meta             Metadata
	)

	for {
		if dataTagValidated && metaTagFound {
			return meta, nil
		}
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				if !metaTagFound && !dataTagFound {
					return Metadata{}, ErrNoTagFound
				} else if !dataTagFound {
					return Metadata{}, ErrDataTagNotFound
				} else {
					return Metadata{}, ErrMetaTagNotFound
				}
			}
			return Metadata{}, err
		} else {
			//Validate that our data tag is actually opening correctly
			if dataTagFound && !dataTagValidated {
				if typed, ok := token.(json.Delim); ok && typed == delimOpenSquareBracket {
					dataTagValidated = true
				} else {
					dataTagFound = false
				}
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
				if !metaTagFound && depth == 1 && typed == "meta" {
					if err := decoder.Decode(&meta); err != nil {
						log.Warnf("Found invalid metatag, skipping")
					} else if meta.Type.IsValid() {
						metaTagFound = true
					}
				}

				if !dataTagFound && depth == 1 && typed == "data" {
					dataTagFound = true
				}
			}
		}
	}
}
