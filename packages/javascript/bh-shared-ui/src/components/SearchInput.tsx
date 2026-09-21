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

import { Input } from 'doodle-ui';
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
        <div className={cn('flex-grow relative', className)}>
            <AppIcon.MagnifyingGlass className='absolute right-2 top-[50%] -mt-[8px] pointer-events-none' />
            <Input
                disabled={disabled}
                id={id}
                variant='outlined'
                aria-label={placeholder}
                placeholder={placeholder}
                onChange={(e) => onInputChange(e.target.value)}
                value={value}
            />
        </div>
    );
}
