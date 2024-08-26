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
	"encoding/binary"
	"io"
)

// Encoding interface defines the methods that all encoding types must implement.
// This interface provides a unified way to handle different encodings throughout the package,
// allowing us to treat all encodings uniformly. This design facilitates easy extension
// and manipulation of different encoding types without altering the core logic.
type Encoding interface {
	// Sequence returns the byte sequence that represents the Byte Order Mark (BOM) for this encoding.
	// This method is crucial for identifying the specific byte sequence that indicates
	// this encoding at the start of a file.
	Sequence() []byte

	// String returns a human-readable string representation of the encoding.
	// This is particularly useful for logging and user interfaces, providing
	// a user-friendly name for the encoding.
	String() string

	// HasSequence checks if the given data starts with this encoding's BOM sequence.
	// This method allows for efficient checking of whether a given byte slice
	// begins with this encoding's BOM, which is essential for encoding detection.
	HasSequence(data Peeker) bool
}

// Peeker interface defines a single method for introspecing the first n number of bytes in the underlying structure without modifying its read state or advancing its cursor.
type Peeker interface {
	Peek(n int) ([]byte, error)
}

// bomEncoding is the concrete implementation of the Encoding interface.
// It encapsulates all necessary information and behavior for a specific encoding,
// providing a consistent structure for handling different encodings. This approach
// allows us to create instances for each supported encoding while maintaining
// a uniform interface for interaction.
type bomEncoding struct {
	encodingType    string                 // A human-readable name for the encoding
	sequence        []byte                 // The BOM sequence for this encoding
	hasSequenceFunc func(data Peeker) bool // Function to check if data starts with this encoding's BOM
}

// String returns the human-readable name of the encoding.
// This method fulfills the Encoding interface and provides a simple way
// to get a string representation of the encoding.
func (s bomEncoding) String() string {
	return s.encodingType
}

// Sequence returns the BOM sequence for this encoding.
// This method fulfills the Encoding interface and provides access to the BOM sequence,
// which is essential for encoding detection and writing files with proper BOMs.
func (s bomEncoding) Sequence() []byte {
	return s.sequence
}

// HasSequence checks if the given data starts with this encoding's BOM sequence.
// This method fulfills the Encoding interface and provides a way to check for
// the presence of this encoding's BOM, which is crucial for encoding detection.
func (s bomEncoding) HasSequence(data Peeker) bool {
	return s.hasSequenceFunc(data)
}

// The following functions are used to check for specific encoding BOMs.
// By defining these as separate functions, we can easily reuse them
// and potentially extend them if more complex checking is needed in the future.
// This approach also keeps the bomEncoding struct clean and simple.

func isUTF32BE(data Peeker) bool {
	if buf, err := data.Peek(4); err != nil {
		return false
	} else {
		return buf[0] == 0x00 && buf[1] == 0x00 && buf[2] == 0xFE && buf[3] == 0xFF
	}
}

func isUTF32LE(data Peeker) bool {
	if buf, err := data.Peek(4); err != nil {
		return false
	} else if buf[0] != 0xFF || buf[1] != 0xFE || buf[2] != 0x00 || buf[3] != 0x00 {
		return false
	} else if buf, err := data.Peek(64); err != nil && err != io.EOF { // BOM + sample code points to check for valid sequences
		return false
	} else if err != nil && err == io.EOF && len(buf)%4 != 0 {
		return false
	} else {
		// Check for valid UTF-32LE sequences
		for i := 4; i+3 < len(buf); i += 4 {
			codePoint := binary.LittleEndian.Uint32(buf[i : i+4])
			if codePoint > 0x10FFFF {
				return false
			}
		}
		// NOTE: There is an edge case where data may may include the UTF16LE BOM and a NULL code point
		// followed by sampled code points that, when calculated, all fall within the unicode range. In this
		// case, distinguishing between UTF32LE and UTF16LE encoded data is impossible from just the BOM + data.
		// With the probability of occurence being low, we're opting to return true
		return true
	}
}

func isUTF8(data Peeker) bool {
	if buf, err := data.Peek(3); err != nil {
		return false
	} else {
		return buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF
	}
}

func isUTF16BE(data Peeker) bool {
	if buf, err := data.Peek(2); err != nil {
		return false
	} else {
		return buf[0] == 0xFE && buf[1] == 0xFF
	}
}

func isUTF16LE(data Peeker) bool {
	if buf, err := data.Peek(2); err != nil {
		return false
	} else {
		if buf[0] == 0xFF && buf[1] == 0xFE {
			if buf, err := data.Peek(4); err != nil {
				return err == io.EOF && len(buf) == 2 // true: has UTF16LE BOM w/ no data, false: is invalid UTF16LE encoding
			} else {
				return !isUTF32LE(data) // true: is UTF16LE data, false: is UTF32LE data
			}
		}
		return false
	}
}

// The following variables define the supported encodings.
// By defining these as package-level variables, we allow easy reference
// throughout the package and by users of the package. This design also
// facilitates potential future extension by simply adding new encoding variables.

// Unknown represents an unknown or unrecognized encoding.
// Having an Unknown encoding allows us to handle cases where
// the encoding cannot be determined, providing a fallback option.
var Unknown Encoding = bomEncoding{
	encodingType:    "Unknown",
	sequence:        nil, // Unknown encoding has no BOM sequence
	hasSequenceFunc: func(data Peeker) bool { return false },
}

// UTF8 represents the UTF-8 encoding.
var UTF8 Encoding = bomEncoding{
	encodingType:    "UTF-8",
	sequence:        []byte{0xEF, 0xBB, 0xBF}, // UTF-8 BOM sequence
	hasSequenceFunc: isUTF8,
}

// UTF16BE represents the UTF-16 Big Endian encoding.
var UTF16BE Encoding = bomEncoding{
	encodingType:    "UTF-16 BE",
	sequence:        []byte{0xFE, 0xFF}, // UTF-16 BE BOM sequence
	hasSequenceFunc: isUTF16BE,
}

// UTF16LE represents the UTF-16 Little Endian encoding.
var UTF16LE Encoding = bomEncoding{
	encodingType:    "UTF-16 LE",
	sequence:        []byte{0xFF, 0xFE}, // UTF-16 LE BOM sequence
	hasSequenceFunc: isUTF16LE,
}

// UTF32BE represents the UTF-32 Big Endian encoding.
var UTF32BE Encoding = bomEncoding{
	encodingType:    "UTF-32 BE",
	sequence:        []byte{0x00, 0x00, 0xFE, 0xFF}, // UTF-32 BE BOM sequence
	hasSequenceFunc: isUTF32BE,
}

// UTF32LE represents the UTF-32 Little Endian encoding.
var UTF32LE Encoding = bomEncoding{
	encodingType:    "UTF-32 LE",
	sequence:        []byte{0xFF, 0xFE, 0x00, 0x00}, // UTF-32 LE BOM sequence
	hasSequenceFunc: isUTF32LE,
}
