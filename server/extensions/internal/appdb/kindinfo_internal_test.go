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
package appdb

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestCheckKindInfoError(t *testing.T) {
	t.Parallel()

	type args struct {
		err error
	}

	type testCase struct {
		name    string
		args    args
		wantErr error
	}

	var (
		databaseError    = errors.New("database error")
		missingKindError = fmt.Errorf("database: %w", &pgconn.PgError{
			Code:           "23503",
			ConstraintName: graphSchemaKindInfoConstraintKindIDFKey,
		})
		duplicatePositionError = &pgconn.PgError{
			Code:           "23505",
			ConstraintName: graphSchemaKindInfoConstraintUniquePosition,
		}
		duplicateInfoKeyError = &pgconn.PgError{
			Code:           "23505",
			ConstraintName: graphSchemaKindInfoConstraintUniqueInfoKey,
		}
		unmappedConstraintError = &pgconn.PgError{
			Code:           "23514",
			ConstraintName: "unknown_constraint",
		}
	)

	tests := []testCase{
		{
			name:    "error_-_maps_missing_kind",
			args:    args{err: missingKindError},
			wantErr: services.ErrKindInfoKindNotFound,
		},
		{
			name:    "error_-_maps_duplicate_position",
			args:    args{err: duplicatePositionError},
			wantErr: services.ErrKindInfoDuplicatePosition,
		},
		{
			name:    "error_-_maps_duplicate_info_key",
			args:    args{err: duplicateInfoKeyError},
			wantErr: services.ErrKindInfoDuplicateInfoKey,
		},
		{
			name:    "error_-_preserves_database_error",
			args:    args{err: databaseError},
			wantErr: databaseError,
		},
		{
			name:    "error_-_preserves_unmapped_constraint",
			args:    args{err: unmappedConstraintError},
			wantErr: unmappedConstraintError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := checkKindInfoError(test.args.err)
			assert.ErrorIs(t, err, test.wantErr)
			assert.ErrorIs(t, err, test.args.err)
		})
	}
}
