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

package graph

import (
	"errors"
	"fmt"
)

var (
	// ErrNoResultsFound may be returned when the result set does not contain a result matching the query specifications.
	ErrNoResultsFound = errors.New("not found")

	// ErrMissingResultExpectation may be returned when the result set does not adhere to driver expectations. For
	// example when a query result does not contain an expected value or variable.
	ErrMissingResultExpectation = errors.New("missing result expectation")

	// ErrUnsupportedDatabaseOperation may be returned to signal to a user that the DAWGS driver they are using does not
	// support the operation they are attempting to execute. This error should be used sparingly. All DAWGS drivers
	// should strive to satisfy all DAWGS contracts even if the resulting implementation is non-optimal.
	ErrUnsupportedDatabaseOperation = errors.New("unsupported database operation")

	// ErrPropertyNotFound is returned when a node or relationship property is found to be nil during type negotiation.
	ErrPropertyNotFound = errors.New("property not found")

	// ErrContextTimedOut is used to mark that an operation was halted due to the context hitting its deadline
	ErrContextTimedOut = errors.New("context timed out")
)

func IsErrNotFound(err error) bool {
	return errors.Is(err, ErrNoResultsFound)
}

func IsErrPropertyNotFound(err error) bool {
	return errors.Is(err, ErrPropertyNotFound)
}

func IsMissingResultExpectation(err error) bool {
	return errors.Is(err, ErrMissingResultExpectation)
}

// NewError returns an error that contains the given query context elements.
func NewError(query string, driverErr error) error {
	return fmt.Errorf("driver error: %w - query: %s", driverErr, query)
}
