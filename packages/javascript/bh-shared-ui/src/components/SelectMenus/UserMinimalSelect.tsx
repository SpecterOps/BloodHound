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

import { Label, Select, SelectContent, SelectItem, SelectPortal, SelectTrigger, SelectValue } from 'doodle-ui';
import { User } from 'js-client-library';
import type { FC } from 'react';
import { useGetUsersMinimal } from '../../hooks/useGetUsers';

type Props = {
    user?: string;
    onSelect: (value: string) => void;
};

// Named using the Minimal keyword as it uses a specific endpoint /bloodhound-users-minimal that gets active users
export const UserMinimalSelect: FC<Props> = ({ user = '', onSelect }) => {
    const { data: users } = useGetUsersMinimal();

    return (
        <div className='flex flex-col gap-2'>
            <Label>Users</Label>

            <Select onValueChange={onSelect} value={user}>
                <SelectTrigger className='w-32' aria-label='User Select'>
                    <SelectValue placeholder='Select' />
                </SelectTrigger>
                <SelectPortal>
                    <SelectContent>
                        <SelectItem className='italic' key='user-unselect' value='-none-'>
                            None
                        </SelectItem>
                        {users?.map((item: User) => (
                            <SelectItem key={`user-${item.id}`} value={item.id}>
                                {item.email_address}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </SelectPortal>
            </Select>
        </div>
    );
};
