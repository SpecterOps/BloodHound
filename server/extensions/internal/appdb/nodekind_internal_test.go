package appdb

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestCheckNodeKindError(t *testing.T) {
	t.Parallel()

	type args struct {
		err  error
		name string
	}

	type testCase struct {
		name        string
		args        args
		wantErr     error
		wantMessage string
	}

	var (
		databaseError  = errors.New("db error")
		duplicateError = fmt.Errorf("creating node kind: %w", &pgconn.PgError{
			Code:           PostgresUniqueViolationCode,
			ConstraintName: graphSchemaNodeKindConstraintUniqueKindID,
		})
		extensionError = fmt.Errorf("creating node kind: %w", &pgconn.PgError{
			Code:           PostgresForeignKeyViolationCode,
			ConstraintName: graphSchemaNodeKindConstraintExtensionIDFKey,
		})
		unrelatedUniqueError = &pgconn.PgError{
			Code:           PostgresUniqueViolationCode,
			ConstraintName: "schema_node_kinds_pkey",
		}
	)

	tests := []testCase{
		{
			name:        "error_-_maps_duplicate_name",
			args:        args{err: duplicateError, name: "User"},
			wantErr:     services.ErrDuplicateSchemaNodeKindName,
			wantMessage: "duplicate schema node kind name: User",
		},
		{
			name:        "error_-_maps_missing_extension",
			args:        args{err: extensionError, name: "User"},
			wantErr:     services.ErrExtensionNotFound,
			wantMessage: "extension not found: " + extensionError.Error(),
		},
		{
			name:        "error_-_preserves_database_error",
			args:        args{err: databaseError, name: "User"},
			wantErr:     databaseError,
			wantMessage: databaseError.Error(),
		},
		{
			name:        "error_-_preserves_unrelated_unique_error",
			args:        args{err: unrelatedUniqueError, name: "User"},
			wantErr:     unrelatedUniqueError,
			wantMessage: unrelatedUniqueError.Error(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := checkNodeKindError(test.args.err, test.args.name)
			assert.ErrorIs(t, err, test.wantErr)
			assert.EqualError(t, err, test.wantMessage)
		})
	}
}
