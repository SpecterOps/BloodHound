// Copyright 2025 Specter Ops, Inc.
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

package upload

import (
	"archive/zip"
	"fmt"
	"io"

	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/packages/go/bomenc"
)

var ZipMagicBytes = []byte{0x50, 0x4b, 0x03, 0x04}

// ReadZippedFile - Util Function to help read zipped files
func ReadZippedFile(zf *zip.File) ([]byte, error) {
	f, err := zf.Open()
	if err != nil {
		return nil, err
	}

	defer f.Close()

	if normFile, err := bomenc.NormalizeToUTF8(f); err != nil {
		return nil, fmt.Errorf("failed to normalize json file: %w", err)
	} else {
		return io.ReadAll(normFile)
	}
}

func ValidateZipFile(reader io.Reader) error {
	bytes := make([]byte, 4)
	if readBytes, err := reader.Read(bytes); err != nil {
		return err
	} else if readBytes < 4 {
		return ingest.ErrInvalidZipFile
	} else {
		for i := 0; i < 4; i++ {
			if bytes[i] != ZipMagicBytes[i] {
				return ingest.ErrInvalidZipFile
			}
		}

		_, err := io.Copy(io.Discard, reader)

		return err
	}
}
