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

import { Tooltip } from 'doodle-ui';
import { useState } from 'react';
import { AppIcon } from '../AppIcon';
import { entityPropertyHelp } from './entityPropertyHelp';

export const EntityPropertyLabel = ({
    property,
    label,
    nodeType,
}: {
    property?: string;
    label: string;
    nodeType?: string;
}) => {
    const [open, setOpen] = useState(false);
    const help =
        property && Object.hasOwn(entityPropertyHelp, property)
            ? entityPropertyHelp[property as keyof typeof entityPropertyHelp]
            : undefined;

    if (!help || !help.nodeTypes.some((kind) => kind === nodeType)) return <>{label}</>;

    return (
        <Tooltip
            open={open}
            onOpenChange={setOpen}
            contentWidth='lg'
            tooltipPosition='left'
            triggerProps={{
                asChild: false,
                'aria-label': `About ${label}`,
                className: 'text-left',
                onFocus: (event) => {
                    // Restore keyboard help after hover help was dismissed with Escape.
                    if (event.currentTarget.matches(':focus-visible')) setOpen(true);
                },
            }}
            tooltip={
                <>
                    <span className='block'>{help.description}</span>
                    <span className='mt-2 block'>{help.location}</span>
                </>
            }>
            <span className='inline-flex items-center gap-1'>
                {label}
                <AppIcon.Info size={14} className='shrink-0' aria-hidden='true' />
            </span>
        </Tooltip>
    );
};
