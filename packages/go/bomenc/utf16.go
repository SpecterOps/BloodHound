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

// utf16ToUTF8 converts a UTF-16 encoded byte slice to UTF-8.
// It handles both big-endian and little-endian encodings, as well as surrogate pairs.
//
// Parameters:
//   - data: A byte slice containing UTF-16 encoded data
//   - bigEndian: A boolean indicating whether the input is big-endian (true) or little-endian (false)
//
// Returns:
//   - A byte slice containing the UTF-8 encoded result
//   - An error if the conversion fails (e.g., due to incomplete input)
//
// Advantages of using bitwise operations for this conversion:
//
//  1. Efficiency: Bitwise operations are extremely fast at the CPU level.
//     They typically execute in a single clock cycle, making them more
//     efficient than arithmetic operations or function calls.
//
//  2. Direct memory manipulation: Bitwise operations allow us to work
//     directly with the binary representation of the data, which is
//     crucial when dealing with different byte-order encodings.
//
//  3. Preserving data integrity: By using bitwise operations, we ensure
//     that we're interpreting the bytes exactly as they are, without any
//     unintended modifications that could occur with higher-level operations.
//
//  4. Endianness handling: Bitwise operations make it easy to handle both
//     big-endian and little-endian encodings with minimal code duplication.
//
//  5. Performance in loops: When processing large amounts of text, the
//     performance benefits of bitwise operations become significant due
//     to the number of times these operations are repeated.
//
//  6. Low-level control: Bitwise operations provide fine-grained control
//     over individual bits, which is necessary for correct interpretation
//     of multi-byte character encodings.
func utf16ToUTF8(data []byte, bigEndian bool) ([]byte, error) {
	var buf bytes.Buffer
	var r rune

	for i := 0; i < len(data); i += 2 {
		if i+1 >= len(data) {
			return nil, errors.New("incomplete UTF-16 sequence")
		}

		var codeUnit uint16
		if bigEndian {
			// Big-endian: first byte is more significant
			// Shift the first byte left by 8 bits and OR it with the second byte
			// Example:
			//   data[i]   = 0x12 (00010010 in binary)
			//   data[i+1] = 0x34 (00110100 in binary)
			//   result    = 0x1234 (0001001000110100 in binary)
			codeUnit = uint16(data[i])<<8 | uint16(data[i+1])
		} else {
			// Little-endian: second byte is more significant
			// Shift the second byte left by 8 bits and OR it with the first byte
			// Example:
			//   data[i]   = 0x34 (00110100 in binary)
			//   data[i+1] = 0x12 (00010010 in binary)
			//   result    = 0x1234 (0001001000110100 in binary)
			codeUnit = uint16(data[i+1])<<8 | uint16(data[i])
		}

		if codeUnit >= 0xD800 && codeUnit <= 0xDBFF {
			if i+3 >= len(data) {
				buf.WriteRune(0xFFFD)
				break
			}

			var lowSurrogate uint16
			if bigEndian {
				lowSurrogate = uint16(data[i+2])<<8 | uint16(data[i+3])
			} else {
				lowSurrogate = uint16(data[i+3])<<8 | uint16(data[i+2])
			}

			if lowSurrogate >= 0xDC00 && lowSurrogate <= 0xDFFF {
				// Combine high and low surrogates into a single code point
				// 1. Subtract 0xD800 from the high surrogate to get the high 10 bits
				// 2. Subtract 0xDC00 from the low surrogate to get the low 10 bits
				// 3. Shift the high 10 bits left by 10 positions
				// 4. OR the result with the low 10 bits
				// 5. Add 0x10000 to get the final code point
				// Example:
				//   codeUnit     = 0xD801 (1101100000000001 in binary)
				//   lowSurrogate = 0xDC37 (1101110000110111 in binary)
				//   Step 1: 0xD801 - 0xD800 = 0x0001
				//   Step 2: 0xDC37 - 0xDC00 = 0x0037
				//   Step 3: 0x0001 << 10 = 0x0400
				//   Step 4: 0x0400 | 0x0037 = 0x0437
				//   Step 5: 0x0437 + 0x10000 = 0x10437 (Code point U+10437)
				r = (rune(codeUnit-0xD800)<<10 | rune(lowSurrogate-0xDC00)) + 0x10000
				i += 2
			} else {
				r = 0xFFFD
			}
		} else if codeUnit >= 0xDC00 && codeUnit <= 0xDFFF {
			r = 0xFFFD
		} else {
			r = rune(codeUnit)
		}

		buf.WriteRune(r)
	}

	return buf.Bytes(), nil
}
