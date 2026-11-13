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

export const datePickerInputClasses = cva('bg-[color:var(--date-picker-input-fill)] peer', {
    variants: {
        variant: {
            outlined: '',
            underlined: [
                'px-2',
                // Border images ignore border-radius, keeping the underline flat without changing the focus outline.
                '[border-image:linear-gradient(var(--input-border-default),var(--input-border-default))_1]',
                'focus:[border-image:linear-gradient(var(--input-border-focus),var(--input-border-focus))_1]',
            ],
        },
    },
});

export const datePickerTriggerClasses =
    'h-6 p-0 opacity-50 hover:opacity-100 peer-hover:opacity-100 focus-visible:opacity-100';

export const datePickerContentClasses = 'w-auto p-3 border border-[color:var(--date-picker-content-border)] z-[1500]';
