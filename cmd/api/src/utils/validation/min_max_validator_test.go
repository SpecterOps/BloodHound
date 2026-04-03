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

	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/cmd/api/src/utils/validation"
)

func TestMinMaxValidator(t *testing.T) {

	type TestIntValidator struct {
		Value any `validate:"integer,min=1,max=365"`
	}

	validValue := 120

	minErr := fmt.Errorf("Value: "+validation.ErrorMin, "1")
	maxErr := fmt.Errorf("Value: "+validation.ErrorMax, "365")
	nonIntErr := fmt.Errorf("Value: "+validation.ErrorNonInt, "non-int")

	var cases = []struct {
		Input  TestIntValidator
		Errors utils.Errors
	}{
		{TestIntValidator{0}, utils.Errors{minErr}},
		{TestIntValidator{validValue}, nil},
		{TestIntValidator{400}, utils.Errors{maxErr}},
		{TestIntValidator{"non-int"}, utils.Errors{nonIntErr}},
	}

	for _, testCase := range cases {
		if errs := validation.Validate(testCase.Input); errs != nil {
			if testCase.Errors == nil {
				t.Errorf("For input: %v, expected errs to be nil: errs = %v\n", testCase.Input, errs)
			} else if errs.Error() != testCase.Errors.Error() {
				t.Errorf("For input: %v, got %v, want %v\n", testCase.Input, errs, testCase.Errors)
			}
		} else if testCase.Errors != nil {
			t.Errorf("For input: %v, expected errs to not be nil\n", testCase.Input)
		}
	}

}
