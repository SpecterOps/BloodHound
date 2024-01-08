// Copyright 2023 Specter Ops, Inc.
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

package pgsql

import (
	"io"
	"strconv"
	"strings"
)

func JoinUint[T uint | uint8 | uint16 | uint32 | uint64](values []T, separator string) string {
	builder := strings.Builder{}

	for idx := 0; idx < len(values); idx++ {
		if idx > 0 {
			builder.WriteString(separator)
		}

		builder.WriteString(strconv.FormatUint(uint64(values[idx]), 10))
	}

	return builder.String()
}

func JoinInt[T int | int8 | int16 | int32 | int64](values []T, separator string) string {
	builder := strings.Builder{}

	for idx := 0; idx < len(values); idx++ {
		if idx > 0 {
			builder.WriteString(separator)
		}

		builder.WriteString(strconv.FormatInt(int64(values[idx]), 10))
	}

	return builder.String()
}

func WriteStrings(writer io.Writer, strings ...string) (int, error) {
	totalBytesWritten := 0

	for idx := 0; idx < len(strings); idx++ {
		if bytesWritten, err := io.WriteString(writer, strings[idx]); err != nil {
			return totalBytesWritten, err
		} else {
			totalBytesWritten += bytesWritten
		}
	}

	return totalBytesWritten, nil
}
