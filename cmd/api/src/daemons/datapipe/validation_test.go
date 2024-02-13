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

package datapipe_test

import (
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func Test_ValidateMetaTag(t *testing.T) {
	validJson := `{
    "meta": {
        "methods": 0,
        "type": "sessions",
        "count": 0,
        "version": 5
    },
    "data": []
	}`

	meta, err := datapipe.ValidateMetaTag(strings.NewReader(validJson))
	assert.Nil(t, err)
	assert.Equal(t, datapipe.DataTypeSession, meta.Type)

	missingDataTag := `{
    "meta": {
        "methods": 0,
        "type": "sessions",
        "count": 0,
        "version": 5
    }
	}`

	meta, err = datapipe.ValidateMetaTag(strings.NewReader(missingDataTag))
	assert.Equal(t, err, datapipe.ErrDataTagNotFound)

	missingMetaTag := `{
    "data": []
	}`

	meta, err = datapipe.ValidateMetaTag(strings.NewReader(missingMetaTag))
	assert.Equal(t, err, datapipe.ErrMetaTagNotFound)

	missingBothTag := `{
	}`

	meta, err = datapipe.ValidateMetaTag(strings.NewReader(missingBothTag))
	assert.Equal(t, err, datapipe.ErrNoTagFound)

	ignoreInvalidTag := `{
	"meta": 0,
    "meta": {
        "methods": 0,
        "type": "sessions",
        "count": 0,
        "version": 5
    },
    "data": []
	}`

	meta, err = datapipe.ValidateMetaTag(strings.NewReader(ignoreInvalidTag))
	assert.Nil(t, err)
	assert.Equal(t, datapipe.DataTypeSession, meta.Type)

	swapOrder := `{
	"data": [],
    "meta": {
        "methods": 0,
        "type": "sessions",
        "count": 0,
        "version": 5
    }
	}`

	meta, err = datapipe.ValidateMetaTag(strings.NewReader(swapOrder))
	assert.Nil(t, err)
	assert.Equal(t, datapipe.DataTypeSession, meta.Type)
}
