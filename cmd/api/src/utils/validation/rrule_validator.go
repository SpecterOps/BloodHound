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

	//Validate that the rrule is a good rule. We're going to require a DTSTART to keep scheduling consistent.
	//We're also going to reject UNTIL/COUNT because it will most likely break the pipeline once it's hit without being invalid
	if rruleStr == "" {
		return nil
	} else if _, err := rrule.StrToRRule(rruleStr); err != nil {
		return append(errs, fmt.Errorf(ErrInvalidRrule, err))
	} else if strings.Contains(strings.ToUpper(rruleStr), "UNTIL") {
		return append(errs, fmt.Errorf(ErrInvalidRrule, "until not supported"))
	} else if strings.Contains(strings.ToUpper(rruleStr), "COUNT") {
		return append(errs, fmt.Errorf(ErrInvalidRrule, "count not supported"))
	} else if !strings.Contains(strings.ToUpper(rruleStr), "DTSTART") {
		return append(errs, fmt.Errorf(ErrInvalidRrule, "dtstart is required"))
	}

	return nil
}
