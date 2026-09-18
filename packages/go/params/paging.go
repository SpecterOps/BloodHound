// Copyright 2026 Specter Ops, Inc.
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

package params

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// pagingParameterLimit is the query parameter name recognized for the limit value. The skip parameter
// name is caller-configurable via PagingConfig.SkipKey.
const pagingParameterLimit = "limit"

// UnlimitedResults is the sentinel limit value a caller may supply to request an unbounded result
// set. It is honored only when the resource owner opts in via PagingConfig.AllowUnlimited.
const UnlimitedResults = -1

// Paging configuration fallbacks applied by NewPagingParser when the corresponding PagingConfig field
// is unset.
const (
	DefaultLimit   = 100
	defaultSkipKey = "skip"
)

// Paging validation sentinels. Callers classify failures with errors.Is and may extract the offending
// parameter and raw value via a *PagingValidationError using errors.AsType. These are intentionally free of
// any transport- or presentation-specific wording.
var (
	ErrInvalidInteger = errors.New("must be a non-negative integer")
)

// PagingConfig declares how a resource owner wants paging parameters parsed and validated. Zero-valued
// fields fall back to defaultLimit, defaultSkipKey, and disallowing unlimited results.
type PagingConfig struct {
	// DefaultLimit is the limit applied when the request does not supply one.
	DefaultLimit int
	// SkipKey is the query parameter name read for the skip value.
	SkipKey string
	// AllowUnlimited permits a request to supply UnlimitedResults (-1) to request an unbounded result set.
	AllowUnlimited bool
}

// PagingValidationError describes a paging validation failure. It wraps one of the validation sentinels
// so it can be classified with errors.Is, while also carrying the offending parameter name and raw value
// so callers can build their own messaging without parsing strings.
type PagingValidationError struct {
	Err       error
	Parameter string
	Value     string
}

func (s *PagingValidationError) Error() string {
	return fmt.Sprintf("%s: %s %q", s.Err, s.Parameter, s.Value)
}

func (s *PagingValidationError) Unwrap() error {
	return s.Err
}

// Paging is the storage-agnostic result of parsing the skip and limit query parameters. An absent limit
// yields the configured default; a Limit of UnlimitedResults means the caller requested an unbounded
// result set and the configuration allowed it.
type Paging struct {
	Skip  int
	Limit int
}

// PagingParser extracts paging values from request query parameters according to its configuration.
type PagingParser struct {
	config PagingConfig
}

// NewPagingParser returns a parser for the skip and limit query parameters, applying fallbacks for any
// unset PagingConfig fields: defaultLimit 100, SkipKey "skip", AllowUnlimited false.
func NewPagingParser(config PagingConfig) PagingParser {
	if config.DefaultLimit == 0 {
		config.DefaultLimit = DefaultLimit
	}

	if config.SkipKey == "" {
		config.SkipKey = defaultSkipKey
	}

	return PagingParser{
		config: config,
	}
}

// ParseAndValidate parses the skip and limit query parameters and validates them in a single pass. The
// skip value is read from the configured skip key; when the request does not contain that key, the
// parser checks the request for defaultSkipKey instead. An absent skip yields zero and an absent limit
// yields the configured default. Skip must be a non-negative integer; limit must be a non-negative
// integer, or UnlimitedResults when the configuration allows it. On failure it returns a
// *PagingValidationError wrapping one of the validation sentinels.
func (s PagingParser) ParseAndValidate(values url.Values) (Paging, error) {
	var (
		paging  = Paging{Limit: s.config.DefaultLimit}
		skipKey = s.config.SkipKey
	)

	if !values.Has(skipKey) {
		skipKey = defaultSkipKey
	}

	if rawSkip := values.Get(skipKey); rawSkip != "" {
		if skip, err := strconv.Atoi(rawSkip); err != nil || skip < 0 {
			return Paging{}, &PagingValidationError{Err: ErrInvalidInteger, Parameter: skipKey, Value: rawSkip}
		} else {
			paging.Skip = skip
		}
	}

	if rawLimit := values.Get(pagingParameterLimit); rawLimit != "" {
		if limit, err := strconv.Atoi(rawLimit); err != nil {
			return Paging{}, &PagingValidationError{Err: ErrInvalidInteger, Parameter: pagingParameterLimit, Value: rawLimit}
		} else if limit == UnlimitedResults && s.config.AllowUnlimited {
			paging.Limit = limit
		} else if limit < 0 {
			return Paging{}, &PagingValidationError{Err: ErrInvalidInteger, Parameter: pagingParameterLimit, Value: rawLimit}
		} else {
			paging.Limit = limit
		}
	}

	return paging, nil
}
