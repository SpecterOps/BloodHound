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
import { BaseSVG, BaseSVGProps } from './utils';

export const Zones: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='zones' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 20 14' fill='none' {...props}>
            <mask id='path-1-inside-1_2573_4201' fill='white'>
                <rect y='7.58826' width='9.41177' height='5.88235' rx='1' />
            </mask>
            <rect
                y='7.58826'
                width='9.41177'
                height='5.88235'
                rx='1'
                stroke='currentColor'
                strokeWidth='3'
                mask='url(#path-1-inside-1_2573_4201)'
            />
            <mask id='path-2-inside-2_2573_4201' fill='white'>
                <rect x='10.5879' y='7.58826' width='9.41177' height='5.88235' rx='1' />
            </mask>
            <rect
                x='10.5879'
                y='7.58826'
                width='9.41177'
                height='5.88235'
                rx='1'
                stroke='currentColor'
                strokeWidth='3'
                mask='url(#path-2-inside-2_2573_4201)'
            />
            <mask id='path-3-inside-3_2573_4201' fill='white'>
                <rect x='4.70605' y='0.529419' width='10.5882' height='5.88235' rx='1' />
            </mask>
            <rect
                x='4.70605'
                y='0.529419'
                width='10.5882'
                height='5.88235'
                rx='1'
                stroke='currentColor'
                strokeWidth='3'
                mask='url(#path-3-inside-3_2573_4201)'
            />
        </BaseSVG>
    );
};

export default Zones;
