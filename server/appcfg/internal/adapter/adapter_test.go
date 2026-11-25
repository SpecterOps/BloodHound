// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package adapter_test

import (
	"testing"

	"github.com/specterops/bloodhound/server/appcfg/internal/adapter"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
	"github.com/specterops/bloodhound/server/appcfg/internal/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigurationAdapter_Dependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		buildReader   func(*testing.T) *services.Reader
		expectedPanic string
	}{
		{
			name: "Success: constructs configuration adapter",
			buildReader: func(t *testing.T) *services.Reader {
				return services.NewReader(mocks.NewMockDatabase(t))
			},
		},
		{
			name:          "Error: nil configuration reader",
			expectedPanic: "appcfg: adapter requires a non-nil configuration reader",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var reader *services.Reader
			if testCase.buildReader != nil {
				reader = testCase.buildReader(t)
			}
			construct := func() {
				configurationAdapter := adapter.NewConfigurationAdapter(reader)
				assert.NotNil(t, configurationAdapter)
			}

			if testCase.expectedPanic != "" {
				assert.PanicsWithValue(t, testCase.expectedPanic, construct)
			} else {
				assert.NotPanics(t, construct)
			}
		})
	}
}
