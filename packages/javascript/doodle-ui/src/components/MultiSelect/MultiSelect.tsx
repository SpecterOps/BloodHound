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
import { cva, VariantProps } from 'class-variance-authority';
import * as React from 'react';
import { cn } from '../utils';

// TODO: move to doodle-ui icons system in follow-up ticket
const CaretDown = ({ className, size = 12 }: { className?: string; size?: number }) => (
    <svg
        width={size}
        height={size}
        viewBox='0 0 10 6'
        fill='currentColor'
        xmlns='http://www.w3.org/2000/svg'
        className={className}>
        <path d='M0 0.977539L5 5.97754L10 0.977539H0Z' />
    </svg>
);

const MultiSelectTriggerVariants = cva(
    'flex w-full items-center justify-between rounded-md border px-3 py-2 text-sm text-input-placeholder-text hover:text-white focus:outline-none focus-visible:focus-ring disabled:cursor-not-allowed disabled:opacity-50',
    {
        variants: {
            variant: {
                default: 'border-input-border-default bg-input-fill hover:border-input-border-hover hover:bg-secondary',
                error: 'border-status-error-main bg-input-fill',
                disabled: 'border-input-border-disabled bg-input-fill-disabled',
            },
        },
        defaultVariants: {
            variant: 'default',
        },
    }
);

interface MultiSelectTriggerProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement>,
        VariantProps<typeof MultiSelectTriggerVariants> {
    open?: boolean;
}

const MultiSelectTrigger = React.forwardRef<HTMLButtonElement, MultiSelectTriggerProps>(
    ({ className, variant, open, children, ...props }, ref) => (
        <button ref={ref} type='button' className={cn(MultiSelectTriggerVariants({ variant, className }))} {...props}>
            <span className='truncate'>{children}</span>
            <CaretDown className={cn('ml-2 shrink-0 transition-transform duration-200', open && 'rotate-180')} />
        </button>
    )
);
MultiSelectTrigger.displayName = 'MultiSelectTrigger';

interface MultiSelectProps {
    value: string[];
    onValueChange: (values: string[]) => void;
    placeholder?: string;
    disabled?: boolean;
    error?: boolean;
    className?: string;
}

interface MultiSelectOption {
    value: string;
    label: string;
    disabled?: boolean;
}

/**
 * Description for MultiSelect
 */
const MultiSelect = ({ ...props }: MultiSelectProps) => {
    return <div {...props} />;
};

export { MultiSelect, MultiSelectTrigger };
export type { MultiSelectOption, MultiSelectProps };
