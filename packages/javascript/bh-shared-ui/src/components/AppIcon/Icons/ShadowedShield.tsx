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

export const ShadowedShield: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG viewBox='0 0 15 16' fill='none' xmlns='http://www.w3.org/2000/svg' name='shadowed-shield' {...props}>
            <BasePath d='M7.49688 0C7.64062 0 7.78437 0.0314219 7.91562 0.0911234L13.8031 2.60173C14.4906 2.89395 15.0031 3.5758 15 4.39906C14.9844 7.5161 13.7094 13.2192 8.325 15.8115C7.80312 16.0628 7.19688 16.0628 6.675 15.8115C1.28751 13.2192 0.0156392 7.5161 1.4214e-05 4.39906C-0.00311078 3.5758 0.509388 2.89395 1.19689 2.60173L7.08125 0.0911234C7.2125 0.0314219 7.35313 0 7.49688 0ZM7.49688 2.09898V13.9796C11.8094 11.8806 12.9687 7.23016 12.9969 4.44619L7.49688 2.10212V2.09898Z' />
        </BaseSVG>
    );
};
