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

// Package pgerrors provides helpers for identifying PostgreSQL constraint
// violations returned by pgx.
package pgerrors

import (
	"errors"
	"slices"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// IsUniqueViolation reports whether err is a PostgreSQL unique_violation.
// When constraintNames are provided, the error's constraint name must match
// one of them.
func IsUniqueViolation(err error, constraintNames ...string) bool {
	return isViolation(err, pgerrcode.UniqueViolation, constraintNames)
}

// IsForeignKeyViolation reports whether err is a PostgreSQL
// foreign_key_violation. When constraintNames are provided, the error's
// constraint name must match one of them.
func IsForeignKeyViolation(err error, constraintNames ...string) bool {
	return isViolation(err, pgerrcode.ForeignKeyViolation, constraintNames)
}

// IsNotNullViolation reports whether err is a PostgreSQL not_null_violation.
// When constraintNames are provided, the error's constraint name must match
// one of them.
func IsNotNullViolation(err error, constraintNames ...string) bool {
	return isViolation(err, pgerrcode.NotNullViolation, constraintNames)
}

// IsCheckViolation reports whether err is a PostgreSQL check_violation.
// When constraintNames are provided, the error's constraint name must match
// one of them.
func IsCheckViolation(err error, constraintNames ...string) bool {
	return isViolation(err, pgerrcode.CheckViolation, constraintNames)
}

func isViolation(err error, code string, constraintNames []string) bool {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) || postgresError == nil || postgresError.Code != code {
		return false
	}

	if len(constraintNames) == 0 {
		return true
	}

	return slices.Contains(constraintNames, postgresError.ConstraintName)
}
