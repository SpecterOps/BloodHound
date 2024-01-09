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

package utils

import (
	"context"
	"io"
)

func CopyWithCtx(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	buffer := make([]byte, 4096)
	var total int64 = 0

	for {
		read, err := src.Read(buffer)

		// check context after read
		if err := ctx.Err(); err != nil {
			return total, err
		}

		if read > 0 {
			written, err := dst.Write(buffer[:read])
			total += int64(written)
			if err != nil {
				return total, err
			}
		}

		if err != nil {
			if err != io.EOF {
				return total, err
			}

			return total, nil
		}

		// check context after writes before next read
		if err := ctx.Err(); err != nil {
			return total, err
		}
	}
}
