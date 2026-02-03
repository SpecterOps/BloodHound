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

import { Input } from '@bloodhoundenterprise/doodleui';
import { cn } from '../utils';
import { AppIcon } from './AppIcon';

interface Props {
    id: string;
    className?: string;
    disabled?: boolean;
    onInputChange: (search: string) => void;
    placeholder?: string;
    value: string;
}

export function SearchInput({
    className = '',
    disabled = false,
    id,
    onInputChange,
    placeholder = 'Search',
    value,
}: Props) {
    return (
        <div
            className={cn(
                'px-2 flex items-center w-1/7 border-b border-neutral-dark-1 dark:border-neutral-light-1',
                className
            )}>
            <AppIcon.MagnifyingGlass className='w-6 h-6 p-0.5' />
            <Input
                disabled={disabled}
                id={id}
                aria-label={placeholder}
                placeholder={placeholder}
                className='h-8 border-none bg-transparent dark:bg-transparent placeholder-neutral-dark-1 dark:placeholder-neutral-light-1 focus-visible:ring-0 focus-visible:ring-offset-0 px-2'
                onChange={(e) => onInputChange(e.target.value)}
                value={value}
            />
        </div>
    );
}
