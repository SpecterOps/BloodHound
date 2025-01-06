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
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/text/encoding/unicode"
)

func TestDetectBOMEncoding(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Encoding
	}{
		{
			name:     "UTF-8 BOM",
			input:    []byte{0xEF, 0xBB, 0xBF, 0x68, 0x65, 0x6C, 0x6C, 0x6F},
			expected: UTF8,
		},
		{
			name:     "UTF-16BE BOM",
			input:    []byte{0xFE, 0xFF, 0x00, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F},
			expected: UTF16BE,
		},
		{
			name:     "UTF-16LE BOM",
			input:    []byte{0xFF, 0xFE, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F, 0x00},
			expected: UTF16LE,
		},
		{
			name:     "UTF-32BE BOM",
			input:    []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65},
			expected: UTF32BE,
		},
		{
			name:     "UTF-32LE BOM",
			input:    []byte{0xFF, 0xFE, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00},
			expected: UTF32LE,
		},
		{
			name:     "No BOM",
			input:    []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F},
			expected: Unknown,
		},
		{
			name:     "Empty input",
			input:    []byte{},
			expected: Unknown,
		},
		{
			name:     "Incomplete UTF-16LE BOM (should not be detected as UTF-16LE)",
			input:    []byte{0xFF, 0xFE, 0x68},
			expected: Unknown,
		},
		{
			name:     "Incomplete UTF-32LE BOM (should not be detected as UTF-32LE)",
			input:    []byte{0xFF, 0xFE, 0x00},
			expected: Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewReader(tt.input))
			result := DetectBOMEncoding(reader)
			assert.Equal(t, tt.expected.String(), result.String(), "DetectBOMEncoding() should return the correct encoding")
		})
	}
}

func TestNormalizeToUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
		encFrom  Encoding
		wantErr  bool
	}{
		{
			name:     "UTF-8 BOM",
			input:    []byte{0xEF, 0xBB, 0xBF, 0x68, 0x65, 0x6C, 0x6C, 0x6F},
			expected: []byte("hello"),
			encFrom:  UTF8,
			wantErr:  false,
		},
		{
			name:     "UTF-16BE BOM",
			input:    []byte{0xFE, 0xFF, 0x00, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F},
			expected: []byte("hello"),
			encFrom:  UTF16BE,
			wantErr:  false,
		},
		{
			name:     "UTF-16LE BOM",
			input:    []byte{0xFF, 0xFE, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F, 0x00},
			expected: []byte("hello"),
			encFrom:  UTF16LE,
			wantErr:  false,
		},
		{
			name:     "UTF-32BE BOM",
			input:    []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6F},
			expected: []byte("hello"),
			encFrom:  UTF32BE,
			wantErr:  false,
		},
		{
			name:     "UTF-32LE BOM",
			input:    []byte{0xFF, 0xFE, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6F, 0x00, 0x00, 0x00},
			expected: []byte("hello"),
			encFrom:  UTF32LE,
			wantErr:  false,
		},
		{
			name:     "No BOM (valid UTF-8)",
			input:    []byte("hello"),
			expected: []byte("hello"),
			encFrom:  Unknown,
			wantErr:  false,
		},
		{
			name:     "No BOM (invalid UTF-8)",
			input:    []byte{0xFF, 0xFE, 0xFD},
			expected: []byte{0xEF, 0xBF, 0xBD, 0xEF, 0xBF, 0xBD, 0xEF, 0xBF, 0xBD},
			encFrom:  Unknown,
			wantErr:  false,
		},
		{
			name:     "Empty input",
			input:    []byte{},
			expected: []byte{},
			encFrom:  Unknown,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewReader(tt.input))
			detectedEnc := DetectBOMEncoding(reader)
			result, err := NormalizeToUTF8(reader)
			assert.NoError(t, err)
			actual, err := io.ReadAll(result)

			if tt.wantErr {
				assert.Error(t, err, "NormalizeToUTF8() should return an error for invalid input")
				return
			}

			require.NoError(t, err, "NormalizeToUTF8() should not return an error for valid input")
			assert.Equal(t, tt.expected, actual, "NormalizedContent() should return the correct normalized content")
			assert.Equal(t, tt.encFrom.String(), detectedEnc.String(), "NormalizedFrom() should return the correct original encoding")
		})
	}
}

// Mock reader for testing error cases
type errorReader struct{}

func (er errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}

func TestNormalizeToUTF8_ReaderError(t *testing.T) {
	reader, err := NormalizeToUTF8(errorReader{})
	assert.NoError(t, err)
	_, err = io.ReadAll(reader)
	assert.Error(t, err, "NormalizeToUTF8() should return an error when the reader fails")
}

func TestNormalizeToUTF8_LargeInput(t *testing.T) {
	type testCase struct {
		name     string
		input    []byte
		expected []byte
		encFrom  Encoding
	}

	// Generate a large data set with 1000 Unicode code points
	// NOTE: We don't want to begin the payload with a NULL character because it would then be impossible to discern between UTF16LE and UTF32LE
	var runes []rune
	for i := 0; i <= 1000; i++ {
		runes = append(runes, rune(i%0x10FFFF))
	}

	utf8Bytes := []byte(string(runes))
	utf16LEBytes, err := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewEncoder().Bytes(utf8Bytes)
	require.NoError(t, err)

	tests := []testCase{
		{
			name:     "Large UTF-16LE input",
			input:    utf16LEBytes,
			expected: utf8Bytes,
			encFrom:  UTF16LE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(bytes.NewReader(tt.input))
			detectedEnc := DetectBOMEncoding(reader)
			assert.Equal(t, tt.encFrom.String(), detectedEnc.String(), "NormalizedFrom() should return the correct original encoding")

			result, err := NormalizeToUTF8(reader)
			if err != nil {
				t.Errorf("NormalizeToUTF8() error = %v", err)
				// Print the first few bytes of the input for debugging
				t.Logf("First 20 bytes of input: %v", tt.input[:20])
				// Print the detected encoding
				assert.NoError(t, err)
				t.Logf("Detected encoding: %v", detectedEnc)
				return
			}

			actual, err := io.ReadAll(result)
			assert.NoError(t, err)

			if !bytes.Equal(tt.expected, actual) {
				t.Errorf("NormalizedContent() = %v, want %v", actual, tt.expected)
				// Print the first few bytes of the result and expected for debugging
				t.Logf("First 20 bytes of result: %v", actual[:20])
				t.Logf("First 20 bytes of expected: %v", tt.expected[:20])
				t.Logf("First 20 bytes of input: %v", tt.input[:20])
			}
		})
	}
}
