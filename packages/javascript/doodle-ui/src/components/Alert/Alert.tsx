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
import { type VariantProps } from 'class-variance-authority';
import { forwardRef, HTMLAttributes, ReactNode } from 'react';
import { AppIcon } from '../../styleguide/components/AppIcons/AppIcons';
import { cn } from '../utils';
import {
    alertActionClasses,
    alertBodyClasses,
    alertDismissClasses,
    alertTitleClasses,
    alertVariants,
} from './Alert.styles';

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

        <div className={alertBodyClasses}>
            {title && <h5 className={alertTitleClasses}>{title}</h5>}
            {children}
        </div>

        <div className='flex mt-1'>
            {action && (
                <button aria-label='Alert action' className={alertActionClasses} onClick={action.onClick} type='button'>
                    {action.label}
                </button>
            )}

            {onClose && (
                <button aria-label='Dismiss alert' className={alertDismissClasses} onClick={onClose} type='button'>
                    <FontAwesomeIcon icon={faClose} className='mt-1 size-4' />
                </button>
            )}
        </div>
    </div>
));

Alert.displayName = 'Alert';

export { Alert };
