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

export const AttackPaths: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='attack-paths' version='1.1' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 768 768' {...props}>
            <g id='attack-paths'></g>
            <BasePath d='M567.273 506.183l122.181-122.181-122.181-122.181v91.637h-154.561c-10.386-94.691-69.033-174.722-150.894-215.345-1.221-50.094-41.541-90.11-91.637-90.11-50.706 0-91.636 40.93-91.636 91.636s40.93 91.637 91.636 91.637c29.019 0 54.371-13.745 71.171-34.822 58.036 31.767 99.883 88.886 109.659 157.004h-94.691c-12.829-35.432-46.429-61.090-86.138-61.090-50.706 0-91.636 40.93-91.636 91.636s40.93 91.637 91.636 91.637c39.71 0 73.309-25.658 86.138-61.091h94.691c-9.774 68.116-51.621 125.236-109.353 157.004-17.105-21.077-42.458-34.822-71.477-34.822-50.706 0-91.636 40.93-91.636 91.636s40.93 91.637 91.636 91.637c50.094 0 90.414-40.014 91.331-90.11 81.861-40.626 140.51-120.654 150.894-215.345h154.865v91.637z' />
        </BaseSVG>
    );
};

export default AttackPaths;
