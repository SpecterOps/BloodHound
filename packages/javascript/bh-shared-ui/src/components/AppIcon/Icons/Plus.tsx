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

export const Plus: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='plus'
            width='20'
            height='20'
            viewBox='0 0 20 20'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath
                d='M14.6152 0.75C17.1718 0.75 19.25 2.82823 19.25 5.38477V14.6152C19.25 17.1718 17.1718 19.25 14.6152 19.25H5.38477C2.82823 19.25 0.75 17.1718 0.75 14.6152V5.38477C0.75 2.82823 2.82823 0.75 5.38477 0.75H14.6152ZM5.38477 2.63477C3.87053 2.63477 2.63477 3.87053 2.63477 5.38477V14.6152C2.63477 16.1295 3.87053 17.3652 5.38477 17.3652H14.6152C16.1295 17.3652 17.3652 16.1295 17.3652 14.6152V5.38477C17.3652 3.87053 16.1295 2.63477 14.6152 2.63477H5.38477Z'
                stroke='black'
                strokeWidth='0.5'
            />
            <BasePath
                d='M9.99994 4.44238C10.5165 4.44238 10.9423 4.86823 10.9423 5.38477V9.05762H14.6152C15.1317 9.05762 15.5575 9.4835 15.5576 10C15.5576 10.5165 15.1317 10.9424 14.6152 10.9424H10.9423V14.6152C10.9423 15.1318 10.5165 15.5576 9.99994 15.5576C9.48344 15.5576 9.05756 15.1317 9.05756 14.6152V10.9424H5.3847C4.86817 10.9424 4.44232 10.5165 4.44232 10C4.44236 9.4835 4.8682 9.05762 5.3847 9.05762H9.05756V5.38477C9.05756 4.86826 9.48344 4.44242 9.99994 4.44238Z'
                stroke='black'
                strokeWidth='0.5'
            />
        </BaseSVG>
    );
};

export default Plus;
