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
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			result, err := DetectBOMEncoding(tt.input)
			require.NoError(t, err)
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
			expected: nil,
			encFrom:  Unknown,
			wantErr:  true,
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
			reader := bytes.NewReader(tt.input)
			result, err := NormalizeToUTF8(reader)

			if tt.wantErr {
				assert.Error(t, err, "NormalizeToUTF8() should return an error for invalid input")
				return
			}

			require.NoError(t, err, "NormalizeToUTF8() should not return an error for valid input")
			assert.Equal(t, tt.expected, result.NormalizedContent(), "NormalizedContent() should return the correct normalized content")
			assert.Equal(t, tt.encFrom.String(), result.NormalizedFrom().String(), "NormalizedFrom() should return the correct original encoding")
		})
	}
}

func TestNormalizeBytesToUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		enc      Encoding
		expected []byte
		wantErr  bool
	}{
		{
			name:     "UTF-8 BOM",
			input:    []byte{0xEF, 0xBB, 0xBF, 0x68, 0x65, 0x6C, 0x6C, 0x6F},
			enc:      UTF8,
			expected: []byte("hello"),
			wantErr:  false,
		},
		{
			name:     "UTF-16BE BOM",
			input:    []byte{0xFE, 0xFF, 0x00, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F},
			enc:      UTF16BE,
			expected: []byte("hello"),
			wantErr:  false,
		},
		{
			name:     "UTF-16LE BOM",
			input:    []byte{0xFF, 0xFE, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F, 0x00},
			enc:      UTF16LE,
			expected: []byte("hello"),
			wantErr:  false,
		},
		{
			name:     "UTF-32BE BOM",
			input:    []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6F},
			enc:      UTF32BE,
			expected: []byte("hello"),
			wantErr:  false,
		},
		{
			name:     "UTF-32LE BOM",
			input:    []byte{0xFF, 0xFE, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6F, 0x00, 0x00, 0x00},
			enc:      UTF32LE,
			expected: []byte("hello"),
			wantErr:  false,
		},
		{
			name:     "No BOM (valid UTF-8)",
			input:    []byte("hello"),
			enc:      Unknown,
			expected: []byte("hello"),
			wantErr:  false,
		},
		{
			name:     "No BOM (invalid UTF-8)",
			input:    []byte{0xFF, 0xFE, 0xFD},
			enc:      Unknown,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Empty input",
			input:    []byte{},
			enc:      Unknown,
			expected: []byte{},
			wantErr:  false,
		},
		{
			name:     "Invalid UTF-16",
			input:    []byte{0xFE, 0xFF, 0x00},
			enc:      UTF16BE,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid UTF-32",
			input:    []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00},
			enc:      UTF32BE,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NormalizeBytesToUTF8(tt.input, tt.enc)

			if tt.wantErr {
				assert.Error(t, err, "NormalizeBytesToUTF8() should return an error for invalid input")
				return
			}

			require.NoError(t, err, "NormalizeBytesToUTF8() should not return an error for valid input")
			assert.Equal(t, tt.expected, result.NormalizedContent(), "NormalizedContent() should return the correct normalized content")
			assert.Equal(t, tt.enc.String(), result.NormalizedFrom().String(), "NormalizedFrom() should return the correct original encoding")
		})
	}
}

// Mock reader for testing error cases
type errorReader struct{}

func (er errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("mock read error")
}

func TestNormalizeToUTF8_ReaderError(t *testing.T) {
	_, err := NormalizeToUTF8(errorReader{})
	assert.Error(t, err, "NormalizeToUTF8() should return an error when the reader fails")
}

func TestNormalizeToUTF8_LargeInput(t *testing.T) {
	type testCase struct {
		name     string
		input    []byte
		expected []byte
		encFrom  Encoding
	}

	// Generate a large input with 1000 Unicode code points
	var utf16LE, expected []byte

	for i := 0; i < 1000; i++ {
		r := rune(i % 0x10FFFF) // Use all possible Unicode code points

		// UTF-16
		u16 := utf16.Encode([]rune{r})
		for _, c := range u16 {
			utf16LE = append(utf16LE, byte(c), byte(c>>8))
		}

		// Expected UTF-8
		buf := make([]byte, 4)
		n := utf8.EncodeRune(buf, r)
		expected = append(expected, buf[:n]...)
	}

	// Add BOM
	utf16LE = append([]byte{0xFF, 0xFE}, utf16LE...)

	tests := []testCase{
		{
			name:     "Large UTF-16LE input",
			input:    utf16LE,
			expected: expected,
			encFrom:  UTF16LE,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.input)
			result, err := NormalizeToUTF8(reader)
			if err != nil {
				t.Errorf("NormalizeToUTF8() error = %v", err)
				// Print the first few bytes of the input for debugging
				t.Logf("First 20 bytes of input: %v", tt.input[:20])
				// Print the detected encoding
				detectedEnc, err := DetectBOMEncoding(tt.input)
				assert.NoError(t, err)
				t.Logf("Detected encoding: %v", detectedEnc)
				return
			}

			assert.Equal(t, tt.encFrom.String(), result.NormalizedFrom().String(), "NormalizedFrom() should return the correct original encoding")

			if !bytes.Equal(tt.expected, result.NormalizedContent()) {
				t.Errorf("NormalizedContent() = %v, want %v", result.NormalizedContent(), tt.expected)
				// Print the first few bytes of the result and expected for debugging
				t.Logf("First 20 bytes of result: %v", result.NormalizedContent()[:20])
				t.Logf("First 20 bytes of expected: %v", tt.expected[:20])

			}
		})
	}
}
