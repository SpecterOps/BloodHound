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

	"github.com/specterops/bloodhound/cmd/api/src/utils/validation"
)

func TestUrlValidator(t *testing.T) {
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

	t.Run("Invalid HTTPS option in struct tag", func(t *testing.T) {
		type testStruct struct {
			URL string `validate:"url,httpsOnly=typo"`
		}

		errs := validation.Validate(&testStruct{URL: "https://test.com"})
		require.Len(t, errs, 1)
		assert.EqualError(t, errs[0], `URL: URL validator option "httpsOnly" must be a boolean`)
	})
}

func TestURLValidatorHostnameOption(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		options map[string]string
		url     string
		wantErr string
	}{
		{
			name:    "hostname is optional by default",
			options: map[string]string{"httpsOnly": "true"},
			url:     "https:///path",
		},
		{
			name:    "hostname requirement can be disabled",
			options: map[string]string{"httpsOnly": "true", "hostnameRequired": "false"},
			url:     "https:///path",
		},
		{
			name:    "hostname is required when enabled",
			options: map[string]string{"httpsOnly": "true", "hostnameRequired": "true"},
			url:     "https:///path",
			wantErr: validation.ErrorURLHostnameRequired,
		},
		{
			name:    "hostname requirement works without requiring HTTPS",
			options: map[string]string{"hostnameRequired": "true"},
			url:     "http://test.com",
		},
		{
			name:    "HTTPS scheme matching is case insensitive when a hostname is required",
			options: map[string]string{"httpsOnly": "true", "hostnameRequired": "true"},
			url:     "HTTPS://test.com",
		},
		{
			name:    "HTTP is rejected when both options are enabled",
			options: map[string]string{"httpsOnly": "true", "hostnameRequired": "true"},
			url:     "http://test.com",
			wantErr: validation.ErrorURLHTTPSRequired,
		},
		{
			name:    "invalid hostname setting fails validation",
			options: map[string]string{"httpsOnly": "true", "hostnameRequired": "typo"},
			url:     "https://test.com",
			wantErr: `URL validator option "hostnameRequired" must be a boolean`,
		},
		{
			name:    "an explicitly empty setting fails validation",
			options: map[string]string{"hostnameRequired": ""},
			url:     "https://test.com",
			wantErr: `URL validator option "hostnameRequired" must be a boolean`,
		},
		{
			name:    "unknown option fails validation",
			options: map[string]string{"hostnameRequire": "true"},
			url:     "https://test.com",
			wantErr: `unknown URL validator option "hostnameRequire"`,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			errs := validation.NewUrlValidator(testCase.options).Validate(testCase.url)

			if testCase.wantErr == "" {
				require.Empty(t, errs)
			} else {
				require.Len(t, errs, 1)
				assert.EqualError(t, errs[0], testCase.wantErr)
			}
		})
	}
}
