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
	"github.com/stretchr/testify/assert"
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
			writer  = bytes.Buffer{}
			badJSON = strings.NewReader("{\"redPill\": true, \"bluePill\": false}")
		)
		err := WriteAndValidateJSON(badJSON, &writer)
		assert.Nil(t, err)
	})
}
