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

export const Findings: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG viewBox='0 0 20 20' fill='none' xmlns='http://www.w3.org/2000/svg' name='findings' {...props}>
            <BasePath d='M16.4803 3.51933L15.1878 4.81183C16.5078 6.141 17.3328 7.97433 17.3328 10.0002C17.3328 14.0518 14.0511 17.3335 9.99943 17.3335C5.94776 17.3335 2.6661 14.0518 2.6661 10.0002C2.6661 6.26016 5.46193 3.18016 9.08276 2.731V4.58266C6.47943 5.02266 4.49943 7.27766 4.49943 10.0002C4.49943 13.0343 6.96526 15.5002 9.99943 15.5002C13.0336 15.5002 15.4994 13.0343 15.4994 10.0002C15.4994 8.4785 14.8853 7.1035 13.8861 6.1135L12.5936 7.406C13.2536 8.07516 13.6661 8.99183 13.6661 10.0002C13.6661 12.026 12.0253 13.6668 9.99943 13.6668C7.9736 13.6668 6.33276 12.026 6.33276 10.0002C6.33276 8.29516 7.5061 6.87433 9.08276 6.46183V8.4235C8.53276 8.74433 8.1661 9.32183 8.1661 10.0002C8.1661 11.0085 8.9911 11.8335 9.99943 11.8335C11.0078 11.8335 11.8328 11.0085 11.8328 10.0002C11.8328 9.32183 11.4661 8.73516 10.9161 8.4235V0.833496H9.99943C4.93943 0.833496 0.832764 4.94016 0.832764 10.0002C0.832764 15.0602 4.93943 19.1668 9.99943 19.1668C15.0594 19.1668 19.1661 15.0602 19.1661 10.0002C19.1661 7.47016 18.1394 5.1785 16.4803 3.51933Z' />
        </BaseSVG>
    );
};
