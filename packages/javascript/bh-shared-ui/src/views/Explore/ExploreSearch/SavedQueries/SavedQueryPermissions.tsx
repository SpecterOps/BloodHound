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
import { Button, Checkbox, ColumnDef, DataTable, Input } from '@bloodhoundenterprise/doodleui';
import { CheckedState } from '@radix-ui/react-checkbox';
import { UserMinimal } from 'js-client-library';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useQuery } from 'react-query';
import { AppIcon } from '../../../../components';
import { useQueryPermissions } from '../../../../hooks';
import { useSelf } from '../../../../hooks/useSelf';
import { apiClient } from '../../../../utils';
import { useSavedQueriesContext } from '../../providers';

type SavedQueryPermissionsProps = {
    sharedIds: string[];
    isPublic: boolean;
    setSharedIds: (ids: string[]) => void;
    setIsPublic: (isPublic: boolean) => void;
};
type ListUser = {
    name: string;
    id: string;
    email: string;
};

const SavedQueryPermissions: React.FC<SavedQueryPermissionsProps> = (props: SavedQueryPermissionsProps) => {
    const { isPublic, sharedIds, setSharedIds, setIsPublic } = props;
    const { selectedQuery } = useSavedQueriesContext();
    const queryId = selectedQuery?.id;

    const [searchTerm, setSearchTerm] = useState<string>('');

    const { getSelfId } = useSelf();
    const { data: selfId } = getSelfId;

    const listUsersQuery = useQuery(['listUsersMinimal'], ({ signal }) =>
        apiClient.listUsersMinimal({ signal }).then((res) => {
            return res.data?.data?.users;
        })
    );

    const { data, isLoading } = useQueryPermissions(queryId as number);

    const idMap = useCallback(() => {
        return listUsersQuery.data
            ?.filter((user: UserMinimal) => user.id !== selfId)
            .map((user: UserMinimal) => {
                return {
                    id: user.id,
                    name: `${user.first_name} ${user.last_name}`,
                    email: user.email_address,
                };
            });
    }, [listUsersQuery.data, selfId]);

    const usersList = useMemo(() => idMap(), [idMap]);
    const allUserIds = useMemo(() => usersList?.map((x) => x.id) ?? [], [usersList]);

    useEffect(() => {
        if (!data) return;
        const initialShared = data.public ? allUserIds : data.shared_to_user_ids ?? [];
        setSharedIds(initialShared);
        setIsPublic(Boolean(data.public));
    }, [data, allUserIds, setSharedIds, setIsPublic]);

    const handleCheckAllChange = (checkedState: CheckedState) => {
        const isTrue = checkedState === true;
        setIsPublic(isTrue);
        setSharedIds(isTrue ? allUserIds : []);
    };

    const handleCheckChange = (sharedUserId: string) => {
        //New query - no queryId present
        if (sharedIds.includes(sharedUserId)) {
            //delete
            setSharedIds(sharedIds.filter((item) => item !== sharedUserId));
        } else {
            // add
            setSharedIds([...sharedIds, sharedUserId]);
        }
        setIsPublic(false);
    };

    const isCheckboxChecked = (id: string) => {
        return sharedIds?.includes(id);
    };

    const getColumns = () => {
        const columns: ColumnDef<ListUser>[] = [
            {
                accessorKey: 'id',
                header: () => {
                    return (
                        <div>
                            <Checkbox
                                checked={isPublic}
                                onCheckedChange={handleCheckAllChange}
                                data-testid='public-query'
                            />
                        </div>
                    );
                },
                cell: ({ row }) => (
                    <div>
                        <Checkbox
                            checked={isCheckboxChecked(row.getValue('id'))}
                            onCheckedChange={() => handleCheckChange(row.getValue('id'))}
                        />
                    </div>
                ),
            },
            {
                accessorKey: 'name',
                header: () => {
                    return <span className='dark:text-neutral-light-1 font-normal'>Set to Public</span>;
                },
                cell: ({ row }) => {
                    const name = row.original.name;
                    const email = row.original.email;
                    return (
                        <div className='dark:text-neutral-light-1 text-nowrap text-black w-full'>
                            <p className='underline mb-0.5'>{name}</p>
                            <p className='text-neutral-600 dark:!text-neutral-300'>{email}</p>
                        </div>
                    );
                },
            },
        ];
        return columns;
    };

    const handleInput = (searchTerm: string) => {
        setSearchTerm(searchTerm);
    };

    const filteredUsers = useMemo(() => {
        if (!searchTerm) return usersList;
        const filtered = usersList?.filter((user) => user.name.toLowerCase().includes(searchTerm.toLowerCase()));
        return filtered;
    }, [searchTerm, usersList]);

    const resetSearch = () => {
        setSearchTerm('');
    };

    return (
        <>
            {isLoading || listUsersQuery.isLoading ? (
                <div>Loading ...</div>
            ) : usersList?.length ? (
                <div>
                    <div className='flex-grow relative mb-2'>
                        <AppIcon.MagnifyingGlass size={16} className='absolute left-5 top-[50%] -mt-[8px]' />
                        <Input
                            type='text'
                            id='query-search'
                            placeholder='Search'
                            value={searchTerm}
                            className='w-full bg-transparent dark:bg-transparent rounded-none border-neutral-dark-5 border-t-0 border-x-0 pl-12'
                            onChange={(event: React.ChangeEvent<HTMLInputElement>) => handleInput(event.target.value)}
                        />
                    </div>
                    <div className='h-[335px] overflow-auto'>
                        {filteredUsers?.length ? (
                            <DataTable
                                TableHeadProps={{
                                    className: 'text-s first:!w-8 pl-3 first:pl-0 first:text-center',
                                }}
                                TableBodyProps={{ className: 'text-s font-roboto' }}
                                TableCellProps={{ className: 'first:!w-8 pl-3 first:pl-0 first:text-center' }}
                                columns={getColumns() as ListUser[]}
                                data={filteredUsers}
                            />
                        ) : (
                            <QueryPermissionsEmpty resetSearch={resetSearch} />
                        )}
                    </div>
                </div>
            ) : (
                <div className='flex flex-col py-8 px-2'>There are currently no users on this account.</div>
            )}
        </>
    );
};

type QueryPermissionsEmptyProps = {
    resetSearch: () => void;
};
const QueryPermissionsEmpty = (props: QueryPermissionsEmptyProps) => {
    const { resetSearch } = props;
    return (
        <div className='flex flex-col py-8 px-2 items-center'>
            <p className='mb-6'>No users match this search term.</p>
            <Button variant='primary' size='small' onClick={resetSearch}>
                Reset Search
            </Button>
        </div>
    );
};

export default SavedQueryPermissions;
