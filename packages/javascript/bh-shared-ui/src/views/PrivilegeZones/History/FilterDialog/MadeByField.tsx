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
import { faClose } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Button,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    Skeleton,
} from 'doodle-ui';
import { BloodHoundString, User } from 'js-client-library';
import { FC, useMemo } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { useBloodHoundUsers } from '../../../../hooks/useBloodHoundUsers';
import { cn } from '../../../../utils';
import { AssetGroupTagHistoryFilters } from '../types';

const filterAndSortUsers = (users: User[]) => {
    return users
        .filter((user) => {
            let hasPermission = false;
            user.roles.forEach((role) => {
                role.permissions.forEach((permission) => {
                    if (permission.name === 'ManageUsers' && permission.authority === 'auth') hasPermission = true;
                });
            });

            return hasPermission;
        })
        .sort((a, b) => (a.email_address || a.principal_name).localeCompare(b.email_address || b.principal_name));
};

const MadeByField: FC<{
    form: UseFormReturn<AssetGroupTagHistoryFilters>;
}> = ({ form }) => {
    const bloodHoundUsersQuery = useBloodHoundUsers();
    const users = useMemo(() => filterAndSortUsers(bloodHoundUsersQuery.data ?? []), [bloodHoundUsersQuery.data]);

    return (
        <FormField
            control={form.control}
            name='madeBy'
            render={({ field }) => (
                <FormItem>
                    <FormLabel aria-labelledby='madeBy'>Made By</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value} defaultValue={field.value}>
                        <div className='flex gap-2'>
                            <FormControl className='w-11/12'>
                                {bloodHoundUsersQuery.isError ? (
                                    <span className='text-error'>
                                        There was an error fetching this data. Please refresh the page to try again.
                                    </span>
                                ) : (
                                    <SelectTrigger>
                                        <SelectValue placeholder='Select' />
                                    </SelectTrigger>
                                )}
                            </FormControl>
                            <Button
                                variant={'text'}
                                disabled={!field.value}
                                className={cn('w-1/12 p-0', { invisible: !field.value })}
                                onClick={() => {
                                    form.setValue(field.name, '');
                                }}>
                                <FontAwesomeIcon icon={faClose} />
                            </Button>
                        </div>
                        {bloodHoundUsersQuery.isLoading ? (
                            <Skeleton className='h-10 w-24' />
                        ) : (
                            <SelectContent>
                                <SelectItem value={BloodHoundString}>{BloodHoundString}</SelectItem>
                                {users.map((user) => (
                                    <SelectItem key={user.id} value={user.email_address || user.id}>
                                        {user.email_address || user.principal_name}
                                    </SelectItem>
                                ))}
                            </SelectContent>
                        )}
                    </Select>
                </FormItem>
            )}
        />
    );
};

export default MadeByField;
