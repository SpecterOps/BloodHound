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
import { BasePath, BaseSVG, BaseSVGProps } from '../utils';

export const Warning: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='warning'
            width='16'
            height='13'
            viewBox='0 0 16 13'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M8 2.73L13.4764 11.6316H2.52364L8 2.73ZM8 0L0 13H16L8 0ZM8.72727 9.57895H7.27273V10.9474H8.72727V9.57895ZM8.72727 5.47368H7.27273V8.21053H8.72727V5.47368Z' />
        </BaseSVG>
    );
};
