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

export const CertStatus: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='cert-status'
            version='1.1'
            xmlns='http://www.w3.org/2000/svg'
            viewBox='0 0 24 24'
            fill='#2C2677'
            {...props}>
            <g id='cert-status'></g>
            <BasePath d='M3.93549 12.0002C3.93549 12.5347 3.50224 12.9679 2.96775 12.9679C2.43325 12.9679 2 12.5347 2 12.0002C2 11.4657 2.43325 11.0324 2.96775 11.0324C3.50224 11.0324 3.93549 11.4657 3.93549 12.0002Z' />

            <BasePath d='M4.74091 6.64696C5.20368 6.9142 5.36233 7.50602 5.0951 7.96889C4.82787 8.43177 4.23593 8.59032 3.77316 8.32319C3.31028 8.05595 3.15163 7.46403 3.41886 7.00115C3.68609 6.53828 4.27803 6.37973 4.74091 6.64696Z' />

            <BasePath d='M8.3207 3.7733C8.58794 4.23618 8.42928 4.82798 7.9665 5.09524C7.50363 5.36248 6.91172 5.20392 6.64446 4.74104C6.3772 4.27817 6.53588 3.68626 6.99876 3.419C7.46153 3.15186 8.05344 3.31042 8.3207 3.7733Z' />

            <BasePath d='M7.96641 18.905C8.42929 19.1722 8.58795 19.7642 8.32071 20.2269C8.05347 20.6898 7.46154 20.8485 6.99867 20.5812C6.53579 20.314 6.37724 19.7221 6.64447 19.2592C6.91171 18.7964 7.50353 18.6378 7.96641 18.905Z' />

            <BasePath d='M5.09506 16.0311C5.3623 16.494 5.20374 17.0858 4.74086 17.353C4.27799 17.6203 3.68608 17.4617 3.41882 16.9988C3.15168 16.5359 3.31024 15.944 3.77312 15.6768C4.236 15.4095 4.8278 15.5682 5.09506 16.0311Z' />

            <BasePath d='M12 2.00016V3.93565C14.1389 3.93565 16.1901 4.78524 17.7024 6.29766C19.2149 7.81008 20.0645 9.86137 20.0645 12.0001C20.0645 14.1388 19.2149 16.1901 17.7024 17.7025C16.19 19.2149 14.1387 20.0645 12 20.0645V22C14.6522 22 17.1957 20.9465 19.071 19.071C20.9464 17.1956 22 14.6522 22 12C22 9.34785 20.9465 6.8043 19.071 4.92903C17.1957 3.0536 14.6522 2 12 2V2.00016Z' />

            <BasePath d='M15.1871 8.73564L10.7096 13.2131L8.49028 10.9937L7.12256 12.3614L10.0258 15.2647H10.0257C10.2071 15.4458 10.4532 15.5477 10.7096 15.5477C10.966 15.5477 11.212 15.4458 11.3935 15.2647L16.5548 10.1034L15.1871 8.73564Z' />
        </BaseSVG>
    );
};
