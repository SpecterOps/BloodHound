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

package gc

import (
	"context"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/server/audit"
)

// defaultAuditRetentionMonths bounds how many months of audit_logs partitions
// are retained. Partitions whose entire range is older than this window are
// dropped. TODO(audit retention): source this from appcfg once the
// retention_months parameter lands (Step I) instead of the constant.
const defaultAuditRetentionMonths = 3

// Daemon holds data relevant to the data daemon
type Daemon struct {
	exitC           chan struct{}
	db              database.Database
	auditMaintainer audit.Maintainer
}

// NewDataPruningDaemon creates a new data pruning daemon. auditMaintainer manages
// the audit_logs range partitions and may be nil, in which case partition
// maintenance is skipped.
func NewDataPruningDaemon(db database.Database, auditMaintainer audit.Maintainer) *Daemon {
	return &Daemon{
		exitC:           make(chan struct{}),
		db:              db,
		auditMaintainer: auditMaintainer,
	}
}

// Name returns the name of the daemon
func (s *Daemon) Name() string {
	return "Data Pruning Daemon"
}

// Start begins the daemon and waits for a stop signal in the exit channel
func (s *Daemon) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)

	defer close(s.exitC)
	defer ticker.Stop()

	// prune sessions and collections and maintain audit partitions once when the daemon starts up
	s.db.SweepSessions(ctx)
	s.db.SweepAssetGroupCollections(ctx)
	s.sweepAuditPartitions(ctx)

	// thereafter, prune conditionally once a day
	for {
		select {
		case <-ticker.C:
			s.db.SweepSessions(ctx)
			s.db.SweepAssetGroupCollections(ctx)
			s.sweepAuditPartitions(ctx)

		case <-s.exitC:
			return
		}
	}
}

// sweepAuditPartitions pre-creates the upcoming audit_logs partition and drops
// partitions older than the retention window. Failures are logged and do not
// stop the daemon; the next tick retries. It is a no-op when no maintainer is
// configured.
func (s *Daemon) sweepAuditPartitions(ctx context.Context) {
	if s.auditMaintainer == nil {
		return
	}

	now := time.Now().UTC()
	if err := s.auditMaintainer.PreCreateNextPartition(ctx, now); err != nil {
		slog.ErrorContext(ctx, "Failed to pre-create next audit partition", attr.Error(err))
	}
	if err := s.auditMaintainer.DropExpiredPartitions(ctx, now, defaultAuditRetentionMonths); err != nil {
		slog.ErrorContext(ctx, "Failed to drop expired audit partitions", attr.Error(err))
	}
}

// Stop passes in a stop signal to the exit channel, thereby killing the daemon
func (s *Daemon) Stop(ctx context.Context) error {
	s.exitC <- struct{}{}

	select {
	case <-s.exitC:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
