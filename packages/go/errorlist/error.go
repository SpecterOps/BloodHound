// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package errorlist

import (
	"errors"
	"strings"
)

func NewBuilder() ErrorBuilder {
	return ErrorBuilder{}
}

type ErrorBuilder struct {
	Errors []error
}

func (s *ErrorBuilder) Add(e error) {
	var graphifyError Error
	if ok := errors.As(e, &graphifyError); ok {
		s.Errors = append(s.Errors, graphifyError.Errors...)
	} else if e != nil {
		s.Errors = append(s.Errors, e)
	}
}

func (s ErrorBuilder) Build() error {
	if len(s.Errors) == 0 {
		return nil
	} else {
		return Error(s)
	}
}

type Error struct {
	Errors []error
}

func (s Error) AsStrings() []string {
	errStrings := make([]string, len(s.Errors))

	for i, err := range s.Errors {
		errStrings[i] = err.Error()
	}

	return errStrings
}

func (s Error) Error() string {
	return strings.Join(s.AsStrings(), "; ")
}
