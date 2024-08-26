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

package bomenc

import (
	"bufio"
	"errors"
	"io"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding/unicode/utf32"
)

// DetectBOMEncoding detects the byte order mark in the given bytes and returns the corresponding Encoding.
// This function is crucial for determining the encoding of incoming data based on its BOM.
func DetectBOMEncoding(data Peeker) Encoding {
	if UTF8.HasSequence(data) {
		return UTF8
	} else if UTF16BE.HasSequence(data) {
		return UTF16BE
	} else if UTF16LE.HasSequence(data) { // NOTE: The byte sequence for UTF16LE matches the first two bytes for UTF32LE; this check's implementation ensures that the BOM does not accidently match UTF32LE
		return UTF16LE
	} else if UTF32LE.HasSequence(data) {
		return UTF32LE
	} else if UTF32BE.HasSequence(data) {
		return UTF32BE
	} else {
		return Unknown
	}
}

// NormalizeToUTF8 converts the input to UTF-8, removing any BOM.
// This function is the main entry point for normalizing data from an io.Reader.
// It's useful when working with streams of data, such as file input.
func NormalizeToUTF8(input io.Reader) (io.Reader, error) {
	buf := bufio.NewReader(input)
	switch DetectBOMEncoding(buf).String() {
	case UTF8.String():
		if _, err := buf.Discard(3); err != nil {
			return nil, err
		} else {
			return buf, nil
		}
	case UTF16LE.String():
		return transformUTF16toUTF8(unicode.LittleEndian, buf)
	case UTF16BE.String():
		return transformUTF16toUTF8(unicode.BigEndian, buf)
	case UTF32LE.String():
		return transformUTF32toUTF8(utf32.LittleEndian, buf)
	case UTF32BE.String():
		return transformUTF32toUTF8(utf32.BigEndian, buf)
	default:
		// Either Unknown or no BOM
		return unicode.UTF8.NewDecoder().Reader(buf), nil
	}
}

// transformUTF16toUTF8 aids NormalizeToUTF8 in converting UTF16 data with a given endienness into UTF8
func transformUTF16toUTF8(endianness unicode.Endianness, data io.Reader) (io.Reader, error) {
	utf16leDecoder := unicode.UTF16(endianness, unicode.UseBOM).NewDecoder()
	return utf16leDecoder.Reader(data), nil
}

// transformUTF32toUTF8 aids NormalizeToUTF8 in converting UTF32 data with a given endienness into UTF8
func transformUTF32toUTF8(endianness utf32.Endianness, data io.Reader) (io.Reader, error) {
	utf16leDecoder := utf32.UTF32(endianness, utf32.UseBOM).NewDecoder()
	return utf16leDecoder.Reader(data), nil
}

// ErrUnknownEncodingInvalidUTF8 ...
var ErrUnknownEncodingInvalidUTF8 = errors.New("unknown encoding and not a valid UTF-8")
