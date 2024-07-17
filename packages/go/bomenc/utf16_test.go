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
	"testing"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUTF16ToUTF8(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		bigEndian bool
		expected  []byte
		wantErr   bool
	}{
		{
			name:      "UTF-16BE Basic ASCII",
			input:     []byte{0x00, 0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F},
			bigEndian: true,
			expected:  []byte("hello"),
			wantErr:   false,
		},
		{
			name:      "UTF-16LE Basic ASCII",
			input:     []byte{0x68, 0x00, 0x65, 0x00, 0x6C, 0x00, 0x6C, 0x00, 0x6F, 0x00},
			bigEndian: false,
			expected:  []byte("hello"),
			wantErr:   false,
		},
		{
			name:      "UTF-16BE with BMP characters",
			input:     []byte{0x00, 0x41, 0x26, 0x3A, 0x00, 0x42},
			bigEndian: true,
			expected:  []byte("A☺B"),
			wantErr:   false,
		},
		{
			name:      "UTF-16LE with BMP characters",
			input:     []byte{0x41, 0x00, 0x3A, 0x26, 0x42, 0x00},
			bigEndian: false,
			expected:  []byte("A☺B"),
			wantErr:   false,
		},
		{
			name:      "UTF-16BE with surrogate pair",
			input:     []byte{0xD8, 0x3D, 0xDE, 0x00},
			bigEndian: true,
			expected:  []byte{0xF0, 0x9F, 0x98, 0x80}, // U+1F600 - 😀 GRINNING FACE emoji
			wantErr:   false,
		},
		{
			name:      "UTF-16LE with surrogate pair (GRINNING FACE emoji)",
			input:     []byte{0x3D, 0xD8, 0x00, 0xDE},
			bigEndian: false,
			expected:  []byte{0xF0, 0x9F, 0x98, 0x80}, // U+1F600 - 😀 GRINNING FACE emoji
			wantErr:   false,
		},
		{
			name:      "UTF-16BE with mixed characters",
			input:     []byte{0x00, 0x48, 0x00, 0x69, 0xD8, 0x3D, 0xDE, 0x00, 0x00, 0x21},
			bigEndian: true,
			expected:  []byte{0x48, 0x69, 0xF0, 0x9F, 0x98, 0x80, 0x21}, // "Hi😀!" (GRINNING FACE emoji)
			wantErr:   false,
		},
		{
			name:      "UTF-16LE with mixed characters (GRINNING FACE emoji)",
			input:     []byte{0x48, 0x00, 0x69, 0x00, 0x3D, 0xD8, 0x00, 0xDE, 0x21, 0x00},
			bigEndian: false,
			expected:  []byte{0x48, 0x69, 0xF0, 0x9F, 0x98, 0x80, 0x21}, // "Hi😀!" (GRINNING FACE emoji)
			wantErr:   false,
		},
		{
			name:      "Incomplete UTF-16BE sequence",
			input:     []byte{0x00, 0x68, 0x00},
			bigEndian: true,
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Incomplete UTF-16LE sequence",
			input:     []byte{0x68, 0x00, 0x65},
			bigEndian: false,
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Empty input",
			input:     []byte{},
			bigEndian: true,
			expected:  nil,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utf16ToUTF8(tt.input, tt.bigEndian)

			if tt.wantErr {
				assert.Error(t, err, "utf16ToUTF8() should return an error for invalid input")
				return
			}

			require.NoError(t, err, "utf16ToUTF8() should not return an error for valid input")
			assert.Equal(t, tt.expected, result, "utf16ToUTF8() should return the correct UTF-8 bytes")
		})
	}
}

func TestUTF16ToUTF8_LargeInput(t *testing.T) {
	// Generate a large input with 1000 Unicode code points
	var largeInputBE, largeInputLE []byte
	var expected []byte

	for i := 0; i < 1000; i++ {
		r := rune(i % 0x10FFFF) // Use all possible Unicode code points

		utf16Sequence := utf16.Encode([]rune{r})
		for _, u16 := range utf16Sequence {
			// Big Endian
			largeInputBE = append(largeInputBE, byte(u16>>8), byte(u16))
			// Little Endian
			largeInputLE = append(largeInputLE, byte(u16), byte(u16>>8))
		}

		// Append UTF-8 encoded rune to expected result
		buf := make([]byte, 4)
		n := utf8.EncodeRune(buf, r)
		expected = append(expected, buf[:n]...)
	}

	t.Run("Large UTF-16BE input", func(t *testing.T) {
		result, err := utf16ToUTF8(largeInputBE, true)
		require.NoError(t, err, "utf16ToUTF8() should not return an error for valid large input")
		assert.Equal(t, expected, result, "utf16ToUTF8() should correctly convert large UTF-16BE input")
	})

	t.Run("Large UTF-16LE input", func(t *testing.T) {
		result, err := utf16ToUTF8(largeInputLE, false)
		require.NoError(t, err, "utf16ToUTF8() should not return an error for valid large input")
		assert.Equal(t, expected, result, "utf16ToUTF8() should correctly convert large UTF-16LE input")
	})
}

func TestUTF16ToUTF8_SurrogateEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		bigEndian bool
		expected  []byte
		wantErr   bool
	}{
		{
			name:      "UTF-16BE High surrogate without low surrogate",
			input:     []byte{0xD8, 0x00, 0x00, 0x41},
			bigEndian: true,
			expected:  []byte{0xEF, 0xBF, 0xBD, 0x41}, // Replacement character followed by 'A'
			wantErr:   false,
		},
		{
			name:      "UTF-16LE High surrogate without low surrogate",
			input:     []byte{0x00, 0xD8, 0x41, 0x00},
			bigEndian: false,
			expected:  []byte{0xEF, 0xBF, 0xBD, 0x41}, // Replacement character followed by 'A'
			wantErr:   false,
		},
		{
			name:      "UTF-16BE Low surrogate without high surrogate",
			input:     []byte{0xDC, 0x00, 0x00, 0x41},
			bigEndian: true,
			expected:  []byte{0xEF, 0xBF, 0xBD, 0x41}, // Replacement character followed by 'A'
			wantErr:   false,
		},
		{
			name:      "UTF-16LE Low surrogate without high surrogate",
			input:     []byte{0x00, 0xDC, 0x41, 0x00},
			bigEndian: false,
			expected:  []byte{0xEF, 0xBF, 0xBD, 0x41}, // Replacement character followed by 'A'
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := utf16ToUTF8(tt.input, tt.bigEndian)

			if tt.wantErr {
				assert.Error(t, err, "utf16ToUTF8() should return an error for invalid input")
				return
			}

			require.NoError(t, err, "utf16ToUTF8() should not return an error for valid input")
			assert.Equal(t, tt.expected, result, "utf16ToUTF8() should return the correct UTF-8 bytes")
		})
	}
}
