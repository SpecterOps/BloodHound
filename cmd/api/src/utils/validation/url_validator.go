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
	ErrorUrlHttpsInvalid     = "invalid https url format"
	ErrorUrlInvalid          = "invalid url format"
	ErrorURLHTTPSRequired    = "URL must use HTTPS"
	ErrorURLHostnameRequired = "URL must include a hostname"
)

// UrlValidator validates URLs through the `url` struct tag or directly.
// The optional hostnameRequired setting requires a URL with a hostname
type UrlValidator struct {
	httpsOnly        bool
	hostnameRequired bool
}

// UrlValidatorWithOptions constructs a URL validator and reports invalid options.
func UrlValidatorWithOptions(options map[string]string) (Validator, error) {
	var validator UrlValidator

	for optionName, optionValue := range options {
		var setting *bool

		switch optionName {
		case "httpsOnly":
			setting = &validator.httpsOnly
		case "hostnameRequired":
			setting = &validator.hostnameRequired
		default:
			return nil, fmt.Errorf("unknown URL validator option %q", optionName)
		}

		if enabled, err := strconv.ParseBool(optionValue); err != nil {
			return nil, fmt.Errorf("URL validator option %q must be a boolean", optionName)
		} else {
			*setting = enabled
		}
	}

	return validator, nil
}

type invalidValidator struct {
	err error
}

func (s invalidValidator) Validate(_ any) utils.Errors {
	return utils.Errors{s.err}
}

// NewUrlValidator adapts URL validator construction to the validator registry.
func NewUrlValidator(options map[string]string) Validator {
	if validator, err := UrlValidatorWithOptions(options); err != nil {
		return invalidValidator{err: err}
	} else {
		return validator
	}
}

// Validate validates that the associated struct fields are in the proper formatting
func (s UrlValidator) Validate(value any) utils.Errors {
	var (
		errs         utils.Errors
		inputUrl, ok = value.(string)
	)

	if !ok {
		errs = append(errs, fmt.Errorf("expected a string value, got %s", reflect.TypeOf(value)))
	}

	if s.hostnameRequired {
		if parsedURL, err := url.Parse(inputUrl); err != nil {
			errs = append(errs, errors.New(ErrorUrlInvalid))
		} else if s.httpsOnly && parsedURL.Scheme != "https" {
			errs = append(errs, errors.New(ErrorURLHTTPSRequired))
		} else if parsedURL.Hostname() == "" {
			errs = append(errs, errors.New(ErrorURLHostnameRequired))
		}
		return errs
	}

	// Preserve the request URI and HTTPS prefix checks for existing callers
	if s.httpsOnly && !strings.HasPrefix(inputUrl, "https://") {
		errs = append(errs, errors.New(ErrorUrlHttpsInvalid))
	}

	if err := ValidUrl(inputUrl); err != nil {
		errs = append(errs, errors.New(ErrorUrlInvalid))
	}

	return errs
}

func ValidUrl(inputUrl string) error {
	if _, err := url.ParseRequestURI(inputUrl); err != nil {
		return err
	}

	return nil
}
