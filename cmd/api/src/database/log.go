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

package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/bhlog"
	"github.com/specterops/bloodhound/bhlog/handlers"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type GormLogAdapter struct {
	SlowQueryWarnThreshold  time.Duration
	SlowQueryErrorThreshold time.Duration
}

func (s *GormLogAdapter) LogMode(level logger.LogLevel) logger.Interface {
	// This is a no-op - logging levels should be set with our global log package
	return s
}

func (s *GormLogAdapter) Info(ctx context.Context, msg string, data ...any) {
	slog.InfoContext(ctx, fmt.Sprintf(msg, data...))
}

func (s *GormLogAdapter) Warn(ctx context.Context, msg string, data ...any) {
	slog.WarnContext(ctx, fmt.Sprintf(msg, data...))
}

func (s *GormLogAdapter) Error(ctx context.Context, msg string, data ...any) {
	slog.ErrorContext(ctx, fmt.Sprintf(msg, data...))
}

func (s *GormLogAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if bhlog.GlobalLevel() > bhlog.LevelDebug {
		return
	}

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		sql, _ := fc()

		if slog.Default().Enabled(ctx, slog.LevelDebug) {
			slog.ErrorContext(ctx, "Database error", "query", sql, "error", err, handlers.GetSlogCallStack())
		} else {
			slog.ErrorContext(ctx, "Database error", "query", sql, "error", err)
		}
	} else {
		elapsed := time.Since(begin)

		if elapsed >= s.SlowQueryErrorThreshold {
			sql, rows := fc()

			if slog.Default().Enabled(ctx, slog.LevelDebug) {
				slog.ErrorContext(ctx, "Slow database query", "duration_ms", elapsed.Milliseconds(), "nums_rows", rows, "sql", sql, handlers.GetSlogCallStack())
			} else {
				slog.ErrorContext(ctx, "Slow database query", "duration_ms", elapsed.Milliseconds(), "num_rows", rows)
			}
		} else if elapsed >= s.SlowQueryWarnThreshold {
			sql, rows := fc()

			if bhlog.GlobalAccepts(bhlog.LevelDebug) {
				slog.WarnContext(ctx, "Slow database query", "duration_ms", elapsed.Milliseconds(), "nums_rows", rows, "sql", sql, handlers.GetSlogCallStack())
			} else {
				slog.WarnContext(ctx, "Slow database query", "duration_ms", elapsed.Milliseconds(), "num_rows", rows)
			}
		}
	}
}
