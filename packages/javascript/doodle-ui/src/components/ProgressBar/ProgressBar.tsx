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
import { Progress as ProgressPrimitive } from '@base-ui/react/progress';
import * as React from 'react';
import { cnWithState } from '../utils';

type ProgressBarProps = Omit<
    ProgressPrimitive.Root.Props,
    'value' | 'min' | 'max' | 'children' | 'render' | 'role' | 'aria-valuenow' | 'aria-valuemin' | 'aria-valuemax'
> & {
    /** The completion percentage, from 0 to 100, or null when progress is unknown. */
    value: number | null;
    /** A human-readable name for the process being tracked. */
    'aria-label': string;
};

/** Displays the completion progress of a task, built on Base UI. */
const ProgressBar = React.forwardRef<HTMLDivElement, ProgressBarProps>(({ className, value, ...props }, ref) => (
    <ProgressPrimitive.Root
        {...props}
        ref={ref}
        value={value}
        min={0}
        max={100}
        className={cnWithState('w-full', className)}>
        <ProgressPrimitive.Track className='relative h-2 w-full overflow-hidden bg-neutral-200 dark:bg-gray-700'>
            <ProgressPrimitive.Indicator className='h-full bg-primary motion-safe:transition-[width] data-[indeterminate]:w-1/3 motion-safe:data-[indeterminate]:animate-progress-indeterminate' />
        </ProgressPrimitive.Track>
    </ProgressPrimitive.Root>
));

ProgressBar.displayName = 'ProgressBar';

export { ProgressBar, type ProgressBarProps };
