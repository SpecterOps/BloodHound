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

package errors

import (
	goerrors "errors"
	"fmt"
)

// Error is a type alias that implements the error interface. This allows a user to assign error compatible values as
// constants in a package. Being an immutable constant makes the resulting error value more suitable as a sentinel
// error. There are some minor compiler optimization benefits to this model as well.
type Error string

// Error returns the string value of the error.
func (s Error) Error() string {
	return string(s)
}

// New is a type casting function for converting a string value into an Error value that implements the error interface.
func New(value string) error {
	return Error(value)
}

// Is reports whether any error in err's chain matches target.
func Is(err error, target error) bool {
	return goerrors.Is(err, target)
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
//
// The chain consists of err itself followed by the sequence of errors obtained by
// repeatedly calling Unwrap.
//
// An error matches target if the error's concrete value is assignable to the value
// pointed to by target, or if the error has a method As(interface{}) bool such that
// As(target) returns true. In the latter case, the As method is responsible for
// setting target.
//
// An error type might provide an As method so it can be treated as if it were a
// different error type.
//
// As panics if target is not a non-nil pointer to either a type that implements
// error, or to any interface type.
func As(err error, target any) bool {
	return goerrors.As(err, target)
}

// The ErrorCollector utilites are useful for aggregating errors across a multi-step
// process so that an early return is avoided and subsequent steps are allowed to execute.
// Ultimately, any errors that do happen can be returned as an aggregated list.
type ErrorCollector []error

func (s *ErrorCollector) Return() error {
	if s.HasErrors() {
		return s
	}

	return nil
}

func (s *ErrorCollector) HasErrors() bool {
	return s.Len() > 0
}

func (s *ErrorCollector) Len() int {
	return len(*s)
}

func (s *ErrorCollector) Collect(e error) { *s = append(*s, e) }

func (s *ErrorCollector) Error() string {
	err := "Collected errors:\n"
	for i, e := range *s {
		err += fmt.Sprintf("\tError %d: %s\n", i, e.Error())
	}

	return err
}
