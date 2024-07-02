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
)

// utf32ToUTF8 converts UTF-32 encoded bytes to UTF-8.
//
// Advantages of using bitwise operations for this conversion:
//
//  1. Efficiency: Bitwise operations are highly optimized at the hardware level,
//     often executing in a single CPU cycle. This makes them faster than
//     arithmetic operations or function calls, especially important when
//     processing large volumes of text.
//
//  2. Direct byte manipulation: Bitwise operations allow us to work directly
//     with the binary representation of the data. This is crucial for correctly
//     interpreting UTF-32 encoded characters, which span four bytes.
//
//  3. Precision: When dealing with character encodings, every bit matters.
//     Bitwise operations ensure we're interpreting the data exactly as intended,
//     without any unintended modifications that could occur with higher-level operations.
//
//  4. Endianness handling: UTF-32 can be either big-endian or little-endian.
//     Bitwise operations provide a clean and efficient way to handle both
//     encodings with minimal code duplication.
//
//  5. Memory efficiency: By manipulating bits directly, we avoid the need
//     for intermediate data structures or type conversions, which can be
//     beneficial for memory usage, especially when processing large files.
//
//  6. Portability: Bitwise operations behave consistently across different
//     hardware architectures, ensuring our code works reliably on various systems.
//
//  7. Educational value: Understanding and using bitwise operations provides
//     insights into low-level data representation, which is valuable knowledge
//     for any programmer working with different character encodings or
//     binary protocols.
func utf32ToUTF8(data []byte, bigEndian bool) ([]byte, error) {
	var buf bytes.Buffer

	for i := 0; i < len(data); i += 4 {
		if i+3 >= len(data) {
			return nil, errors.New("incomplete UTF-32 sequence")
		}

		var r rune
		if bigEndian {
			// Big Endian UTF-32 to rune conversion
			// In big endian, the most significant byte comes first

			// Step 1: Convert the first byte to a uint32 and shift it left by 24 bits
			// This operation moves the bits of the first byte to the highest-order position
			// Example: if data[i] is 0x12, after shift it becomes 0x12000000
			firstByte := uint32(data[i]) << 24

			// Step 2: Convert the second byte to a uint32 and shift it left by 16 bits
			// This operation moves the bits of the second byte to the second highest-order position
			// Example: if data[i+1] is 0x34, after shift it becomes 0x00340000
			secondByte := uint32(data[i+1]) << 16

			// Step 3: Convert the third byte to a uint32 and shift it left by 8 bits
			// This operation moves the bits of the third byte to the second lowest-order position
			// Example: if data[i+2] is 0x56, after shift it becomes 0x00005600
			thirdByte := uint32(data[i+2]) << 8

			// Step 4: Convert the fourth byte to a uint32
			// The fourth byte remains in the lowest-order position
			// Example: if data[i+3] is 0x78, it remains 0x00000078
			fourthByte := uint32(data[i+3])

			// Step 5: Combine all four bytes using bitwise OR
			// This operation merges the four bytes into a single 32-bit value
			// Example: 0x12000000 | 0x00340000 | 0x00005600 | 0x00000078 = 0x12345678
			r = rune(firstByte | secondByte | thirdByte | fourthByte)

			// The resulting rune r now contains the 32-bit UTF-32 code point
		} else {
			// Little Endian UTF-32 to rune conversion
			// In little endian, the least significant byte comes first

			// Step 1: Convert the fourth byte to a uint32 and shift it left by 24 bits
			// This operation moves the bits of the fourth byte to the highest-order position
			// Example: if data[i+3] is 0x12, after shift it becomes 0x12000000
			fourthByte := uint32(data[i+3]) << 24

			// Step 2: Convert the third byte to a uint32 and shift it left by 16 bits
			// This operation moves the bits of the third byte to the second highest-order position
			// Example: if data[i+2] is 0x34, after shift it becomes 0x00340000
			thirdByte := uint32(data[i+2]) << 16

			// Step 3: Convert the second byte to a uint32 and shift it left by 8 bits
			// This operation moves the bits of the second byte to the second lowest-order position
			// Example: if data[i+1] is 0x56, after shift it becomes 0x00005600
			secondByte := uint32(data[i+1]) << 8

			// Step 4: Convert the first byte to a uint32
			// The first byte remains in the lowest-order position
			// Example: if data[i] is 0x78, it remains 0x00000078
			firstByte := uint32(data[i])

			// Step 5: Combine all four bytes using bitwise OR
			// This operation merges the four bytes into a single 32-bit value
			// Example: 0x12000000 | 0x00340000 | 0x00005600 | 0x00000078 = 0x12345678
			r = rune(fourthByte | thirdByte | secondByte | firstByte)

			// The resulting rune r now contains the 32-bit UTF-32 code point
		}

		// Write the rune to the buffer
		// The WriteRune method automatically handles the conversion from the rune to UTF-8
		buf.WriteRune(r)
	}

	return buf.Bytes(), nil
}
