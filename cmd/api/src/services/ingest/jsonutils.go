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

package ingest

import (
	"encoding/json"
	"fmt"
)

func expectOpeningSquareBracket(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding %s array: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '[' {
		return fmt.Errorf("error opening %s array: expected '[', got %v", name, tok)
	}
	return nil
}

func expectClosingSquareBracket(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding %s array: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != ']' {
		return fmt.Errorf("error closing %s array: expected ']', got %v", name, tok)
	}
	return nil
}

func expectOpeningCurlyBracket(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding %s object: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '{' {
		return fmt.Errorf("error opening %s object: expected '{', got %v", name, tok)
	}
	return nil
}

func expectClosingCurlyBracket(decoder *json.Decoder, name string) error {
	tok, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("error decoding %s object: %w", name, err)
	}
	delim, ok := tok.(json.Delim)
	if !ok || delim != '}' {
		return fmt.Errorf("error closing %s object: expected '}', got %v", name, tok)
	}
	return nil
}
