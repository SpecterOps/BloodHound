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

import { cn } from '../utils';

export type StatusType = 'good' | 'bad' | 'pending';

type Props = {
    label?: string;
    pulse?: boolean;
    status: StatusType;
};

const STATUS_COLORS: Record<StatusType, string> = {
    good: 'fill-[#BCD3A8]', // Light Olive Green
    bad: 'fill-[#D9442E]', // Red
    pending: 'fill-[#5CC3AD]', // Aqua Green
};

/**
 * Displays a colored circle and label corresponding to a given status value and type.
 *
 * @example
 * ```tsx
 * const status = <StatusIndicator label="Success" status="good" />
 * ```
 */
export const StatusIndicator: React.FC<Props> = ({ label = '', pulse = false, status }) => {
    const color = STATUS_COLORS[status];
    return (
        <span className='inline-flex items-center'>
            <span className='mr-1.5'>
                <svg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 12 12' fill='none'>
                    <circle cx='6' cy='6' r='6' className={cn(color, pulse && 'animate-pulse')} />
                </svg>
            </span>
            {label && <span>{label}</span>}
        </span>
    );
};
