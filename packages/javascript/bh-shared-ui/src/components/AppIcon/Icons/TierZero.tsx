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

export const TierZero: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG viewBox='0 0 14 12' fill='none' xmlns='http://www.w3.org/2000/svg' name='tier-zero' {...props}>
            <BasePath d='M4.60739 1.26551L6.99998 3.71234L9.39257 1.26551H4.60739ZM10.4973 1.97851L8.51757 4.00175H12.0449L10.4973 1.97851ZM11.8645 5.26464H6.99998H2.13551L6.99998 10.4293L11.8645 5.26464ZM1.95504 4.00175H5.4824L3.5027 1.97851L1.95504 4.00175ZM13.8305 5.05679L7.4867 11.7922C7.36366 11.9237 7.18592 12 6.99998 12C6.81405 12 6.63904 11.9237 6.51326 11.7922L0.169481 5.05679C-0.0410666 4.83315 -0.057473 4.49901 0.128465 4.25696L3.19098 0.257838C3.31403 0.0973471 3.5109 0 3.71872 0H10.2812C10.4891 0 10.6859 0.0947161 10.809 0.257838L13.8715 4.25696C14.0574 4.49901 14.0383 4.83315 13.8305 5.05679Z' />
        </BaseSVG>
    );
};
