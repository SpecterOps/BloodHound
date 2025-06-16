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

package validation_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/src/utils/validation"
)

func TestUrlValidator(t *testing.T) {
	t.Parallel()
	t.Run("Valid URL", func(t *testing.T) {
		type testStruct struct {
			URL string `validate:"url"`
		}

		errs := validation.Validate(&testStruct{URL: "http://test.com"})
		require.Len(t, errs, 0)
	})

	t.Run("Invalid HTTP URL", func(t *testing.T) {
		type testStruct struct {
			URL string `validate:"url"`
		}
		errs := validation.Validate(&testStruct{URL: "bloodhound"})
		require.Len(t, errs, 1)
		assert.ErrorContains(t, errs[0], validation.ErrorUrlInvalid)
	})

	t.Run("Invalid HTTPS URL", func(t *testing.T) {
		type testStruct struct {
			URL string `validate:"url,httpsOnly=true"`
		}
		errs := validation.Validate(&testStruct{URL: "http://bloodhound.com"})
		require.Len(t, errs, 1)
		assert.ErrorContains(t, errs[0], validation.ErrorUrlHttpsInvalid)
	})
}
