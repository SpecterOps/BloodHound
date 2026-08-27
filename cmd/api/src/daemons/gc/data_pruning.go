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
	"sync"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/server/audit"
)

// defaultAuditRetentionMonths bounds how many months of audit_logs partitions
// are retained. Partitions whose entire range is older than this window are
// dropped.
// TODO(audit retention): source this from appcfg once the
// retention_months parameter lands (Step I) instead of the constant.
const defaultAuditRetentionMonths = 3

// Daemon holds data relevant to the data daemon
type Daemon struct {
	stopC           chan struct{}
	doneC           chan struct{}
	stopOnce        sync.Once
	db              database.Database
	auditMaintainer audit.Maintainer
}

// NewDataPruningDaemon creates a new data pruning daemon. auditMaintainer manages
// the audit_logs range partitions and may be nil, in which case partition
// maintenance is skipped.
func NewDataPruningDaemon(db database.Database, auditMaintainer audit.Maintainer) *Daemon {
	return &Daemon{
		stopC:           make(chan struct{}),
		doneC:           make(chan struct{}),
		db:              db,
		auditMaintainer: auditMaintainer,
	}
}

// Name returns the name of the daemon
func (s *Daemon) Name() string {
	return "Data Pruning Daemon"
}

// Start begins the daemon and runs until Stop is called or ctx is canceled,
// closing doneC on exit so Stop can wait for the run loop to finish.
func (s *Daemon) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)

	defer close(s.doneC)
	defer ticker.Stop()

	// prune once when the daemon starts up
	s.runDataPruning(ctx)

	// thereafter, prune conditionally once a day
	for {
		select {
		case <-ticker.C:
			s.runDataPruning(ctx)

		case <-s.stopC:
			return

		case <-ctx.Done():
			return
		}
	}
}

// runDataPruning prunes sessions and asset group collections and maintains the
// audit_logs partitions. It is the single entry point for a pruning pass so the
// startup and ticker call sites stay in lockstep.
func (s *Daemon) runDataPruning(ctx context.Context) {
	s.db.SweepSessions(ctx)
	s.db.SweepAssetGroupCollections(ctx)
	s.sweepAuditPartitions(ctx)
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
	if err := s.auditMaintainer.CreateNextPartition(ctx, now); err != nil {
		slog.ErrorContext(ctx, "Failed to pre-create next audit partition", attr.Error(err))
	}
	if err := s.auditMaintainer.DropExpiredPartitions(ctx, now, defaultAuditRetentionMonths); err != nil {
		slog.ErrorContext(ctx, "Failed to drop expired audit partitions", attr.Error(err))
	}
}

// Stop signals the run loop to exit and waits for it to finish, returning early
// if ctx is canceled before the daemon completes.
func (s *Daemon) Stop(ctx context.Context) error {
	s.stopOnce.Do(func() {
		close(s.stopC)
	})

	select {
	case <-s.doneC:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
