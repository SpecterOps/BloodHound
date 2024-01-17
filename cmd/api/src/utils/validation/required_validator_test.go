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

func TestRequiredValidator(t *testing.T) {

	type Foo struct {
		String  string  `validate:"required"`
		Int     int     `validate:"required"`
		Pointer *string `validate:"required"`
	}

	var (
		nilPointer *string
		someString = "foo"
	)
	stringError := fmt.Errorf("String: %s", validation.NewRequiredError(""))
	intError := fmt.Errorf("Int: %s", validation.NewRequiredError(0))
	pointerError := fmt.Errorf("Pointer: %s", validation.NewRequiredError(nilPointer))

	var cases = []struct {
		Input  Foo
		Errors []error
	}{
		{Foo{}, []error{stringError, intError, pointerError}[:]},
		{Foo{someString, 1, &someString}, nil},
		{Foo{someString, 0, &someString}, []error{intError}[:]},
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
