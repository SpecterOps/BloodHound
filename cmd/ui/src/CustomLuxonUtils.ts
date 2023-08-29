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

// Adapted from the example provided here
// https://github.com/mui/material-ui-pickers/issues/1626#issuecomment-612031743
// Using this class instead of LuxonUtils with DateTimePicker will cause the DateTimePicker to use Sunday as the first
// day of the week instead of Monday

import LuxonUtils from '@date-io/luxon';
import { DateTime } from 'luxon';

class CustomLuxonUtils extends LuxonUtils {
    dayNames = ['S', 'M', 'T', 'W', 'T', 'F', 'S'];

    getWeekdays() {
        return this.dayNames;
    }

    getWeekArray(date: DateTime) {
        const endDate = date.endOf('month').plus({ days: 1 }).endOf('week');
        const startDate = date.startOf('month').startOf('week').minus({ days: 1 });

        const days = endDate.diff(startDate, 'days').days;

        const weeks: DateTime[][] = [];
        new Array(Math.round(days))
            .fill(0)
            .map((_, i) => i)
            .map((day) => startDate.plus({ days: day }))
            .forEach((v, i) => {
                if (i === 0 || (i % 7 === 0 && i > 6)) {
                    weeks.push([v]);
                    return;
                }

                weeks[weeks.length - 1].push(v);
            });
        // a consequence of all this shifting back/forth 1 day is that you might end up with a week
        // where all the days are actually in the previous or next month.
        // this happens when the first day of the month is Sunday (Dec 2019 or Mar 2020 are examples)
        // or the last day of the month is Sunday (May 2020 or Jan 2021 is one example)
        // so we're only including weeks where ANY day is in the correct month to handle that
        return weeks.filter((w) => w.some((d) => d.month === date.month));
    }
}

export default CustomLuxonUtils;
