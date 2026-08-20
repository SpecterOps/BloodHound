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

// Package audit is the wireup module for the audit feature: the single place
// where the store and service are composed. Consumers depend on this public
// package rather than the internal sub-packages.
package audit

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/audit/internal/appdb"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

// Entry is the domain input callers hand to the audit service for a single
// audited action.
type Entry = services.Entry

// Service records the intent/success/failure lifecycle of an audited action.
type Service = services.Service

// Maintainer manages the lifecycle of the audit_logs range partitions, used by
// the GC daemon to pre-create upcoming partitions and drop expired ones.
type Maintainer = services.Maintainer

// Register composes the audit store and service against the provided connection
// pool, returning the Service and the Maintainer (both satisfied by the Store).
func Register(pool *pgxpool.Pool) (*Service, Maintainer) {
	var (
		store   = appdb.NewStore(pool)
		service = services.NewService(store)
	)

	return service, store
}
