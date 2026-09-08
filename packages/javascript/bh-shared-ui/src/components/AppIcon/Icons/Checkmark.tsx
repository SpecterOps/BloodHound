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

export const Checkmark: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='checkmark'
            width='16'
            height='12'
            viewBox='0 0 16 12'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M5.08471 9.46756L1.29164 5.73602L0 6.99776L5.08471 12L16 1.26174L14.7175 0L5.08471 9.46756Z' />
        </BaseSVG>
    );
};

export default Checkmark;
