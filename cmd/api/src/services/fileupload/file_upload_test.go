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

package fileupload

import (
	"bytes"
	"github.com/specterops/bloodhound/src/model/ingest"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWriteAndValidateJSON(t *testing.T) {
	t.Run("trigger invalid json on bad json", func(t *testing.T) {
		var (
			writer  = bytes.Buffer{}
			badJSON = strings.NewReader("{[]}")
		)
		err := WriteAndValidateJSON(badJSON, &writer)
		assert.ErrorIs(t, err, ErrInvalidJSON)
	})

	t.Run("succeed on good json", func(t *testing.T) {
		var (
			writer   = bytes.Buffer{}
			goodJSON = strings.NewReader(`{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "data": []}`)
		)
		err := WriteAndValidateJSON(goodJSON, &writer)
		assert.Nil(t, err)
	})

	t.Run("succeed on utf-8 BOM json", func(t *testing.T) {
		var (
			writer = bytes.Buffer{}
		)

		file, err := os.Open("../../test/fixtures/fixtures/utf8bomjson.json")
		assert.Nil(t, err)
		err = WriteAndValidateJSON(io.Reader(file), &writer)
		assert.Nil(t, err)
	})
}

func TestWriteAndValidateZip(t *testing.T) {
	t.Run("valid zip file is ok", func(t *testing.T) {
		var (
			writer = bytes.Buffer{}
		)

		file, err := os.Open("../../test/fixtures/fixtures/goodzip.zip")
		assert.Nil(t, err)

		err = WriteAndValidateZip(io.Reader(file), &writer)
		assert.Nil(t, err)
	})

	t.Run("invalid bytes causes error", func(t *testing.T) {
		var (
			writer = bytes.Buffer{}
			badZip = strings.NewReader("123123")
		)

		err := WriteAndValidateZip(badZip, &writer)
		assert.Equal(t, err, ingest.ErrInvalidZipFile)
	})
}
