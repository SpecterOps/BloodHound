// Copyright 2023 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/src/utils/validation"
)

func TestPasswordValidator(t *testing.T) {

	type Foo struct {
		Password string `validate:"password,length=12,lower=2,upper=2,special=2,numeric=2"`
	}

	lengthErr := fmt.Errorf("Password: "+validation.ErrorPasswordLength, 12)
	lowerErr := fmt.Errorf("Password: "+validation.ErrorPasswordLower, 2)
	upperErr := fmt.Errorf("Password: "+validation.ErrorPasswordUpper, 2)
	specialErr := fmt.Errorf("Password: "+validation.ErrorPasswordSpecial, 2)
	numericErr := fmt.Errorf("Password: "+validation.ErrorPasswordNumeric, 2)

	var cases = []struct {
		Input  Foo
		Errors []error
	}{
		{Foo{}, []error{lengthErr, lowerErr, upperErr, specialErr, numericErr}[:]},
		{Foo{"abcDEF***123"}, nil},
		{Foo{"abcDEF***aaa"}, []error{numericErr}[:]},
	}

	for _, tc := range cases {
		if errs := validation.Validate(tc.Input); errs != nil {
			if tc.Errors == nil {
				t.Errorf("For input: %v, expected errs to be nil: errs = %v\n", tc.Input, errs)
			} else if len(errs) != len(tc.Errors) {
				t.Errorf("For input: %v, got %v, want %v\n", tc.Input, errs, tc.Errors)
			}
		} else if tc.Errors != nil {
			t.Errorf("For input: %v, expected errs to not be nil\n", tc.Input)
		}
	}
}
