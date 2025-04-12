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
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/model/ingest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dataTagAssertion struct {
	rawString string
	err       error
}

func TestSeekToDataTag(t *testing.T) {
	t.Run("seek to data tag", func(t *testing.T) {
		key := "data"
		assertions := generateAssertionsForKey(key)

		for _, assertion := range assertions {
			r := strings.NewReader(assertion.rawString)
			j := json.NewDecoder(r)

			err := datapipe.SeekToKey(j, key, 1)
			assert.ErrorIs(t, err, assertion.err)
		}
	})

	t.Run("seek to nodes tag at depth 2", func(t *testing.T) {
		r := strings.NewReader(`{"graph":{"nodes":[]}}`)
		j := json.NewDecoder(r)

		err := datapipe.SeekToKey(j, "nodes", 2)
		require.Nil(t, err)
	})
}

func generateAssertionsForKey(key string) []dataTagAssertion {
	return []dataTagAssertion{
		{
			rawString: fmt.Sprintf("{\"%s\": []}", key),
			err:       nil,
		},
		{
			rawString: fmt.Sprintf("{\"%s\": {}}", key),
			err:       ingest.ErrInvalidDataTag,
		},
		{
			rawString: fmt.Sprintf("{\"%s\": ]}", key),
			err:       ingest.ErrJSONDecoderInternal,
		},
		{
			rawString: "",
			err:       ingest.ErrDataTagNotFound,
		},
		{
			rawString: "{[]}",
			err:       ingest.ErrJSONDecoderInternal,
		},
		{
			rawString: fmt.Sprintf("{\"%s\": \"oops\"}", key),
			err:       ingest.ErrInvalidDataTag,
		},
		{
			rawString: "{\"nothing\": [}",
			err:       ingest.ErrJSONDecoderInternal,
		},
		{
			rawString: fmt.Sprintf(`{"meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "%s": []}`, key),
			err:       nil,
		},
		{
			rawString: fmt.Sprintf(`{"test": {"%s": {}}, "meta": {"methods": 0, "type": "sessions", "count": 0, "version": 5}, "%s": []}`, key, key),
			err:       nil,
		},
	}
}
