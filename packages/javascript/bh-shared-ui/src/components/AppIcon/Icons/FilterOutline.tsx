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

export const FilterOutline: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='filter-outline'
            width='24'
            height='23'
            viewBox='0 0 24 23'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath
                d='M19.7592 0.5H4.24077C2.45538 0.5 1 1.95538 1 3.74077C1 4.60385 1.33846 5.41615 1.94769 6.03385L8.61538 12.7015V20.4608C8.61538 21.5862 9.52923 22.5 10.6546 22.5C11.1962 22.5 11.7208 22.28 12.1015 21.8992L14.64 19.3608C15.1223 18.8785 15.3846 18.2438 15.3846 17.5669V12.6931L22.0523 6.02538C22.6615 5.41615 23 4.60385 23 3.74077C23 1.95538 21.5446 0.5 19.7592 0.5ZM20.8508 4.83231L14.1915 11.5C13.87 11.8215 13.6923 12.2446 13.6923 12.7015V17.5754C13.6923 17.7954 13.5992 18.0154 13.4469 18.1762L10.9085 20.7146C10.7054 20.9092 10.3077 20.74 10.3077 20.4608V12.7015C10.3077 12.2531 10.13 11.8215 9.80846 11.5085L3.14923 4.83231C2.86154 4.54462 2.69231 4.14692 2.69231 3.74077C2.69231 2.88615 3.38615 2.19231 4.24077 2.19231H19.7592C20.6138 2.19231 21.3077 2.88615 21.3077 3.74077C21.3077 4.14692 21.1385 4.54462 20.8508 4.83231Z'
                stroke='black'
                strokeWidth='0.5'
            />
        </BaseSVG>
    );
};
