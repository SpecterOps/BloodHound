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
	"fmt"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/cmd/api/src/utils/validation"
)

func TestDurationValidator(t *testing.T) {

	type Epic struct {
		Duration time.Duration `validate:"duration,min=P1D,max=P14D"`
	}

	fortnight := time.Hour * 24 * 14

	minErr := fmt.Errorf("Duration: "+validation.ErrorDurationMin, "P1D")
	maxErr := fmt.Errorf("Duration: "+validation.ErrorDurationMax, "P14D")

	var cases = []struct {
		Input  Epic
		Errors utils.Errors
	}{
		{Epic{time.Hour}, utils.Errors{minErr}},
		{Epic{fortnight}, nil},
		{Epic{fortnight + time.Hour}, utils.Errors{maxErr}},
	}

	for _, tc := range cases {
		if errs := validation.Validate(tc.Input); errs != nil {
			if tc.Errors == nil {
				t.Errorf("For input: %v, expected errs to be nil: errs = %v\n", tc.Input, errs)
			} else if errs.Error() != tc.Errors.Error() {
				t.Errorf("For input: %v, got %v, want %v\n", tc.Input, errs, tc.Errors)
			}
		} else if tc.Errors != nil {
			t.Errorf("For input: %v, expected errs to not be nil\n", tc.Input)
		}
	}
}
