// Copyright 2025 Specter Ops, Inc.
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

import { faker } from '@faker-js/faker/locale/en';
import { faCaretDown, faCaretUp, faEllipsis, faGem, faUser, faUserGroup } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { ColumnDef } from '@tanstack/react-table';

export const getData = (entries: number): Finding[] =>
    Array(entries)
        .fill(undefined)
        .map(() => ({
            id: faker.animal.petName(),
            nonTierZeroPrincipal: `${faker.person.firstName()} ${faker.person.lastName()}`,
            tierZeroPrincipal: `${faker.person.firstName()} ${faker.person.lastName()}`,
            firstSeen: faker.date.month(),
        }));

interface Finding {
    id: string;
    nonTierZeroPrincipal: string;
    tierZeroPrincipal: string;
    firstSeen: string;
}

export const getColumns = (sortOrder?: string | null, handleSort?: (sortBy: string) => void) => {
    const columns: ColumnDef<Finding>[] = [
        {
            accessorKey: '',
            id: 'action-menu',
            size: 50,
            maxSize: 50,
            cell: () => (
                <button className='pl-4'>
                    <FontAwesomeIcon icon={faEllipsis} className='rotate-90 dark:text-neutral-light-1' />
                </button>
            ),
        },
        {
            accessorKey: 'nonTierZeroPrincipal',
            header: () => {
                return <span className='dark:text-neutral-light-1'>Non Tier Zero Principal</span>;
            },
            cell: ({ row }) => {
                return (
                    <div className='flex justify-start items-center max-w-2 dark:text-neutral-light-1 text-nowrap text-black'>
                        <span className='p-2 border border-black rounded-full bg-[#17E625] mr-2 size-4 flex justify-center items-center'>
                            <FontAwesomeIcon icon={faUser} className='text-xs' />
                        </span>
                        {row.getValue('nonTierZeroPrincipal')}
                    </div>
                );
            },
        },
        {
            size: 20,
            accessorKey: 'tierZeroPrincipal',
            header: () => (
                <div className='flex justify-start items-center dark:text-neutral-light-1 text-nowrap'>
                    <span className='p-2 rounded-full bg-black mr-2 size-4 flex justify-center items-center'>
                        <FontAwesomeIcon icon={faGem} className='text-xs' color='white' />
                    </span>
                    Tier Zero Principal
                    {sortOrder !== undefined ? (
                        <button onClick={() => handleSort?.('tierZeroPrincipal')} className='p-2'>
                            {!sortOrder && <FontAwesomeIcon icon={faCaretDown} color='black' />}
                            {sortOrder === 'asc' && <FontAwesomeIcon icon={faCaretUp} color='dodgerblue' />}
                            {sortOrder === 'desc' && <FontAwesomeIcon icon={faCaretDown} color='dodgerblue' />}
                        </button>
                    ) : null}
                </div>
            ),
            cell: ({ row }) => {
                return (
                    <div className='flex justify-start items-center dark:text-neutral-light-1 text-black'>
                        <span className='p-2 border border-black rounded-full bg-[#DBE617] mr-2 size-4 flex justify-center items-center'>
                            <FontAwesomeIcon icon={faUserGroup} className='text-xs' />
                        </span>
                        {row.getValue('tierZeroPrincipal')}
                    </div>
                );
            },
        },
        {
            accessorKey: 'firstSeen',
            header: () => {
                return <span className='dark:text-neutral-light-1'>First Seen</span>;
            },
            cell: ({ row }) => {
                return <span className='dark:text-neutral-light-1 text-black'>{row.getValue('firstSeen')}</span>;
            },
        },
    ];
    return columns;
};
