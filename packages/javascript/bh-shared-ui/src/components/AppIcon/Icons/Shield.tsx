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

export const Shield: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG viewBox='0 0 20 20' fill='none' xmlns='http://www.w3.org/2000/svg' name='shield' {...props}>
            <BasePath d='M5.14211 7.87826L3.86914 9.15123L6.27194 11.554L6.27077 11.5552C6.62275 11.906 7.19294 11.906 7.54492 11.5552L12.0502 7.04995L10.7772 5.77698L6.90778 9.64519L5.14211 7.87826Z' />
            <BasePath d='M8.40981 0.207665L7.80911 0L7.20841 0.207665L1.2014 2.31012L0 2.73367V9.97272C0.00234634 11.3173 0.391867 12.6336 1.1228 13.7622C1.85373 14.8921 2.89439 15.7861 4.12045 16.3398L7.06998 17.668L7.80911 18L8.54824 17.6668L11.4978 16.3387V16.3398C12.7238 15.7861 13.7657 14.8909 14.4966 13.7611C15.2264 12.6312 15.617 11.3149 15.6182 9.96914V2.7301L14.4168 2.30655L8.40981 0.207665ZM13.8161 9.96905C13.8161 10.9675 13.5275 11.946 12.9855 12.7837C12.4423 13.6225 11.6691 14.2866 10.7586 14.6961L7.80911 16.0242L4.85958 14.6961C3.94916 14.2866 3.17596 13.6225 2.63275 12.7837C2.09071 11.946 1.8021 10.9675 1.8021 9.96905V4.0101L7.80911 1.90765L13.8161 4.0101V9.96905Z' />
        </BaseSVG>
    );
};
