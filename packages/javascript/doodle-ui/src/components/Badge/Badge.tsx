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
import { cn } from '../utils';

interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
    label: string;
    icon?: React.ReactNode;
    color?: string;
    backgroundColor?: string;
}

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
    ({ label, icon, color, backgroundColor, className, ...rest }, ref) => {
        return (
            <div
                ref={ref}
                {...rest}
                className={cn([
                    'inline-flex items-center justify-center rounded min-w-16 h-8 p-2 bg-neutral-light-3 dark:bg-neutral-dark-3 text-neutral-dark-1 dark:text-white border border-neutral-light-5 dark:border-neutral-dark-5',
                    icon && 'pr-3',
                    className,
                ])}
                style={{
                    borderColor: color,
                    backgroundColor: backgroundColor,
                }}>
                {icon && <span style={{ color }}>{icon}</span>}
                {label}
            </div>
        );
    }
);

export { Badge };
