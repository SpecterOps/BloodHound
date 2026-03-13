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
import { cva, type VariantProps } from 'class-variance-authority';
import { ColorOptions } from '../../types';
import { cn, getConditionalStyles, getCssColor } from '../utils';

// These variants are used to determine the component props and wrapper element
const RiskBadgePropVariants = cva('flex justify-center items-center rounded-full bg-neutral-light-3', {
    variants: {
        type: {
            labeled: 'size-[32px] p-2 shadow-none w-auto p-0 leading-[1] ',
            'sm-circle': 'size-4 p-1 drop-shadow',
            'md-circle': 'size-8 p-2 drop-shadow',
        },
    },
});

const RiskBadgeContentVariants = cva('rounded-full size-full', {
    variants: {
        outlined: {
            true: 'border-solid border-2 shadow-none bg-transparent',
        },
        type: {
            labeled: 'size-full px-6 py-2 border-none text-center',
            'sm-circle': 'shadow-inner1xl',
            'md-circle': 'shadow-inner1xl',
        },
    },
    compoundVariants: [
        {
            outlined: true,
            type: 'md-circle',
            className: 'border-2',
        },
        {
            outlined: true,
            type: 'sm-circle',
            className: 'border',
        },
    ],
});

export interface RiskBadgeProps
    extends React.HTMLAttributes<HTMLDivElement>,
        VariantProps<typeof RiskBadgePropVariants> {
    color?: ColorOptions;
    outlined: boolean;
    label?: string;
}

function RiskBadge(props: RiskBadgeProps) {
    const { className, color: _color = 'secondary', outlined = false, type, label = 'md-circle', ...rest } = props;

    const cssColor = getCssColor(_color);
    const labeled = type === 'labeled';

    const riskBadgeStyle = getConditionalStyles(
        [!outlined, { backgroundColor: cssColor }],
        [!!outlined, { borderColor: cssColor }],
        [
            !!(outlined && labeled),
            { boxShadow: `inset 0 0 0 2px ${cssColor}` }, // rather than shifting the elements sizing to account for the border, we can use inner shadow
        ]
    );

    return (
        <div role='status' className={cn(RiskBadgePropVariants({ type }), className)} {...rest}>
            <div style={riskBadgeStyle} className={cn(RiskBadgeContentVariants({ outlined, type }))}>
                {labeled ? label : null}
            </div>
        </div>
    );
}

export { RiskBadge };
