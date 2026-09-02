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

export const SortDesc: React.FC<BaseSVGProps> = (props) => {
    return (
        <BaseSVG name='sort-desc' version='1.1' viewBox='0 0 1200 1200' xmlns='http://www.w3.org/2000/svg' {...props}>
            <BasePath
                d='m138.91 688.66c-3.9375 3.9961-6.918 8.582-8.9375 13.473-2.0664 4.9805-3.207 10.445-3.207 16.172 0 5.7305 1.1406 11.195 3.207 16.176 2.0195 4.8867 5 9.4766 8.9375 13.473l0.25781 0.26172 427.27 427.27c16.504 16.504 43.254 16.504 59.758 0 1.8984-1.8984 3.582-3.9375 5.0469-6.0781l422.2-421.16 0.17969-0.17969c7.5781-7.6367 12.254-18.148 12.254-29.758 0-5.7188-1.1367-11.176-3.1953-16.152-2.0312-4.9062-5.0195-9.5078-8.9766-13.52l-0.32422-0.32812c-7.6367-7.5742-18.152-12.254-29.758-12.254h-854.6c-10.715 0-21.426 4.0508-29.648 12.148-0.15234 0.15234-0.30859 0.30469-0.46094 0.46094zm914.82-189.46c3.9375-4 6.918-8.582 8.9375-13.473 2.0664-4.9844 3.207-10.445 3.207-16.176 0-5.7266-1.1406-11.191-3.207-16.172-2.0195-4.8906-5-9.4766-8.9375-13.473l-0.26172-0.26172-427.27-427.27c-16.504-16.5-43.254-16.5-59.758 0-1.8984 1.9023-3.582 3.9375-5.043 6.082l-422.2 421.16-0.18359 0.18359c-7.5742 7.6367-12.25 18.148-12.25 29.754 0 5.7227 1.1367 11.18 3.1953 16.156 2.0273 4.9062 5.0195 9.5078 8.9766 13.52l0.32422 0.32422c7.6367 7.5742 18.152 12.254 29.758 12.254h854.6c10.715 0 21.426-4.0469 29.645-12.145 0.15625-0.15234 0.3125-0.30859 0.46484-0.46094zm-456.92-396.71 324.8 324.81-650.41-0.003907z'
                fillRule='evenodd'
            />
        </BaseSVG>
    );
};
