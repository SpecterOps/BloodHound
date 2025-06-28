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

export const Upgrade: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            viewBox='0 0 20 20'
            style={{ color: '#ED8537' }}
            xmlns='http://www.w3.org/2000/svg'
            name='upgrade'
            {...props}>
            <BasePath d='M14 0H6L0.02 6L0 18C0 19.1 0.9 20 2 20H14C15.1 20 16 19.1 16 18V2C16 0.9 15.1 0 14 0ZM14 18H2V6.83L6.83 2H14V18Z' />
            <BasePath d='M9 13H7V15H9V13Z' />
            <BasePath d='M9 6H7V11H9V6Z' />
        </BaseSVG>
    );
};
