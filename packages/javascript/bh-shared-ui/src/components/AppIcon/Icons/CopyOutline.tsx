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

export const CopyOutline: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='copy-outline'
            width='13'
            height='16'
            viewBox='0 0 13 16'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M9.17647 1.6L1.52941 1.6L1.52941 12C1.52941 12.44 1.18529 12.8 0.764706 12.8C0.344118 12.8 0 12.44 0 12L0 1.6C0 0.719999 0.688235 0 1.52941 0H9.17647C9.59706 0 9.94118 0.36 9.94118 0.8C9.94118 1.24 9.59706 1.6 9.17647 1.6ZM13 4.8L13 14.4C13 15.28 12.3118 16 11.4706 16L4.58824 16C3.74706 16 3.05882 15.28 3.05882 14.4L3.05882 4.8C3.05882 3.92 3.74706 3.2 4.58824 3.2L11.4706 3.2C12.3118 3.2 13 3.92 13 4.8ZM11.4706 4.8L4.58824 4.8L4.58824 14.4L11.4706 14.4L11.4706 4.8Z' />
        </BaseSVG>
    );
};

export default CopyOutline;
