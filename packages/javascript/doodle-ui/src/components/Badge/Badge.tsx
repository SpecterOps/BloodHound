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
import { type VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';
import { badgeIconClasses, badgeLabelClasses, BadgeVariants } from './Badge.styles';

type NamedColor = NonNullable<VariantProps<typeof BadgeVariants>['color']>;

type BadgeProps = Omit<React.HTMLAttributes<HTMLDivElement>, 'color'> &
    Omit<VariantProps<typeof BadgeVariants>, 'hasIcon'> & {
        label: string;
        icon?: React.ReactNode;
        iconClassName?: string;
    };

const Badge = React.forwardRef<HTMLDivElement, BadgeProps>(
    ({ className, color, icon, iconClassName, iconPosition, label, variant, ...rest }, ref) => {
        return (
            <div
                ref={ref}
                {...rest}
                className={cn(
                    BadgeVariants({
                        variant,
                        color,
                        hasIcon: !!icon,
                        iconPosition: icon ? iconPosition ?? 'left' : undefined,
                    }),
                    className
                )}>
                {icon && (
                    // badge-icon does not have actual properties, if you want to leverage it, you would have to target the className and define the properties via the Badge component instance in the parent
                    <span className={cn(badgeIconClasses, iconClassName)}>{icon}</span>
                )}
                {/* badge-label does not have actual properties, if you want to leverage it, you would have to target the
                className and define the properties via the Badge component instance in the parent */}
                <span className={badgeLabelClasses}>{label}</span>
            </div>
        );
    }
);

Badge.displayName = 'Badge';

export { Badge };
export type { NamedColor };
