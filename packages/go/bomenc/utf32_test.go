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
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUTF32ToUTF8(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		bigEndian bool
		expected  []byte
		wantErr   bool
	}{
		{
			name:      "UTF-32BE Basic ASCII",
			input:     []byte{0x00, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6F},
			bigEndian: true,
			expected:  []byte("hello"),
			wantErr:   false,
		},
		{
			name:      "UTF-32LE Basic ASCII",
			input:     []byte{0x68, 0x00, 0x00, 0x00, 0x65, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6C, 0x00, 0x00, 0x00, 0x6F, 0x00, 0x00, 0x00},
			bigEndian: false,
			expected:  []byte("hello"),
			wantErr:   false,
		},
		{
			name:      "UTF-32BE with BMP characters",
			input:     []byte{0x00, 0x00, 0x00, 0x41, 0x00, 0x00, 0x26, 0x3A, 0x00, 0x00, 0x00, 0x42},
			bigEndian: true,
			expected:  []byte("A☺B"),
			wantErr:   false,
		},
		{
			name:      "UTF-32LE with BMP characters",
			input:     []byte{0x41, 0x00, 0x00, 0x00, 0x3A, 0x26, 0x00, 0x00, 0x42, 0x00, 0x00, 0x00},
			bigEndian: false,
			expected:  []byte("A☺B"),
			wantErr:   false,
		},
		{
			name:      "UTF-32BE with non-BMP character",
			input:     []byte{0x00, 0x01, 0xF4, 0x00},
			bigEndian: true,
			expected:  []byte("🐀"),
			wantErr:   false,
		},
		{
			name:      "UTF-32LE with non-BMP character",
			input:     []byte{0x00, 0xF4, 0x01, 0x00},
			bigEndian: false,
			expected:  []byte("🐀"),
			wantErr:   false,
		},
		{
			name:      "Incomplete UTF-32BE sequence",
			input:     []byte{0x00, 0x00, 0x00},
			bigEndian: true,
			expected:  nil,
			wantErr:   true,
		},
		{
			name:      "Incomplete UTF-32LE sequence",
			input:     []byte{0x00, 0x00, 0x00},
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
			result, err := utf32ToUTF8(tt.input, tt.bigEndian)

			if tt.wantErr {
				assert.Error(t, err, "utf32ToUTF8() should return an error for invalid input")
				return
			}

			require.NoError(t, err, "utf32ToUTF8() should not return an error for valid input")
			assert.Equal(t, tt.expected, result, "utf32ToUTF8() should return the correct UTF-8 bytes")
		})
	}
}

func TestUTF32ToUTF8_LargeInput(t *testing.T) {
	// Generate a large input with 1000 characters
	largeInputBE := make([]byte, 4000)
	largeInputLE := make([]byte, 4000)
	var expected []byte

	for i := 0; i < 1000; i++ {
		codePoint := rune(i % 0x10FFFF) // Use all possible Unicode code points

		// Big Endian
		largeInputBE[i*4] = byte(codePoint >> 24)
		largeInputBE[i*4+1] = byte(codePoint >> 16)
		largeInputBE[i*4+2] = byte(codePoint >> 8)
		largeInputBE[i*4+3] = byte(codePoint)

		// Little Endian
		largeInputLE[i*4] = byte(codePoint)
		largeInputLE[i*4+1] = byte(codePoint >> 8)
		largeInputLE[i*4+2] = byte(codePoint >> 16)
		largeInputLE[i*4+3] = byte(codePoint >> 24)

		// Append UTF-8 encoded rune to expected result
		buf := make([]byte, 4)
		n := utf8.EncodeRune(buf, codePoint)
		expected = append(expected, buf[:n]...)
	}

	t.Run("Large UTF-32BE input", func(t *testing.T) {
		result, err := utf32ToUTF8(largeInputBE, true)
		require.NoError(t, err, "utf32ToUTF8() should not return an error for valid large input")
		assert.Equal(t, expected, result, "utf32ToUTF8() should correctly convert large UTF-32BE input")
	})

	t.Run("Large UTF-32LE input", func(t *testing.T) {
		result, err := utf32ToUTF8(largeInputLE, false)
		require.NoError(t, err, "utf32ToUTF8() should not return an error for valid large input")
		assert.Equal(t, expected, result, "utf32ToUTF8() should correctly convert large UTF-32LE input")
	})
}
