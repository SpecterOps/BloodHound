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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/validator.go -package=mocks . Validator

import (
	"fmt"
	"reflect"
)

const tagName = "validate"

const (
	ErrorValidation   = "validation failed:\n%8s"
	ErrorUnmatchedTag = "no validator registered that matches tag: %s"
)

type Validator interface {
	Validate(value any) []error
}

func Validate(obj any) []error {
	value := reflect.ValueOf(obj)
	errs := []error{}

	for i := 0; i < value.NumField(); i++ {
		validatorTag := value.Type().Field(i).Tag.Get(tagName)

		if validatorTag == "" || validatorTag == "-" {
			continue
		} else if validator := validatorFactory.NewValidatorFromTag(validatorTag); validator == nil {
			errs = append(errs, fmt.Errorf(ErrorUnmatchedTag, validatorTag))
		} else if validationErrs := validator.Validate(value.Field(i).Interface()); validationErrs != nil {
			for _, e := range validationErrs {
				errs = append(errs, fmt.Errorf("%s: %s", value.Type().Field(i).Name, e.Error()))
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
