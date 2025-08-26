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

export const Disabled: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='disabled'
            xmlns='http://www.w3.org/2000/svg'
            width='16'
            height='16'
            viewBox='0 0 16 16'
            fill='none'
            {...props}>
            <BasePath
                d='M11.475 12.8906L3.10938 4.525C2.40937 5.50625 2 6.70625 2 8C2 11.3125 4.6875 14 8 14C9.29688 14 10.4969 13.5906 11.475 12.8906ZM12.8906 11.475C13.5906 10.4938 14 9.29375 14 8C14 4.6875 11.3125 2 8 2C6.70312 2 5.50313 2.40937 4.525 3.10938L12.8906 11.475ZM0 8C0 3.58125 3.58125 0 8 0C12.4187 0 16 3.58125 16 8C16 12.4187 12.4187 16 8 16C3.58125 16 0 12.4187 0 8Z'
                strokeWidth='0.25'
                stroke='currentColor'
            />
        </BaseSVG>
    );
};

export default Disabled;
