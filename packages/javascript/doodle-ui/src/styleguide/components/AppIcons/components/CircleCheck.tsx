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

export const CircleCheck: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='circle-check'
            width='16'
            height='16'
            viewBox='0 0 16 16'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M8 16C3.58125 16 0 12.4187 0 8C0 3.58125 3.58125 0 8 0C12.4187 0 16 3.58125 16 8C16 12.4187 12.4187 16 8 16ZM8 1.5C4.40938 1.5 1.5 4.40938 1.5 8C1.5 11.5906 4.40938 14.5 8 14.5C11.5906 14.5 14.5 11.5906 14.5 8C14.5 4.40938 11.5906 1.5 8 1.5ZM10.2094 5.30937C10.4531 4.975 10.9219 4.9 11.2563 5.14375C11.5906 5.3875 11.6656 5.85625 11.4219 6.19063L7.60625 11.4406C7.47812 11.6188 7.27812 11.7312 7.05937 11.7469C6.84062 11.7625 6.625 11.6844 6.47188 11.5312L4.725 9.78438C4.43125 9.49063 4.43125 9.01562 4.725 8.725C5.01875 8.43438 5.49375 8.43125 5.78438 8.725L6.90938 9.85L10.2094 5.3125V5.30937Z' />
        </BaseSVG>
    );
};
