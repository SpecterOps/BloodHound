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

export const ButtonGroupVariants = cva(
    [
        'flex w-fit items-stretch',
        '[&>*:focus-visible]:relative [&>*:focus-visible]:z-10',
        '[&>input]:min-w-0 [&>input]:flex-1',
        'has-[[data-slot=button-group]]:gap-2',
    ],
    {
        variants: {
            orientation: {
                horizontal: [
                    '[&>*:not(:last-child):not([data-slot=button-group])]:rounded-r-none',
                    '[&>*:not(:first-child):not([data-slot=button-group])]:rounded-l-none',
                    '[&>*:not(:first-child):not([data-slot=button-group])]:border-l-0',
                ],
                vertical: [
                    'flex-col',
                    '[&>*:not(:last-child):not([data-slot=button-group])]:rounded-b-none',
                    '[&>*:not(:first-child):not([data-slot=button-group])]:rounded-t-none',
                    '[&>*:not(:first-child):not([data-slot=button-group])]:border-t-0',
                ],
            },
        },
        defaultVariants: { orientation: 'horizontal' },
    }
);

export const buttonGroupTextClasses =
    'flex items-center gap-2 rounded border border-border-control bg-surface-subtle px-3 py-2 text-sm text-main [&_svg]:pointer-events-none [&_svg]:shrink-0';

export const buttonGroupSeparatorClasses = [
    'relative shrink-0 self-stretch bg-border-control',
    'data-[orientation=vertical]:my-px data-[orientation=vertical]:h-auto data-[orientation=vertical]:w-px',
    'data-[orientation=horizontal]:mx-px data-[orientation=horizontal]:h-px data-[orientation=horizontal]:w-auto',
];
