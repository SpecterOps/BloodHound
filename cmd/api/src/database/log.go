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
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/specterops/bloodhound/log"
)

type GormLogAdapter struct {
	SlowQueryWarnThreshold  time.Duration
	SlowQueryErrorThreshold time.Duration
}

func (s *GormLogAdapter) LogMode(level logger.LogLevel) logger.Interface {
	// This is a no-op - logging levels should be set with our global log package
	return s
}

func (s GormLogAdapter) Log(event log.Event, msg string, data ...any) {
	event.Msgf(msg, data...)
}

func (s *GormLogAdapter) Info(ctx context.Context, msg string, data ...any) {
	s.Log(log.Info(), msg, data...)
}

func (s *GormLogAdapter) Warn(ctx context.Context, msg string, data ...any) {
	s.Log(log.Warn(), msg, data...)
}

func (s *GormLogAdapter) Error(ctx context.Context, msg string, data ...any) {
	s.Log(log.Error(), msg, data...)
}

func (s *GormLogAdapter) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if log.GlobalLevel() > log.LevelDebug {
		return
	}

	if err != nil && err != gorm.ErrRecordNotFound {
		sql, _ := fc()

		if log.GlobalAccepts(log.LevelDebug) {
			log.Error().Fault(err).Msgf("Database error for query: %s", sql)
		} else {
			log.Error().Fault(err).Stack().Msgf("Database error for query: %s", sql)
		}
	} else {
		elapsed := time.Since(begin)

		if elapsed >= s.SlowQueryErrorThreshold {
			sql, rows := fc()

			if log.GlobalAccepts(log.LevelDebug) {
				log.Errorf("Slow database query took %d ms addressing %d rows: %s", elapsed.Milliseconds(), rows, sql)
			} else {
				log.Error().Stack().Msgf("Slow database query took %d ms addressing %d rows.", elapsed.Milliseconds(), rows)
			}
		} else if elapsed >= s.SlowQueryWarnThreshold {
			sql, rows := fc()

			if log.GlobalAccepts(log.LevelDebug) {
				log.Warnf("Slow database query took %d ms addressing %d rows: %s", elapsed.Milliseconds(), rows, sql)
			} else {
				log.Warn().Stack().Msgf("Slow database query took %d ms addressing %d rows.", elapsed.Milliseconds(), rows)
			}
		}
	}
}
