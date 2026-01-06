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

package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/cypher/models/pgsql"
	"github.com/specterops/dawgs/cypher/models/pgsql/format"
	"gorm.io/gorm"
)

type sqlFilter struct {
	sqlString string
	params    []any
}

var ErrInvalidSortDirection = errors.New("invalid sort direction")

func CheckError(tx *gorm.DB) error {
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}

	return tx.Error
}

// NullUUID returns a uuid.NullUUID struct i.e a UUID that can be null in pg
func NullUUID(value uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{
		UUID:  value,
		Valid: true,
	}
}

// buildSQLFilter builds a PGSQL syntax-correct SQLFilter result from the given Filters struct. This function
// uses the PGSQL AST to ensure formatted SQL correctness.
// TODO: could still use more refactoring to reduce cyclomatic complexity, mostly hoisted from the old model implementation
func buildSQLFilter(filters model.Filters) (sqlFilter, error) {
	var (
		whereClauseFragment pgsql.Expression
		filter              sqlFilter
	)

	for name, filterOperations := range filters {
		var (
			formattedName = strings.TrimSpace(strings.ToLower(name))

			columnReference          pgsql.Expression = pgsql.Identifier(formattedName)
			innerWhereClauseFragment pgsql.Expression
			needsParenthetical       bool
		)

		for _, filter := range filterOperations {
			var (
				operator     pgsql.Operator
				filterValue  = filter.Value
				isNullValue  = filterValue == model.NullString
				literalValue pgsql.Literal
				setOperator  = pgsql.OperatorAnd
			)

			switch filter.Operator {
			case model.GreaterThan:
				operator = pgsql.OperatorGreaterThan

			case model.GreaterThanOrEquals:
				operator = pgsql.OperatorGreaterThanOrEqualTo

			case model.LessThan:
				operator = pgsql.OperatorLessThan

			case model.LessThanOrEquals:
				operator = pgsql.OperatorLessThanOrEqualTo

			case model.Equals:
				if isNullValue {
					operator = pgsql.OperatorIs
				} else {
					operator = pgsql.OperatorEquals
				}

			case model.NotEquals:
				if isNullValue {
					operator = pgsql.OperatorIsNot
				} else {
					operator = pgsql.OperatorNotEquals
				}

			case model.ApproximatelyEquals:
				operator = pgsql.OperatorLike
				filterValue = "%" + filterValue + "%"

			default:
				return sqlFilter{}, fmt.Errorf("invalid operator specified")
			}

			if isNullValue {
				literalValue = pgsql.NullLiteral()
			} else if valueInt64, err := strconv.ParseInt(filterValue, 10, 64); err == nil {
				literalValue = pgsql.NewLiteral(valueInt64, pgsql.Int8)
			} else if valueFloat64, err := strconv.ParseFloat(filterValue, 64); err == nil {
				literalValue = pgsql.NewLiteral(valueFloat64, pgsql.Float8)
			} else if valueBool, err := strconv.ParseBool(filterValue); err == nil {
				literalValue = pgsql.NewLiteral(valueBool, pgsql.Boolean)
			} else if val, err := pgsql.AsLiteral(filterValue); err != nil {
				return sqlFilter{}, fmt.Errorf("invalid filter value specified for %s: %w", name, err)
			} else {
				literalValue = val
			}

			if filter.SetOperator == model.FilterOr {
				needsParenthetical = true
				setOperator = pgsql.OperatorOr
			}

			if innerWhereClauseFragment == nil {
				innerWhereClauseFragment = pgsql.NewBinaryExpression(
					columnReference,
					operator,
					literalValue,
				)
			} else {
				innerWhereClauseFragment = pgsql.NewBinaryExpression(innerWhereClauseFragment, setOperator, pgsql.NewBinaryExpression(
					columnReference,
					operator,
					literalValue,
				))
			}
		}

		// OR statements need parens between AND clauses
		if needsParenthetical {
			innerWhereClauseFragment = pgsql.NewParenthetical(innerWhereClauseFragment).AsExpression()
		}
		whereClauseFragment = pgsql.OptionalAnd(whereClauseFragment, innerWhereClauseFragment)
	}

	if whereClauseFragment != nil {
		if sqlFragment, err := format.SyntaxNode(whereClauseFragment); err != nil {
			return filter, fmt.Errorf("failed formatting SQL filter: %w", err)
		} else {
			filter = sqlFilter{
				sqlString: sqlFragment,
			}
		}
	}

	return filter, nil
}

func buildSQLSort(sorts model.Sort) (string, error) {
	var (
		sortColumns []string
		orderSqlStr string
	)
	if len(sorts) > 0 {
		for _, item := range sorts {
			var dirString string
			switch item.Direction {
			case model.AscendingSortDirection:
				dirString = "ASC"
			case model.DescendingSortDirection:
				dirString = "DESC"
			case model.InvalidSortDirection:
				return "", ErrInvalidSortDirection
			}

			sortColumns = append(sortColumns, item.Column+" "+dirString)
		}
		orderSqlStr = "ORDER BY " + strings.Join(sortColumns, ", ")
	}

	return orderSqlStr, nil
}

type Transaction struct {
	tx         *gorm.DB
	auditEntry model.AuditEntry
}

// getTransaction - if t is not nil, use the transaction passed to us. Otherwise create a transaction from the context.
func (s *BloodhoundDB) getTransaction(ctx context.Context, t *Transaction) *gorm.DB {
	if t != nil && t.tx != nil {
		return t.tx
	}
	return s.db.WithContext(ctx)
}

func (s *BloodhoundDB) BeginTransaction(ctx context.Context) Transaction {
	var t Transaction
	t.tx = s.db.WithContext(ctx).Begin()
	return t
}

func (s *BloodhoundDB) BeginAuditableTransaction(ctx context.Context, auditEntry model.AuditEntry) (Transaction, error) {
	var (
		commitID, err = uuid.NewV4()
		t             Transaction
	)
	t.auditEntry = auditEntry

	if err != nil {
		return Transaction{}, fmt.Errorf("commitID could not be created: %w", err)
	}

	auditEntry.CommitID = commitID
	auditEntry.Status = model.AuditLogStatusIntent

	if err := s.AppendAuditLog(ctx, auditEntry); err != nil {
		return Transaction{}, fmt.Errorf("could not append intent to audit log: %w", err)
	}

	t.tx = s.db.WithContext(ctx).Begin()
	return t, nil
}

func (s *BloodhoundDB) CommitTransaction(ctx context.Context, t *Transaction) error {
	err := t.tx.Commit().Error

	// if we have an audit entry associated with this transaction, then complete the audit log result
	if t.auditEntry.CommitID != uuid.Nil {
		if err != nil {
			t.auditEntry.Status = model.AuditLogStatusFailure
			t.auditEntry.ErrorMsg = err.Error()
		} else {
			t.auditEntry.Status = model.AuditLogStatusSuccess
		}

		if err := s.AppendAuditLog(ctx, t.auditEntry); err != nil {
			return fmt.Errorf("could not append %s to audit log: %w", t.auditEntry.Status, err)
		}
	}

	return err
}

func (s *BloodhoundDB) Rollback(ctx context.Context, t *Transaction) error {
	err := t.tx.Rollback().Error

	// if we have an audit entry associated with this transaction, then complete the audit log result
	if err == nil && t.auditEntry.CommitID != uuid.Nil {
		t.auditEntry.Status = model.AuditLogStatusFailure
		t.auditEntry.ErrorMsg = "transaction rolled back"

		if err := s.AppendAuditLog(ctx, t.auditEntry); err != nil {
			return fmt.Errorf("could not append %s to audit log: %w", t.auditEntry.Status, err)
		}
	}

	return err
}
