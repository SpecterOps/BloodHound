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

import React from 'react';
import { BasePath, BaseSVG, BaseSVGProps } from './utils';

export const LinesWithMagnifyingGlass: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='lines-with-magnifying-class'
            width='16'
            height='10'
            viewBox='0 0 16 10'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M4 2.30769H0V0.769231H4V2.30769ZM4 4.61538H0V6.15385H4V4.61538ZM14.872 10L11.808 7.05385C11.168 7.45385 10.416 7.69231 9.6 7.69231C7.392 7.69231 5.6 5.96923 5.6 3.84615C5.6 1.72308 7.392 0 9.6 0C11.808 0 13.6 1.72308 13.6 3.84615C13.6 4.63077 13.352 5.35385 12.936 5.96154L16 8.91539L14.872 10ZM12 3.84615C12 2.57692 10.92 1.53846 9.6 1.53846C8.28 1.53846 7.2 2.57692 7.2 3.84615C7.2 5.11538 8.28 6.15385 9.6 6.15385C10.92 6.15385 12 5.11538 12 3.84615ZM0 10H8V8.46154H0V10Z' />
        </BaseSVG>
    );
};
