// Copyright 2026 Specter Ops, Inc.
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

package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm"
)

const (
	samlConsumedIdentifiersTableName = "saml_consumed_identifiers"
	// these must match the saml_identifier_type enum in the saml_consumed_identifiers migration.
	samlIdentifierTypeResponse  = "response"
	samlIdentifierTypeAssertion = "assertion"
)

// SAMLConsumedData defines the methods for persisting and deleting SAML identifiers consumed during login,
// stored in the saml_consumed_identifiers table.
type SAMLConsumedData interface {
	CreateSAMLConsumedIdentifiers(ctx context.Context, ssoProviderID int32, idpIssuer, responseID, assertionID string, expiresAt time.Time) error
	SweepSAMLConsumedIdentifiers(ctx context.Context) error
}

// CreateSAMLConsumedIdentifiers inserts the SAMLResponse and assertion from a single SAML login so they cannot be replayed.
// Both identifiers are inserted together or not at all. If either identifier has already been consumed, it returns
// ErrSAMLIdentifierAlreadyConsumed and no records are written.
func (s *BloodhoundDB) CreateSAMLConsumedIdentifiers(ctx context.Context, ssoProviderID int32, idpIssuer, responseID, assertionID string,
	expiresAt time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(fmt.Sprintf(`INSERT INTO %s
			(sso_provider_id, idp_issuer, identifier_type, identifier, expires_at)
			VALUES
				(?, ?, ?, ?, ?),
				(?, ?, ?, ?, ?)
			ON CONFLICT DO NOTHING`, samlConsumedIdentifiersTableName),
			ssoProviderID, idpIssuer, samlIdentifierTypeResponse, responseID, expiresAt,
			ssoProviderID, idpIssuer, samlIdentifierTypeAssertion, assertionID, expiresAt,
		)

		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 2 {
			slog.WarnContext(
				ctx,
				"SAMLResponse or assertion already consumed (possible replay)",
				slog.Int("sso_provider_id", int(ssoProviderID)),
				slog.String("idp_issuer", idpIssuer),
			)
			return ErrSAMLIdentifierAlreadyConsumed
		}
		return nil
	})
}

// SweepSAMLConsumedIdentifiers deletes all SAMLResponse and assertion identifiers that have expired
func (s *BloodhoundDB) SweepSAMLConsumedIdentifiers(ctx context.Context) error {
	result := s.db.WithContext(ctx).Exec(fmt.Sprintf(`DELETE FROM %s WHERE expires_at < NOW()`,
		samlConsumedIdentifiersTableName))

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		slog.DebugContext(ctx, "SAML identifiers cleanup", slog.Int64("rows_affected", result.RowsAffected))
	}
	return nil
}
