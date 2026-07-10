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

import { faClose } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { cva, type VariantProps } from 'class-variance-authority';
import { forwardRef, HTMLAttributes, ReactNode } from 'react';
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { cn } from '../utils';

const alertVariants = cva('flex text-sm w-full rounded px-4 py-2 ', {
    variants: {
        variant: {
            // TODO: Update in https://specterops.atlassian.net/browse/BED-8069
            // default: 'bg-status-indeterminate-fill text-contrast',
            default: 'bg-status-indeterminate-fill text-neutral-1',
            error: 'bg-status-error-fill text-status-error-text [&>div>svg]:text-status-error-main',
            info: 'bg-status-info-fill text-status-info-text [&>div>svg]:text-status-info-main',
            success: 'bg-status-success-fill text-status-success-text [&>div>svg]:text-status-success-main',
            warning: 'bg-status-warning-fill text-status-warning-text [&>div>svg]:text-status-warning-main',
        },
    },
    defaultVariants: {
        variant: 'default',
    },
});

const iconMap = {
    default: <AppIcon.Info />,
    error: <AppIcon.Error />,
    info: <AppIcon.Info />,
    success: <AppIcon.CircleCheck />,
    warning: <AppIcon.Warning />,
};

const Alert = forwardRef<
    HTMLDivElement,
    Omit<HTMLAttributes<HTMLDivElement>, 'children'> &
        VariantProps<typeof alertVariants> & {
            action?: { label: string; onClick: () => void };
            children: ReactNode;
            onClose?: () => void;
            title?: string;
        }
>(({ action, children, className, onClose, title, variant = 'default', ...props }, ref) => (
    <div ref={ref} role='alert' className={cn(alertVariants({ variant }), className)} {...props}>
        <div className='my-3 mr-2'>{iconMap[variant ?? 'default']}</div>

        <div className='flex flex-col flex-grow py-2 gap-1'>
            {title && <h5 className={cn('text-base/6 font-medium tracking-tight')}>{title}</h5>}
            {children}
        </div>

        <div className='flex mt-1'>
            {action && (
                <button
                    aria-label='Alert action'
                    className='h-6 rounded-sm px-1 font-medium uppercase focus:outline-none focus-visible:focus-ring'
                    onClick={action.onClick}
                    type='button'>
                    {action.label}
                </button>
            )}

            {onClose && (
                <button
                    aria-label='Dismiss alert'
                    className='ml-2 inline-flex h-6 items-center justify-center rounded-sm px-1 focus:outline-none focus-visible:focus-ring'
                    onClick={onClose}
                    type='button'>
                    <FontAwesomeIcon icon={faClose} className='mt-1 size-4' />
                </button>
            )}
        </div>
    </div>
));

Alert.displayName = 'Alert';

export { Alert };
