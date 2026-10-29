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

var earliestPartitionMonth = time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

func (s *Store) CreateNextPartition(ctx context.Context, asOf time.Time) error {
	var (
		currentMonth = firstOfMonth(asOf)
		nextMonth    = currentMonth.AddDate(0, 1, 0)
	)

	for month := currentMonth; !month.After(nextMonth); month = month.AddDate(0, 1, 0) {
		if err := s.createPartition(ctx, month); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) createPartition(ctx context.Context, month time.Time) error {
	var (
		name  = partitionName(month)
		query = fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s PARTITION OF %s FOR VALUES FROM ('%s') TO ('%s')`,
			name,
			tableAuditLogs,
			month.Format(partitionDateFormat),
			month.AddDate(0, 1, 0).Format(partitionDateFormat),
		)
	)

	if _, err := s.database.Exec(ctx, query); err != nil {
		return fmt.Errorf("creating audit partition %s: %w", name, err)
	}

	return nil
}

func (s *Store) DropExpiredPartitions(ctx context.Context, asOf time.Time, retentionMonths int) error {
	if retentionMonths < 1 {
		return fmt.Errorf("dropping audit partitions: retention months must be positive")
	}
	if asOf.After(time.Now().UTC()) {
		return fmt.Errorf("dropping audit partitions: as-of time is in the future")
	}

	var cutoff = firstOfMonth(asOf).AddDate(0, -retentionMonths, 0)
	for month := earliestPartitionMonth; month.Before(cutoff); month = month.AddDate(0, 1, 0) {
		var name = partitionName(month)
		if _, err := s.database.Exec(ctx, fmt.Sprintf(`DROP TABLE IF EXISTS %s`, name)); err != nil {
			return fmt.Errorf("dropping audit partition %s: %w", name, err)
		}
	}

	return nil
}

func partitionName(value time.Time) string {
	return fmt.Sprintf("%s_%s", tableAuditLogs, value.Format(partitionNameFormat))
}

func firstOfMonth(value time.Time) time.Time {
	var utcValue = value.UTC()
	return time.Date(utcValue.Year(), utcValue.Month(), 1, 0, 0, 0, 0, time.UTC)
}
