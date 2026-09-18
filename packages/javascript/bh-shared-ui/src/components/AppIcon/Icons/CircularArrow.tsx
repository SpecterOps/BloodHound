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

export const CircularArrow: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='circular-arrow'
            width='16'
            height='16'
            viewBox='0 0 16 16'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M2.79099 9.11557C2.79099 7.68313 3.37265 6.3809 4.31893 5.43462L3.08616 4.20185C1.83602 5.46066 1.05469 7.19696 1.05469 9.11557C1.05469 12.6576 3.70254 15.5746 7.13174 16V14.2463C4.67487 13.8296 2.79099 11.694 2.79099 9.11557ZM14.9451 9.11557C14.9451 5.27835 11.8371 2.17037 7.99989 2.17037C7.9478 2.17037 7.89571 2.17906 7.84362 2.17906L8.7899 1.23277L7.56581 0L4.52729 3.03852L7.56581 6.07705L8.7899 4.85296L7.8523 3.91536C7.90439 3.91536 7.95648 3.90667 7.99989 3.90667C10.8735 3.90667 13.2088 6.242 13.2088 9.11557C13.2088 11.694 11.3249 13.8296 8.86804 14.2463V16C12.2972 15.5746 14.9451 12.6576 14.9451 9.11557Z' />
        </BaseSVG>
    );
};

export default CircularArrow;
