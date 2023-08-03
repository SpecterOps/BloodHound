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

package validation

import (
	"reflect"

	"github.com/specterops/bloodhound/src/utils"
)

type requiredError struct {
	value any
}

func (s requiredError) Error() string {
	return "property is required and must not be empty or zero in value"
}

func NewRequiredError(value any) error {
	return requiredError{value}
}

type RequiredValidator struct{}

func (s RequiredValidator) Validate(value any) []error {
	errs := utils.Errors{}
	if reflect.ValueOf(value).IsZero() {
		errs = append(errs, NewRequiredError(value))
		return errs
	} else {
		return nil
	}
}

func NewRequiredValidator(params map[string]string) Validator {
	return RequiredValidator{}
}
