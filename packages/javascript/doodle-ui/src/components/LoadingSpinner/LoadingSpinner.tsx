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

import { faCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import type { ComponentPropsWithoutRef } from 'react';
import { cn } from '../utils';
import { LoadingSpinnerIconVariants, LoadingSpinnerVariants } from './LoadingSpinner.styles';

type LoadingSpinnerProps = Omit<ComponentPropsWithoutRef<'span'>, 'children'> &
    (
        | { variant?: 'inline'; spinnerClassName?: never }
        | {
              variant: 'overlay';
              /** Style the overlay's icon without changing its backdrop. For inline spinners, use className. */
              spinnerClassName?: string;
          }
    );

function LoadingSpinner({ variant = 'inline', className, spinnerClassName, ...props }: LoadingSpinnerProps) {
    return (
        <span
            role='progressbar'
            aria-label='Loading'
            className={cn(LoadingSpinnerVariants({ variant }), className)}
            {...props}>
            <FontAwesomeIcon
                icon={faCircle}
                className={cn(LoadingSpinnerIconVariants({ variant }), variant === 'overlay' && spinnerClassName)}
            />
        </span>
    );
}

export { LoadingSpinner };
