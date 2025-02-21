// Copyright 2025 Specter Ops, Inc.
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

import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const CalendarDay: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG viewBox='0 0 18 20' fill='none' xmlns='http://www.w3.org/2000/BaseSVG' name='calendar-day' {...props}>
            <BasePath d='M6 14.5C5.3 14.5 4.70833 14.2583 4.225 13.775C3.74167 13.2917 3.5 12.7 3.5 12C3.5 11.3 3.74167 10.7083 4.225 10.225C4.70833 9.74167 5.3 9.5 6 9.5C6.7 9.5 7.29167 9.74167 7.775 10.225C8.25833 10.7083 8.5 11.3 8.5 12C8.5 12.7 8.25833 13.2917 7.775 13.775C7.29167 14.2583 6.7 14.5 6 14.5ZM2 20C1.45 20 0.979167 19.8042 0.5875 19.4125C0.195833 19.0208 0 18.55 0 18V4C0 3.45 0.195833 2.97917 0.5875 2.5875C0.979167 2.19583 1.45 2 2 2H3V0H5V2H13V0H15V2H16C16.55 2 17.0208 2.19583 17.4125 2.5875C17.8042 2.97917 18 3.45 18 4V18C18 18.55 17.8042 19.0208 17.4125 19.4125C17.0208 19.8042 16.55 20 16 20H2ZM2 18H16V8H2V18Z' />
        </BaseSVG>
    );
};

export default CalendarDay;
