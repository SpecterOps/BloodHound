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

export const LinedPaper: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            viewBox='0 0 20 20'
            fill='black'
            stroke='black'
            xmlns='http://www.w3.org/2000/BaseSVG'
            name='lined-paper'
            {...props}>
            <BasePath d='M15.625 1H4.375C3.47989 1 2.62145 1.35558 1.98851 1.98851C1.35558 2.62145 1 3.47989 1 4.375V15.625C1 16.5201 1.35558 17.3786 1.98851 18.0115C2.62145 18.6444 3.47989 19 4.375 19H12.443C13.0398 19 13.6121 18.7629 14.034 18.341L18.341 14.034C18.5499 13.8251 18.7156 13.577 18.8287 13.304C18.9418 13.0311 19 12.7385 19 12.443V4.375C19 3.47989 18.6444 2.62145 18.0115 1.98851C17.3786 1.35558 16.5201 1 15.625 1ZM13.375 17.409V14.5C13.3753 14.2017 13.494 13.9158 13.7049 13.7049C13.9158 13.494 14.2017 13.3753 14.5 13.375H17.409L13.375 17.409ZM17.875 12.25H14.5C13.9033 12.25 13.331 12.4871 12.909 12.909C12.4871 13.331 12.25 13.9033 12.25 14.5V17.875H4.375C3.77847 17.8743 3.20657 17.637 2.78477 17.2152C2.36296 16.7934 2.12568 16.2215 2.125 15.625V4.375C2.12567 3.77847 2.36294 3.20656 2.78475 2.78475C3.20656 2.36294 3.77847 2.12567 4.375 2.125H15.625C16.2215 2.12567 16.7934 2.36294 17.2152 2.78475C17.6371 3.20656 17.8743 3.77847 17.875 4.375V12.25Z' />
        </BaseSVG>
    );
};
