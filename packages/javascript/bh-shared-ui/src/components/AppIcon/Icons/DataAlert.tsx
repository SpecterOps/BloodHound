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

export const DataAlert: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='file-magnifying-glass'
            xmlns='http://www.w3.org/2000/svg'
            viewBox='0 0 18 22'
            fill='none'
            {...props}>
            <BasePath
                d='M18 2H10L4.02 8L4 20C4 21.1 4.9 22 6 22H18C19.1 22 20 21.1 20 20V4C20 2.9 19.1 2 18 2ZM18 20H6V8.83L10.83 4H18V20Z'
                stroke='#ED8537'
                strokeWidth='0.25'
            />
            <BasePath d='M13 15H11V17H13V15Z' stroke='#ED8537' strokeWidth='0.25' />
            <BasePath d='M13 8H11V13H13V8Z' stroke='#ED8537' strokeWidth='0.25' />
        </BaseSVG>
    );
};

export default DataAlert;
