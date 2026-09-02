// Copyright 2025 Specter Ops, Inc.
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
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/teambition/rrule-go"
)

const (
	ErrInvalidRrule = "invalid rrule specified: %s"
)

// type RRuleValidator struct implements the Validator interface which allows the usage of the `rrule` struct tag
// to ensure a string is in a valid rrule format by calling `Validate`
type RRuleValidator struct {
}

// NewRRuleValidator returns a new Validator
func NewRRuleValidator(_ map[string]string) Validator {
	return RRuleValidator{}
}

func (s RRuleValidator) Validate(value any) utils.Errors {
	var (
		rruleStr string
		ok       bool
		errs     = utils.Errors{}
	)

	if rruleStr, ok = value.(string); !ok {
		return append(errs, fmt.Errorf(ErrInvalidRrule, value))
	}

	if rruleStr == "" {
		return nil
	}

	if _, err := ValidateRRule(rruleStr); err != nil {
		return append(errs, err)
	}

	return nil
}

// ValidateRRule validates that the rrule string is parseable and conforms to scheduling constraints.
// Valid rules require a DTSTART to keep scheduling consistent.
// Valid rules must not contain UNTIL or COUNT because they will break the pipeline once exhausted.
func ValidateRRule(rruleStr string) (*rrule.RRule, error) {
	var upperRRule = strings.ToUpper(rruleStr)

	if rule, err := rrule.StrToRRule(rruleStr); err != nil {
		return nil, fmt.Errorf(ErrInvalidRrule, err)
	} else if strings.Contains(upperRRule, "UNTIL") {
		return nil, fmt.Errorf(ErrInvalidRrule, "until not supported")
	} else if strings.Contains(upperRRule, "COUNT") {
		return nil, fmt.Errorf(ErrInvalidRrule, "count not supported")
	} else if !strings.Contains(upperRRule, "DTSTART") {
		return nil, fmt.Errorf(ErrInvalidRrule, "dtstart is required")
	} else {
		return rule, nil
	}
}
