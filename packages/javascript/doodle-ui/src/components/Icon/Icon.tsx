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
import { cn } from '../utils';

export interface IconProps {
    'aria-hidden'?: boolean;
    'aria-label': string;
    children: React.ReactElement;
    className?: string;
    hideTooltip?: boolean;
}

export const Icon: React.FC<IconProps> = ({
    'aria-hidden': ariaHidden = false,
    'aria-label': ariaLabel,
    children,
    className,
    hideTooltip = false,
}) => {
    const iconElement = React.cloneElement(
        children,
        ariaHidden ? { 'aria-hidden': true } : { 'aria-label': ariaLabel }
    );

    if (ariaHidden || hideTooltip) return <span className={cn('inline-flex', className)}>{iconElement}</span>;

    return (
        <Tooltip tooltip={ariaLabel} contentProps={{ side: 'bottom', align: 'start' }}>
            <span className={cn('inline-flex', className)}>{iconElement}</span>
        </Tooltip>
    );
};

Icon.displayName = 'Icon';
