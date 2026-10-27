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

export const alertVariants = cva('flex text-sm w-full rounded px-4 py-2 ', {
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

export const alertBodyClasses = 'flex flex-col justify-center flex-grow py-2 gap-1';

export const alertTitleClasses = 'text-base/6 font-medium tracking-tight';

export const alertActionClasses =
    'h-6 rounded-sm px-1 font-medium uppercase focus:outline-none focus-visible:focus-ring';

export const alertDismissClasses =
    'ml-2 inline-flex h-6 items-center justify-center rounded-sm px-1 focus:outline-none focus-visible:focus-ring';
