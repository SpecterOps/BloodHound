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

export const Trash: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='trash' version='1.1' viewBox='0 0 15 19' xmlns='http://www.w3.org/2000/svg' {...props}>
            <BasePath d='M 11.597 6.5 L 11.597 16.5 L 3.403 16.5 L 3.403 6.5 L 11.597 6.5 Z M 10.061 0.5 L 4.939 0.5 L 3.915 1.5 L 0.33 1.5 L 0.33 3.5 L 14.67 3.5 L 14.67 1.5 L 11.085 1.5 L 10.061 0.5 Z M 13.646 4.5 L 1.354 4.5 L 1.354 16.5 C 1.354 17.6 2.276 18.5 3.403 18.5 L 11.597 18.5 C 12.724 18.5 13.646 17.6 13.646 16.5 L 13.646 4.5 Z' />
        </BaseSVG>
    );
};
