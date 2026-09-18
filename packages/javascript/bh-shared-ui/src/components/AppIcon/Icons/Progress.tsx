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

export const Progress: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG
            name='progress'
            width='16'
            height='16'
            viewBox='0 0 16 16'
            fill='none'
            xmlns='http://www.w3.org/2000/svg'
            {...props}>
            <BasePath d='M6.8001 2.8001C6.8001 2.1376 7.3376 1.6001 8.0001 1.6001C8.6626 1.6001 9.2001 2.1376 9.2001 2.8001C9.2001 3.4626 8.6626 4.0001 8.0001 4.0001C7.3376 4.0001 6.8001 3.4626 6.8001 2.8001ZM6.8001 13.2001C6.8001 12.5376 7.3376 12.0001 8.0001 12.0001C8.6626 12.0001 9.2001 12.5376 9.2001 13.2001C9.2001 13.8626 8.6626 14.4001 8.0001 14.4001C7.3376 14.4001 6.8001 13.8626 6.8001 13.2001ZM2.8001 6.8001C3.4626 6.8001 4.0001 7.3376 4.0001 8.0001C4.0001 8.6626 3.4626 9.2001 2.8001 9.2001C2.1376 9.2001 1.6001 8.6626 1.6001 8.0001C1.6001 7.3376 2.1376 6.8001 2.8001 6.8001ZM12.0001 8.0001C12.0001 7.3376 12.5376 6.8001 13.2001 6.8001C13.8626 6.8001 14.4001 7.3376 14.4001 8.0001C14.4001 8.6626 13.8626 9.2001 13.2001 9.2001C12.5376 9.2001 12.0001 8.6626 12.0001 8.0001ZM3.4751 10.8276C3.9451 10.3576 4.7026 10.3576 5.1726 10.8276C5.6426 11.2976 5.6426 12.0551 5.1726 12.5251C4.7026 12.9951 3.9451 12.9951 3.4751 12.5251C3.0051 12.0551 3.0051 11.2976 3.4751 10.8276ZM3.4751 3.4751C3.9451 3.0051 4.7026 3.0051 5.1726 3.4751C5.6426 3.9451 5.6426 4.7026 5.1726 5.1726C4.7026 5.6426 3.9451 5.6426 3.4751 5.1726C3.0051 4.7026 3.0051 3.9451 3.4751 3.4751ZM12.5251 10.8276C12.9951 11.2976 12.9951 12.0551 12.5251 12.5251C12.0551 12.9951 11.2976 12.9951 10.8276 12.5251C10.3576 12.0551 10.3576 11.2976 10.8276 10.8276C11.2976 10.3576 12.0551 10.3576 12.5251 10.8276Z' />
        </BaseSVG>
    );
};

export default Progress;
