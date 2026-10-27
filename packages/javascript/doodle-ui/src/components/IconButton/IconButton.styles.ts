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

import { cva } from 'class-variance-authority';
import { buttonBaseClasses, primaryClasses, secondaryClasses } from '../Button/Button.styles';
import { cn } from '../utils';

export const defaultIconButtonClasses = cn(
    'hover:text-primary',
    'active:bg-transparent active:text-secondary',
    'focus-visible:ring-transparent focus-visible:ring-offset-0 focus-visible:ring-offset-transparent focus-visible:text-primary'
);

export const IconButtonVariants = cva(
    cn(
        buttonBaseClasses,
        'inline-grid h-fit aspect-square box-border',
        'shrink-0 place-items-center align-middle rounded-full border-0 p-2'
    ),
    {
        variants: {
            variant: {
                default: defaultIconButtonClasses,
                primary: primaryClasses,
                secondary: secondaryClasses,
            },
        },
        defaultVariants: {
            variant: 'default',
        },
    }
);

export const iconButtonIconClasses =
    'inline-flex size-[var(--icon-button-icon-size)] shrink-0 items-center justify-center [&>*]:size-full [&_svg]:size-full';
