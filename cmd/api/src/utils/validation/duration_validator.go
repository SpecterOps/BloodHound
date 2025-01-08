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

package validation

import (
	"fmt"
	"time"

	iso8601 "github.com/channelmeter/iso8601duration"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/utils"
)

const (
	ErrorDuration    = "invalid iso duration provided %v"
	ErrorDurationMin = "must be >= %s"
	ErrorDurationMax = "must be <= %s"
)

type DurationValidator struct {
	min, max   string // Used for consistent error output
	minD, maxD time.Duration
}

func NewDurationValidator(params map[string]string) Validator {
	validator := DurationValidator{}

	if minD, ok := params["min"]; ok {
		validator.min = params["min"]
		if duration, err := iso8601.FromString(minD); err != nil {
			log.Warnf(fmt.Sprintf("NewDurationValidator invalid min limit provided %s", minD))
		} else {
			validator.minD = duration.ToDuration()
		}
	}

	if maxD, ok := params["max"]; ok {
		validator.max = params["max"]
		if duration, err := iso8601.FromString(maxD); err != nil {
			log.Warnf(fmt.Sprintf("NewDurationValidator invalid max limit provided %s", maxD))
		} else {
			validator.maxD = duration.ToDuration()
		}
	}

	return validator
}

func (s DurationValidator) okMin(d time.Duration) bool {
	return d >= s.minD
}

func (s DurationValidator) okMax(d time.Duration) bool {
	return d <= s.maxD
}

func (s DurationValidator) Validate(value any) utils.Errors {
	var (
		d    time.Duration
		ok   bool
		errs = utils.Errors{}
	)
	if d, ok = value.(time.Duration); !ok {
		return append(errs, fmt.Errorf(ErrorDuration, value))
	}

	if s.minD > 0 && !s.okMin(d) {
		errs = append(errs, fmt.Errorf(ErrorDurationMin, s.min))
	}

	if s.maxD > 0 && !s.okMax(d) {
		errs = append(errs, fmt.Errorf(ErrorDurationMax, s.max))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}
