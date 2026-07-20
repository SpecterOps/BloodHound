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

// PreCreateNextPartition ensures the partition for the month AFTER asOf exists.
// Partition names and range bounds mirror the migration exactly
// (audit_logs_YYYY_MM, FROM first-of-month TO first-of-next-month). The DDL is
// safe from injection because every value is derived from a time.Time, never
// from external input.
func (s *Store) PreCreateNextPartition(ctx context.Context, asOf time.Time) error {
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

// DropExpiredPartitions drops every monthly partition whose entire range is
// older than the retention window. The cutoff is the first day of the month
// retentionMonths before asOf; any partition for a month strictly before the
// cutoff is fully expired. Drops are idempotent (DROP TABLE IF EXISTS) so a name
// that no longer exists is a no-op, and the default partition is never touched.
func (s *Store) DropExpiredPartitions(ctx context.Context, asOf time.Time, retentionMonths int) error {
	var (
		cutoff = firstOfMonth(asOf).AddDate(0, -retentionMonths, 0)
		month  = earliestPartitionMonth
		name   string
		err    error
	)
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
