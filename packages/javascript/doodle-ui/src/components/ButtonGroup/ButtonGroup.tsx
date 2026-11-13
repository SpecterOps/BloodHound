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
import { mergeProps } from '@base-ui/react/merge-props';
import { Separator } from '@base-ui/react/separator';
import { useRender } from '@base-ui/react/use-render';
import type { VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn, cnWithState } from '../utils';
import { ButtonGroupVariants, buttonGroupSeparatorClasses, buttonGroupTextClasses } from './ButtonGroup.styles';

export interface ButtonGroupProps
    extends React.ComponentPropsWithoutRef<'div'>,
        VariantProps<typeof ButtonGroupVariants> {}

export const ButtonGroup = React.forwardRef<HTMLDivElement, ButtonGroupProps>(function ButtonGroup(
    { className, orientation = 'horizontal', ...props },
    ref
) {
    return (
        <div
            {...props}
            ref={ref}
            role='group'
            data-slot='button-group'
            data-orientation={orientation}
            className={cn(ButtonGroupVariants({ orientation }), className)}
        />
    );
});

export type ButtonGroupTextProps = useRender.ComponentProps<'div'>;

export const ButtonGroupText = React.forwardRef<HTMLDivElement, ButtonGroupTextProps>(function ButtonGroupText(
    { className, render, ...props },
    ref
) {
    return useRender({
        defaultTagName: 'div',
        ref,
        render,
        state: { slot: 'button-group-text' },
        props: mergeProps<'div'>({ className: cn(buttonGroupTextClasses, className) }, props),
    });
});

export type ButtonGroupSeparatorProps = Separator.Props;

export const ButtonGroupSeparator = React.forwardRef<HTMLDivElement, ButtonGroupSeparatorProps>(
    function ButtonGroupSeparator({ className, orientation = 'vertical', ...props }, ref) {
        return (
            <Separator
                {...props}
                ref={ref}
                data-slot='button-group-separator'
                orientation={orientation}
                className={cnWithState(buttonGroupSeparatorClasses, className)}
            />
        );
    }
);
