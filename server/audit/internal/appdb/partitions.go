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
package appdb

import (
	"context"
	"fmt"
	"time"
)

const (
	partitionNameFormat = "2006_01"
	partitionDateFormat = "2006-01-02"
)

// earliestPartitionMonth is the first month for which the initial migration
// created partitions (see 20260707000001_v9_audit_log_partitioning.sql). It
// bounds the drop scan so we never loop unbounded looking for old partitions.
var earliestPartitionMonth = time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

// CreateNextPartition ensures the partition for the month AFTER asOf exists.
// Names and bounds mirror the migration; the DDL is injection-safe because every
// value derives from a time.Time.
func (s *Store) CreateNextPartition(ctx context.Context, asOf time.Time) error {
	var (
		next = firstOfMonth(asOf).AddDate(0, 1, 0)
		name = partitionName(next)
		ddl  = fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')`,
			name, tableAuditLogs,
			next.Format(partitionDateFormat),
			next.AddDate(0, 1, 0).Format(partitionDateFormat),
		)
		err error
	)
	if _, err = s.db.Exec(ctx, ddl); err != nil {
		return fmt.Errorf("creating audit partition %s: %w", name, err)
	}
	return nil
}

// DropExpiredPartitions drops every monthly partition whose entire range is older
// than the retention window (retentionMonths before asOf). Drops are idempotent
// and the default partition is never touched.
func (s *Store) DropExpiredPartitions(ctx context.Context, asOf time.Time, retentionMonths int) error {
	var (
		cutoff = firstOfMonth(asOf).AddDate(0, -retentionMonths, 0)
		month  = earliestPartitionMonth
		name   string
		err    error
	)
	// A future asOf would slide the cutoff forward and drop partitions still
	// holding live data. The only legitimate value is ~now, so reject it rather
	// than risk destructive data loss.
	if asOf.After(time.Now().UTC()) {
		return fmt.Errorf("dropping audit partitions: asOf %s is in the future", asOf.UTC().Format(time.RFC3339))
	}
	for month.Before(cutoff) {
		name = partitionName(month)
		if _, err = s.db.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s`, name)); err != nil {
			return fmt.Errorf("dropping audit partition %s: %w", name, err)
		}
		month = month.AddDate(0, 1, 0)
	}
	return nil
}

// partitionName returns the partition table name for the month containing t,
// e.g. audit_logs_2024_01.
func partitionName(t time.Time) string {
	return fmt.Sprintf("%s_%s", tableAuditLogs, t.Format(partitionNameFormat))
}

// firstOfMonth normalizes t to midnight UTC on the first day of its month.
func firstOfMonth(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
}
