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
import { BasePath, BaseSVG, type BaseSVGProps } from './utils';

export const Earth: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='earth' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' {...props}>
            <BasePath
                transform='translate(2 2) scale(1.25)'
                d='M8 0C3.584 0 0 3.584 0 8C0 12.416 3.584 16 8 16C12.416 16 16 12.416 16 8C16 3.584 12.416 0 8 0ZM1.6 8C1.6 7.512 1.664 7.032 1.768 6.576L5.592 10.4V11.2C5.592 12.08 6.312 12.8 7.192 12.8V14.344C4.048 13.944 1.6 11.256 1.6 8ZM12.712 12.32C12.504 11.672 11.912 11.2 11.192 11.2H10.392V8.8C10.392 8.36 10.032 8 9.592 8H4.792V6.4H6.392C6.832 6.4 7.192 6.04 7.192 5.6V4H8.792C9.672 4 10.392 3.28 10.392 2.4V2.072C12.736 3.016 14.4 5.32 14.4 8C14.4 9.664 13.752 11.184 12.712 12.32Z'
            />
        </BaseSVG>
    );
};

export default Earth;
