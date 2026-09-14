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
import * as React from 'react';
import { Tooltip } from '../Tooltip';

export interface IconProps {
    children: React.ReactElement;
    tooltip?: React.ReactNode;
}

export const Icon: React.FC<IconProps> = ({ children, tooltip }) => {
    if (!tooltip) {
        return children;
    }

    return (
        <Tooltip tooltip={tooltip} contentProps={{ side: 'bottom', align: 'start' }}>
            <span className='inline-flex'>{children}</span>
        </Tooltip>
    );
};

Icon.displayName = 'Icon';
