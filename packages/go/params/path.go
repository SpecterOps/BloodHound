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
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
)

// Path validation sentinels classify extraction and parsing failures with errors.Is.
var (
	ErrMissingPathParameter   = errors.New("path parameter is missing")
	ErrMalformedPathParameter = errors.New("path parameter is malformed")
)

// PathValidationError describes a path parameter failure. Err is either
// ErrMissingPathParameter or ErrMalformedPathParameter, identifiable with errors.Is.
// Callers can use errors.As to inspect Parameter and Value without parsing Error().
// Cause is nil for missing parameters and contains the original parser error for
// malformed parameters. Unwrap exposes Cause to errors.Is and errors.As; callers
// that include parser details in a response can use Cause.Error() directly.
// HTTP status codes and response messages remain the caller's responsibility.
type PathValidationError struct {
	Err       error
	Parameter string
	Value     string
	Cause     error
}

func (s *PathValidationError) Error() string {
	if s.Cause != nil {
		return fmt.Sprintf("%s: %s %q: %s", s.Err, s.Parameter, s.Value, s.Cause)
	}
	return fmt.Sprintf("%s: %s %q", s.Err, s.Parameter, s.Value)
}

// Is identifies the validation sentinel independently of the parser error.
func (s *PathValidationError) Is(target error) bool {
	return target == s.Err
}

// Unwrap preserves access to the original parser error.
func (s *PathValidationError) Unwrap() error {
	return s.Cause
}

// PathUUID parses a named Gorilla Mux path variable as a UUID, including uuid.Nil.
// An absent key returns ErrMissingPathParameter; an empty or invalid value returns
// ErrMalformedPathParameter. Failures return uuid.Nil and a *PathValidationError.
func PathUUID(request *http.Request, parameter string) (uuid.UUID, error) {
	var value, exists = mux.Vars(request)[parameter]

	if !exists {
		return uuid.Nil, &PathValidationError{Err: ErrMissingPathParameter, Parameter: parameter}
	}

	if identifier, err := uuid.FromString(value); err != nil {
		return uuid.Nil, &PathValidationError{Err: ErrMalformedPathParameter, Parameter: parameter, Value: value, Cause: err}
	} else {
		return identifier, nil
	}
}

// PathInt32 parses a named Gorilla Mux path variable as a signed decimal int32.
// Zero and negative values are accepted. Absent keys return ErrMissingPathParameter;
// empty, invalid, or out-of-range values return ErrMalformedPathParameter.
// Failures return zero and a *PathValidationError.
func PathInt32(request *http.Request, parameter string) (int32, error) {
	var value, err = pathInt(request, parameter, 32)
	return int32(value), err
}

// PathInt64 parses a named Gorilla Mux path variable as a signed decimal int64.
// Zero and negative values are accepted. Absent keys return ErrMissingPathParameter;
// empty, invalid, or out-of-range values return ErrMalformedPathParameter.
// Failures return zero and a *PathValidationError.
func PathInt64(request *http.Request, parameter string) (int64, error) {
	return pathInt(request, parameter, 64)
}

func pathInt(request *http.Request, parameter string, bitSize int) (int64, error) {
	var value, exists = mux.Vars(request)[parameter]

	if !exists {
		return 0, &PathValidationError{Err: ErrMissingPathParameter, Parameter: parameter}
	}

	if identifier, err := strconv.ParseInt(value, 10, bitSize); err != nil {
		return 0, &PathValidationError{Err: ErrMalformedPathParameter, Parameter: parameter, Value: value, Cause: err}
	} else {
		return identifier, nil
	}
}
