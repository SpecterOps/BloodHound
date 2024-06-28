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
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

// Normalizer interface defines the methods for accessing normalized content.
// This allows consumers of the package to work with normalized data without
// knowing the specifics of how the normalization was performed.
type Normalizer interface {
	// NormalizedContent returns the normalized (UTF-8) content.
	NormalizedContent() []byte
	// NormalizedFrom returns the original encoding of the content before normalization.
	NormalizedFrom() Encoding
}

// normalizer is the concrete implementation of the Normalizer interface.
type normalizer struct {
	normalizedContent []byte   // The normalized (UTF-8) content
	normalizedFrom    Encoding // The original encoding before normalization
}

// NormalizedContent returns the normalized (UTF-8) content.
func (s normalizer) NormalizedContent() []byte {
	return s.normalizedContent
}

// NormalizedFrom returns the original encoding of the content before normalization.
func (s normalizer) NormalizedFrom() Encoding {
	return s.normalizedFrom
}

// DetectBOMEncoding detects the byte order mark in the given bytes and returns the corresponding Encoding.
// This function is crucial for determining the encoding of incoming data based on its BOM.
func DetectBOMEncoding(data []byte) (Encoding, error) {
	// Check for UTF-8 BOM
	if len(data) >= 3 && bytes.Equal(data[:3], UTF8.Sequence()) {
		return UTF8, nil
	}

	if len(data) >= 2 {
		// Check for UTF-16BE BOM
		if bytes.Equal(data[:2], UTF16BE.Sequence()) {
			return UTF16BE, nil
		}
		// Check for potential UTF-16LE or UTF-32LE BOM
		if bytes.Equal(data[:2], UTF16LE.Sequence()) {
			// We need to differentiate between UTF-16LE and UTF-32LE
			if len(data) >= 8 {
				isLikelyUTF16LE := data[2] != 0 || data[3] != 0 || data[4] != 0 || data[5] != 0
				isLikelyUTF32LE := data[2] == 0 && data[3] == 0 && data[6] == 0 && data[7] == 0
				switch {
				case isLikelyUTF32LE && isLikelyUTF16LE:
					return UTF32LE, nil
				case isLikelyUTF16LE && !isLikelyUTF32LE:
					return UTF16LE, nil
				case isLikelyUTF32LE && !isLikelyUTF16LE:
					return UTF32BE, nil
				default:
					return UTF16LE, nil
				}
			}
			// If we can't determine definitively, default to Unknown
			return Unknown, nil
		}
	}

	// Check for UTF-32BE BOM
	if len(data) >= 4 && bytes.Equal(data[:4], UTF32BE.Sequence()) {
		return UTF32BE, nil
	}

	// If no BOM is detected, return Unknown
	return Unknown, nil
}

// NormalizeToUTF8 converts the input to UTF-8, removing any BOM.
// This function is the main entry point for normalizing data from an io.Reader.
// It's useful when working with streams of data, such as file input.
func NormalizeToUTF8(input io.Reader) (Normalizer, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}

	detectedBOMEncoding, err := DetectBOMEncoding(data)
	if err != nil {
		return nil, err
	}

	return NormalizeBytesToUTF8(data, detectedBOMEncoding)
}

// ErrUnknownEncodingInvalidUTF8 ...
var ErrUnknownEncodingInvalidUTF8 = errors.New("unknown encoding and not a valid UTF-8")

// NormalizeBytesToUTF8 converts the given bytes to UTF-8 based on the specified encoding.
// This function is the core of the normalization process, handling different encodings
// and converting them to UTF-8.
func NormalizeBytesToUTF8(data []byte, enc Encoding) (Normalizer, error) {
	var (
		content []byte
		err     error
	)

	switch enc.String() {
	case UTF8.String():
		content = data[min(len(enc.Sequence()), len(data)):]
	case UTF16BE.String():
		content, err = utf16ToUTF8(data[min(len(enc.Sequence()), len(data)):], true)
	case UTF16LE.String():
		content, err = utf16ToUTF8(data[min(len(enc.Sequence()), len(data)):], false)
	case UTF32BE.String():
		content, err = utf32ToUTF8(data[min(len(enc.Sequence()), len(data)):], true)
	case UTF32LE.String():
		content, err = utf32ToUTF8(data[min(len(enc.Sequence()), len(data)):], false)
	case Unknown.String():
		if utf8.Valid(data) {
			content = data
		} else {
			return nil, ErrUnknownEncodingInvalidUTF8
		}
	}

	if err != nil {
		return nil, err
	}

	return normalizer{
		normalizedContent: content,
		normalizedFrom:    enc,
	}, nil
}
