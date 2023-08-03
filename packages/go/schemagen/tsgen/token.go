// Copyright 2023 Specter Ops, Inc.
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

package tsgen

import (
	"fmt"
	"io"
)

// tokenType describes a token's type.
type tokenType string

const (
	tokenTypeNull       tokenType = "null"
	tokenTypeOperator   tokenType = "operator"
	tokenTypeDelimiter  tokenType = "delimiter"
	tokenTypeFormatting tokenType = "formatting"
	tokenTypeIdentifier tokenType = "identifier"
	tokenTypeQualifier  tokenType = "qualifier"
	tokenTypeLiteral    tokenType = "literal"
	tokenTypeKeyword    tokenType = "keyword"
	tokenTypeAccessor   tokenType = "accessor"
)

// FormatLiteral takes an untyped interface{} value and attempts to return the correct TypeScript literal representation
// for it.
func FormatLiteral(value any) (string, error) {
	switch value.(type) {
	case bool, int, float64, float32, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		// default constant types can be left bare
		return fmt.Sprintf("%#v", value), nil

	case string:
		return fmt.Sprintf("'%s'", value), nil

	default:
		return "", fmt.Errorf("unsupported type for literal: %T", value)
	}
}

// token is a literal syntax token that has a token.Type and a token.Value that is type-negotiated upon rendering.
type token struct {
	Type  tokenType
	Value any
}

func (s token) Render(writer io.Writer) error {
	switch s.Type {
	case tokenTypeLiteral:
		if formatted, err := FormatLiteral(s.Value); err != nil {
			return err
		} else if _, err := writer.Write([]byte(formatted)); err != nil {
			return err
		}

	case tokenTypeKeyword, tokenTypeOperator, tokenTypeFormatting, tokenTypeDelimiter:
		if _, err := writer.Write([]byte(fmt.Sprintf("%s", s.Value))); err != nil {
			return err
		}

	case tokenTypeIdentifier, tokenTypeQualifier:
		if strValue, typeOK := s.Value.(string); !typeOK {
			return fmt.Errorf("expected string type as identifier value but got %T(%v)", s.Value, s.Value)
		} else if _, err := writer.Write([]byte(strValue)); err != nil {
			return err
		}
	}

	return nil
}

func (s token) IsNull() bool {
	return s.Type == tokenTypeNull
}
