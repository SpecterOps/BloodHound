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

export const CheckboxVariants = cva(
    [
        'peer shrink-0 rounded-sm border border-checkbox-border ring-offset-background',
        'enabled:hover:border-checkbox-hover enabled:active:border-primary-variant',
        'dark:border-checkbox-border dark:hover:border-checkbox-hover dark:active:border-primary-variant',
        'focus:outline-none',
        'data-[state=checked]:bg-checkbox-fill data-[state=checked]:text-checkbox-check dark:data-[state=checked]:bg-primary-main',
        'enabled:hover:data-[state=checked]:border-checkbox-hover enabled:hover:data-[state=checked]:bg-checkbox-hover enabled:active:data-[state=checked]:bg-primary-variant',
        'dark:data-[state=checked]:hover:bg-checkbox-hover dark:data-[state=checked]:active:bg-primary-variant',
        'enabled:hover:data-[state=indeterminate]:border-checkbox-hover enabled:hover:data-[state=indeterminate]:bg-secondary',
        'enabled:data-[state=checked]:border-primary enabled:data-[state=checked]:bg-primary enabled:data-[state=checked]:text-checkbox-check',
        'enabled:data-[state=indeterminate]:border-primary enabled:data-[state=indeterminate]:bg-primary enabled:data-[state=indeterminate]:text-checkbox-check',
        'disabled:cursor-not-allowed disabled:border-input-border-disabled disabled:bg-input-fill-disabled disabled:text-icon-disabled',
        'dark:disabled:bg-input-fill-disabled-dark dark:disabled:border-input-border-disabled dark:hover:disabled:border-input-border-disabled',
        'enabled:aria-[invalid=true]:data-[state=unchecked]:border-status-error-main',
    ],
    {
        variants: {
            size: {
                lg: 'size-[24px]',
                md: 'size-[18px]',
                sm: 'size-[12px]',
            },
            focusRing: {
                true: 'focus-visible:focus-ring',
                false: '',
            },
        },
        defaultVariants: {
            size: 'md',
            focusRing: true,
        },
    }
);
