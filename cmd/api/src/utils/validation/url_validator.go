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
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/utils"
)

const (
	ErrorUrlHttpsInvalid = "invalid https url format"
	ErrorUrlInvalid      = "invalid url format"
)

// UrlValidator implements the Validator interface which allows the usage of the `url` struct tag
// to ensure a string is in a valid url format by calling `Validate`
type UrlValidator struct {
	httpsOnly bool
}

// NewUrlValidator returns a new Validator
func NewUrlValidator(params map[string]string) Validator {
	validator := UrlValidator{}

	if val, ok := params["httpsOnly"]; ok {
		validator.httpsOnly, _ = strconv.ParseBool(val)
	}

	return validator
}

// Validate validates that the associated struct fields are in the proper formatting
func (s UrlValidator) Validate(value any) utils.Errors {
	var (
		errs         = utils.Errors{}
		inputUrl, ok = value.(string)
	)

	if !ok {
		errs = append(errs, fmt.Errorf("expected a string value, got %s", reflect.TypeOf(value)))
	}

	if s.httpsOnly {
		if !strings.HasPrefix(inputUrl, "https://") {
			errs = append(errs, errors.New(ErrorUrlHttpsInvalid))
		}
	}

	if err := ValidUrl(inputUrl); err != nil {
		errs = append(errs, errors.New(ErrorUrlInvalid))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func ValidUrl(inputUrl string) error {
	if _, err := url.ParseRequestURI(inputUrl); err != nil {
		return err
	}

	return nil
}
