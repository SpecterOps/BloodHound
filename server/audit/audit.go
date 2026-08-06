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

// Package audit is the wireup module for the audit feature. It is the single
// place where the audit store and service are composed; the layered
// sub-packages themselves remain unaware of each other. Consumers (middleware,
// the module registry, the GC daemon) depend on this public package rather than
// the internal sub-packages.
package audit

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/audit/internal/appdb"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

// Entry is the domain input callers hand to the audit service for a single
// audited action. It is re-exported from the internal services package so
// consumers depend only on this public package.
type Entry = services.Entry

// Service records the intent/success/failure lifecycle of an audited action.
type Service = services.Service

// Maintainer manages the lifecycle of the audit_logs range partitions. The GC
// daemon depends on this port to pre-create upcoming partitions and drop
// partitions older than the configured retention window.
type Maintainer = services.Maintainer

// Register composes the audit store and service against the provided PostgreSQL
// connection pool. It returns the Service used to record audit entries and the
// Maintainer used by the GC daemon to manage audit partitions. The concrete
// Store satisfies both ports.
func Register(pool *pgxpool.Pool) (*Service, Maintainer) {
	var (
		store   = appdb.NewStore(pool)
		service = services.NewService(store)
	)

	return service, store
}
