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

//go:generate go tool mockery

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/server/audit/internal/appdb"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

type Entry = services.Entry

// Service records the intent and outcome rows that make up one legacy audit event.
type Service interface {
	Begin(ctx context.Context, entry Entry) (uuid.UUID, error)
	Complete(ctx context.Context, commitID uuid.UUID, entry Entry, operationErr error) error
}

// Register constructs the audit module's store and service layers.
func Register(pool *pgxpool.Pool, identityResolver auth.IdentityResolver) Service {
	if pool == nil {
		panic("audit: Register requires a non-nil Pool")
	}
	if identityResolver == nil {
		panic("audit: Register requires a non-nil IdentityResolver")
	}

	return services.NewService(appdb.NewStore(pool), identityResolver)
}
