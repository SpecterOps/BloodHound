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

// Package audit provides shared audit-log behavior for feature modules.
package audit

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/server/audit/internal/appdb"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

type Entry = services.AuditLogEntry
type Action = services.Action

const (
	ActionUpdateParameter              Action = "UpdateParameter"
	ActionToggleEarlyAccessFeatureFlag Action = "ToggleEarlyAccessFeatureFlag"
)

var ErrAuditLog = services.ErrAuditLog

type auditService interface {
	Begin(ctx context.Context, entry Entry) (uuid.UUID, error)
	Complete(ctx context.Context, commitID uuid.UUID, entry Entry, operationErr error) error
}

// AuditLogAdapter exposes audit logging to consuming feature slices without
// exposing the audit slice's internal service implementation.
type AuditLogAdapter struct {
	service auditService
}

// NewAuditLogAdapter constructs the audit module's store and service layers.
func NewAuditLogAdapter(pool *pgxpool.Pool, identityResolver auth.IdentityResolver) *AuditLogAdapter {
	if pool == nil {
		panic("audit: NewAuditLogAdapter requires a non-nil Pool")
	}
	if identityResolver == nil {
		panic("audit: NewAuditLogAdapter requires a non-nil IdentityResolver")
	}

	return &AuditLogAdapter{service: services.NewService(appdb.NewStore(pool), identityResolver)}
}

func (s *AuditLogAdapter) Begin(ctx context.Context, entry Entry) (uuid.UUID, error) {
	return s.service.Begin(ctx, entry)
}

func (s *AuditLogAdapter) Complete(ctx context.Context, commitID uuid.UUID, entry Entry, operationErr error) error {
	return s.service.Complete(ctx, commitID, entry, operationErr)
}
