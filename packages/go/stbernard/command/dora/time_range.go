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

package dora

import (
	"fmt"
	"time"
)

func calculateLastFiscalQuarterAt(fiscalStartMonth int, referenceTime time.Time) (time.Time, time.Time) {
	var (
		fiscalStartYear = referenceTime.Year()
		currentMonth    = int(referenceTime.Month())
	)

	if currentMonth < fiscalStartMonth {
		fiscalStartYear--
	}

	monthsIntoFiscalYear := (referenceTime.Year()-fiscalStartYear)*12 + currentMonth - fiscalStartMonth
	completedQuarterInFiscalYear := monthsIntoFiscalYear/3 - 1
	if completedQuarterInFiscalYear < 0 {
		completedQuarterInFiscalYear = 3
		fiscalStartYear--
	}

	quarterStart := time.Date(fiscalStartYear, time.Month(fiscalStartMonth), 1, 0, 0, 0, 0, time.UTC).
		AddDate(0, completedQuarterInFiscalYear*3, 0)
	quarterEnd := quarterStart.AddDate(0, 3, 0).Add(-time.Second)

	return quarterStart, quarterEnd
}

func resolveTimeRange(
	days int,
	startDate string,
	endDate string,
	lastQuarter bool,
	fiscalStartMonth int,
	referenceTime time.Time,
) (time.Time, time.Time, error) {
	if lastQuarter {
		if fiscalStartMonth < 1 || fiscalStartMonth > 12 {
			return time.Time{}, time.Time{}, fmt.Errorf("fiscal start month must be between 1 and 12")
		}

		startTime, endTime := calculateLastFiscalQuarterAt(fiscalStartMonth, referenceTime)
		return startTime, endTime, nil
	}

	if startDate != "" {
		startTime, err := time.Parse(time.DateOnly, startDate)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid start date format (use YYYY-MM-DD): %w", err)
		}

		endTime := referenceTime
		if endDate != "" {
			endTime, err = time.Parse(time.DateOnly, endDate)
			if err != nil {
				return time.Time{}, time.Time{}, fmt.Errorf("invalid end date format (use YYYY-MM-DD): %w", err)
			}
			endTime = endTime.Add(24*time.Hour - time.Second)
		}

		if startTime.After(endTime) {
			return time.Time{}, time.Time{}, fmt.Errorf(
				"start date (%s) cannot be after end date (%s)",
				startTime.Format(time.DateOnly),
				endTime.Format(time.DateOnly),
			)
		}

		return startTime, endTime, nil
	}

	startTime := referenceTime.AddDate(0, 0, -days)
	return startTime, referenceTime, nil
}
