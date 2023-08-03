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

package frontend

import (
	"errors"
	"fmt"
)

type SyntaxError struct {
	Line            int
	Column          int
	OffendingSymbol any
	Message         string
}

func (s SyntaxError) Error() string {
	return fmt.Sprintf("line %d:%d %s", s.Line, s.Column, s.Message)
}

var (
	ErrUpdateClauseNotSupported            = errors.New("updating clauses are not supported")
	ErrUserSpecifiedParametersNotSupported = errors.New("user-specified parameters are not supported")
	ErrProcedureInvocationNotSupported     = errors.New("procedure invocation is not supported")

	ErrInvalidInput = errors.New("invalid input")
)
