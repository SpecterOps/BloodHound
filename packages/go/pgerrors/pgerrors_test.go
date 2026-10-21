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

package pgerrors_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/packages/go/pgerrors"
)

func TestViolationHelpers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		code    string
		matches func(error, ...string) bool
	}{
		{name: "unique", code: pgerrcode.UniqueViolation, matches: pgerrors.IsUniqueViolation},
		{name: "foreign key", code: pgerrcode.ForeignKeyViolation, matches: pgerrors.IsForeignKeyViolation},
		{name: "not null", code: pgerrcode.NotNullViolation, matches: pgerrors.IsNotNullViolation},
		{name: "check", code: pgerrcode.CheckViolation, matches: pgerrors.IsCheckViolation},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			postgresError := &pgconn.PgError{
				Code:           testCase.code,
				ConstraintName: "target_constraint",
			}
			wrappedError := fmt.Errorf("database operation failed: %w", postgresError)

			assertions := []struct {
				name            string
				err             error
				constraintNames []string
				want            bool
			}{
				{name: "direct violation", err: postgresError, want: true},
				{name: "wrapped violation", err: wrappedError, want: true},
				{name: "matching constraint", err: postgresError, constraintNames: []string{"target_constraint"}, want: true},
				{name: "one of multiple constraints", err: postgresError, constraintNames: []string{"other_constraint", "target_constraint"}, want: true},
				{name: "non-matching constraint", err: postgresError, constraintNames: []string{"other_constraint"}, want: false},
				{name: "non-matching code", err: &pgconn.PgError{Code: "08006", ConstraintName: "target_constraint"}, want: false},
				{name: "ordinary error", err: errors.New("ordinary error"), want: false},
				{name: "nil error", err: nil, want: false},
				{name: "typed nil PostgreSQL error", err: (*pgconn.PgError)(nil), want: false},
			}

			for _, assertion := range assertions {
				t.Run(assertion.name, func(t *testing.T) {
					t.Parallel()

					got := testCase.matches(assertion.err, assertion.constraintNames...)
					if got != assertion.want {
						t.Errorf("got %t, want %t", got, assertion.want)
					}
				})
			}
		})
	}
}
