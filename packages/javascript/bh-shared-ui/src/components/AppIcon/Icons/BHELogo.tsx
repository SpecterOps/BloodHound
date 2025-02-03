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
import { BaseSVG, BasePath, BaseSVGProps } from './utils';

export const BHELogo: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='bhe-logo' version='1.1' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1049 768' {...props}>
            <g id='icomoon-ignore'></g>
            <BasePath d='M574.181 180.835c-33.638 0-62.689 17.966-78.361 45.105h-156.723c-68.042 0-127.29-37.842-158.253-93.652l-24.081-42.049c139.138 0 278.278 0 417.415 0 66.893 0 125.378 36.695 156.341 90.595 60.395 16.054 120.791 32.109 180.805 48.547l-111.236 192.27-166.66 63.072-59.248 102.825-104.738-180.805h-104.353l208.708 361.224 125.761-217.501 167.044-63.072 181.952-314.973-262.606-70.334c-57.72-63.834-118.88-100.531-206.795-102.061h-579.109l105.12 181.952c51.221 87.916 144.107 136.079 239.288 134.551h151.372c15.671 26.755 44.722 45.105 78.361 45.105 49.693 0 90.212-40.52 90.212-90.595 0-49.693-40.519-90.212-90.212-90.212z' />
        </BaseSVG>
    );
};
