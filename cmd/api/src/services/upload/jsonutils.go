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
	"encoding/json"
	"fmt"

	"github.com/specterops/bloodhound/src/model/ingest"
)

func expectOpenArray(decoder *json.Decoder, name string) error {
	if tok, err := decoder.Token(); err != nil {
		return fmt.Errorf("error decoding %s array: %w", name, err)
	} else if delim, ok := tok.(json.Delim); !ok || delim != ingest.DelimOpenSquareBracket {
		return fmt.Errorf("error opening %s array: expected '[', got %v", name, tok)
	}
	return nil
}

func expectClosingArray(decoder *json.Decoder, name string) error {
	if tok, err := decoder.Token(); err != nil {
		return fmt.Errorf("error decoding %s array: %w", name, err)
	} else if delim, ok := tok.(json.Delim); !ok || delim != ingest.DelimCloseSquareBracket {
		return fmt.Errorf("error opening %s array: expected ']', got %v", name, tok)
	}
	return nil
}

func expectOpenObject(decoder *json.Decoder, name string) error {
	if tok, err := decoder.Token(); err != nil {
		return fmt.Errorf("error decoding %s object: %w", name, err)
	} else if delim, ok := tok.(json.Delim); !ok || delim != ingest.DelimOpenBracket {
		return fmt.Errorf("error opening %s object: expected '{', got %v", name, tok)
	}
	return nil
}

func expectClosingObject(decoder *json.Decoder, name string) error {
	if tok, err := decoder.Token(); err != nil {
		return fmt.Errorf("error decoding %s object: %w", name, err)
	} else if delim, ok := tok.(json.Delim); !ok || delim != ingest.DelimCloseBracket {
		return fmt.Errorf("error closing %s object: expected '}', got %v", name, tok)
	}
	return nil
}
