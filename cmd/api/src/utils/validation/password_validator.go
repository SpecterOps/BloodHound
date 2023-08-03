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
	"fmt"
	"strconv"
	"unicode"

	"github.com/specterops/bloodhound/src/utils"
)

const (
	ErrorPassword        = "failed to meet password requirements:\n%v"
	ErrorPasswordLength  = "must have at least %d characters"
	ErrorPasswordLower   = "must have at least %d lowercase characters"
	ErrorPasswordUpper   = "must have at least %d uppercase characters"
	ErrorPasswordSpecial = "must have at least %d special characters"
	ErrorPasswordNumeric = "must have at least %d numeric characters"
)

type PasswordValidator struct {
	Length, Lower, Upper, Special, Numeric int
}

func NewPasswordValidator(params map[string]string) Validator {
	validator := PasswordValidator{}

	if length, ok := params["length"]; ok {
		validator.Length, _ = strconv.Atoi(length)
	}

	if lower, ok := params["lower"]; ok {
		validator.Lower, _ = strconv.Atoi(lower)
	}

	if upper, ok := params["upper"]; ok {
		validator.Upper, _ = strconv.Atoi(upper)
	}

	if special, ok := params["special"]; ok {
		validator.Special, _ = strconv.Atoi(special)
	}

	if numeric, ok := params["numeric"]; ok {
		validator.Numeric, _ = strconv.Atoi(numeric)
	}

	return validator
}

func (s PasswordValidator) okLen(passwd string) bool {
	return len(passwd) >= s.Length
}

func (s PasswordValidator) okLower(count int) bool {
	return count >= s.Lower
}

func (s PasswordValidator) okUpper(count int) bool {
	return count >= s.Upper
}

func (s PasswordValidator) okSpecial(count int) bool {
	return count >= s.Special
}

func (s PasswordValidator) okNumeric(count int) bool {
	return count >= s.Numeric
}

func (s PasswordValidator) ok(lower, upper, special, numeric int) bool {
	return s.okLower(lower) && s.okUpper(upper) && s.okSpecial(special) && s.okNumeric(numeric)
}

func (s PasswordValidator) Validate(value any) []error {
	var countLower, countUpper, countSpecial, countNumeric int
	passwd := value.(string)
	errs := utils.Errors{}

	for _, char := range passwd {
		if s.ok(countLower, countUpper, countSpecial, countNumeric) {
			break
		} else {
			switch {
			case unicode.IsLower(char):
				countLower++
			case unicode.IsUpper(char):
				countUpper++
			case unicode.IsPunct(char) || unicode.IsSymbol(char):
				countSpecial++
			case unicode.IsNumber(char):
				countNumeric++
			}
		}
	}

	if !s.okLen(passwd) {
		errs = append(errs, fmt.Errorf(ErrorPasswordLength, s.Length))
	}

	if !s.okLower(countLower) {
		errs = append(errs, fmt.Errorf(ErrorPasswordLower, s.Lower))
	}

	if !s.okUpper(countUpper) {
		errs = append(errs, fmt.Errorf(ErrorPasswordUpper, s.Upper))
	}

	if !s.okSpecial(countSpecial) {
		errs = append(errs, fmt.Errorf(ErrorPasswordSpecial, s.Special))
	}

	if !s.okNumeric(countNumeric) {
		errs = append(errs, fmt.Errorf(ErrorPasswordNumeric, s.Numeric))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
