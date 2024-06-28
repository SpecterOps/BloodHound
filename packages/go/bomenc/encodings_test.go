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

	"github.com/stretchr/testify/assert"
)

func TestEncodingInterface(t *testing.T) {
	encodings := []struct {
		name     string
		encoding Encoding
	}{
		{name: "Unknown", encoding: Unknown},
		{name: "UTF8", encoding: UTF8},
		{name: "UTF16BE", encoding: UTF16BE},
		{name: "UTF16LE", encoding: UTF16LE},
		{name: "UTF32BE", encoding: UTF32BE},
		{name: "UTF32LE", encoding: UTF32LE},
	}

	for _, tt := range encodings {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.encoding.String(), "Encoding String() should not be empty")
			if tt.encoding.String() != Unknown.String() {
				assert.NotEmpty(t, tt.encoding.Sequence(), "Encoding Sequence() should not be empty for non-Unknown encodings")
			}
			// Test HasSequence method
			if tt.encoding.String() != Unknown.String() {
				assert.True(t, tt.encoding.HasSequence(tt.encoding.Sequence()), "HasSequence() should return true for its own sequence")
			}
		})
	}
}

func TestEncodingValues(t *testing.T) {
	tests := []struct {
		name         string
		encoding     Encoding
		expectedType string
		expectedSeq  []byte
	}{
		{
			name:         "Unknown",
			encoding:     Unknown,
			expectedType: "Unknown",
			expectedSeq:  nil,
		},
		{
			name:         "UTF-8",
			encoding:     UTF8,
			expectedType: "UTF-8",
			expectedSeq:  []byte{0xEF, 0xBB, 0xBF},
		},
		{
			name:         "UTF-16 BE",
			encoding:     UTF16BE,
			expectedType: "UTF-16 BE",
			expectedSeq:  []byte{0xFE, 0xFF},
		},
		{
			name:         "UTF-16 LE",
			encoding:     UTF16LE,
			expectedType: "UTF-16 LE",
			expectedSeq:  []byte{0xFF, 0xFE},
		},
		{
			name:         "UTF-32 BE",
			encoding:     UTF32BE,
			expectedType: "UTF-32 BE",
			expectedSeq:  []byte{0x00, 0x00, 0xFE, 0xFF},
		},
		{
			name:         "UTF-32 LE",
			encoding:     UTF32LE,
			expectedType: "UTF-32 LE",
			expectedSeq:  []byte{0xFF, 0xFE, 0x00, 0x00},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedType, tt.encoding.String(), "Encoding type should match")
			assert.Equal(t, tt.expectedSeq, tt.encoding.Sequence(), "Encoding sequence should match")
			if tt.encoding.String() != Unknown.String() {
				assert.True(t, tt.encoding.HasSequence(tt.expectedSeq), "HasSequence() should return true for the expected sequence")
			}
		})
	}
}

func TestBOMEncoding(t *testing.T) {
	testCases := []struct {
		name           string
		encoding       bomEncoding
		expectedString string
		expectedSeq    []byte
		testData       []byte
		hasSequence    bool
	}{
		{
			name: "Custom encoding",
			encoding: bomEncoding{
				encodingType: "Custom",
				sequence:     []byte{0x01, 0x02, 0x03},
				hasSequenceFunc: func(data []byte) bool {
					return len(data) >= 3 && data[0] == 0x01 && data[1] == 0x02 && data[2] == 0x03
				},
			},
			expectedString: "Custom",
			expectedSeq:    []byte{0x01, 0x02, 0x03},
			testData:       []byte{0x01, 0x02, 0x03, 0x04},
			hasSequence:    true,
		},
		{
			name: "Empty encoding",
			encoding: bomEncoding{
				encodingType:    "",
				sequence:        []byte{},
				hasSequenceFunc: func(data []byte) bool { return len(data) == 0 },
			},
			expectedString: "",
			expectedSeq:    []byte{},
			testData:       []byte{0x01, 0x02, 0x03},
			hasSequence:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.encoding.String(), "bomEncoding String() should return correct value")
			assert.Equal(t, tc.expectedSeq, tc.encoding.Sequence(), "bomEncoding Sequence() should return correct value")
			assert.Equal(t, tc.hasSequence, tc.encoding.HasSequence(tc.testData), "bomEncoding HasSequence() should return correct value")
		})
	}
}

func TestEncodingEquality(t *testing.T) {
	testCases := []struct {
		name     string
		enc1     Encoding
		enc2     Encoding
		expected bool
	}{
		{
			name:     "Same encoding",
			enc1:     UTF8,
			enc2:     UTF8,
			expected: true,
		},
		{
			name:     "Different encodings",
			enc1:     UTF8,
			enc2:     UTF16BE,
			expected: false,
		},
		{
			name:     "Unknown and other encoding",
			enc1:     Unknown,
			enc2:     UTF8,
			expected: false,
		},
		{
			name:     "Both Unknown",
			enc1:     Unknown,
			enc2:     Unknown,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.enc1.String() == tc.enc2.String(), "Encoding equality check should be correct")
		})
	}
}

func TestHasSequence(t *testing.T) {
	testCases := []struct {
		name     string
		encoding Encoding
		input    []byte
		expected bool
	}{
		{"UTF-8 with correct BOM", UTF8, []byte{0xEF, 0xBB, 0xBF, 0x68, 0x65, 0x6C, 0x6C, 0x6F}, true},
		{"UTF-8 without BOM", UTF8, []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F}, false},
		{"UTF-16BE with correct BOM", UTF16BE, []byte{0xFE, 0xFF, 0x00, 0x68, 0x00, 0x65}, true},
		{"UTF-16BE without BOM", UTF16BE, []byte{0x00, 0x68, 0x00, 0x65}, false},
		{"UTF-16LE with correct BOM", UTF16LE, []byte{0xFF, 0xFE, 0x68, 0x00, 0x65, 0x00}, true},
		{"UTF-16LE without BOM", UTF16LE, []byte{0x68, 0x00, 0x65, 0x00}, false},
		{"UTF-32BE with correct BOM", UTF32BE, []byte{0x00, 0x00, 0xFE, 0xFF, 0x00, 0x00, 0x00, 0x68}, true},
		{"UTF-32BE without BOM", UTF32BE, []byte{0x00, 0x00, 0x00, 0x68}, false},
		{"UTF-32LE with correct BOM", UTF32LE, []byte{0xFF, 0xFE, 0x00, 0x00, 0x68, 0x00, 0x00, 0x00}, true},
		{"UTF-32LE without BOM", UTF32LE, []byte{0x68, 0x00, 0x00, 0x00}, false},
		{"Unknown encoding", Unknown, []byte{0x68, 0x65, 0x6C, 0x6C, 0x6F}, false},
		{"Empty input", UTF8, []byte{}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.encoding.HasSequence(tc.input)
			assert.Equal(t, tc.expected, result, "HasSequence() should correctly identify BOM presence")
		})
	}
}
