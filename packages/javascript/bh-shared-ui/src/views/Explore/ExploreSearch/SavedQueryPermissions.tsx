import { Checkbox, ColumnDef, DataTable, Input } from '@bloodhoundenterprise/doodleui';
import { CheckedState } from '@radix-ui/react-checkbox';
import { useEffect, useMemo, useState } from 'react';
import { useQuery, useQueryClient } from 'react-query';
import { AppIcon } from '../../../components';
import { useQueryPermissions } from '../../../hooks';
import { apiClient } from '../../../utils';

import {} from 'react-query';
type SavedQueryPermissionsProps = {
    queryId?: number;
    sharedIds: string[];
    setSharedIds: (ids: string[]) => void;
};

const SavedQueryPermissions: React.FC<SavedQueryPermissionsProps> = (props: SavedQueryPermissionsProps) => {
    const { queryId, sharedIds, setSharedIds } = props;
    const [shareAll, setShareAll] = useState<boolean>(false);
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [filteredUsers, setFilteredUsers] = useState<any[]>([]);

    const getSelf = useQuery(['getSelf'], ({ signal }) => apiClient.getSelf({ signal }).then((res) => res.data.data));

    const listUsersQuery = useQuery(['listUsers'], ({ signal }) =>
        apiClient.listUsers({ signal }).then((res) => res.data?.data?.users)
    );

    const { data, isLoading, error, isError } = useQueryPermissions(queryId as number);

    function idMap() {
        return listUsersQuery.data
            ?.filter((user: any) => user.id !== getSelf.data.id)
            .map((user: any) => {
                return {
                    name: user.principal_name,
                    id: user.id,
                };
            });
    }

    const usersList = useMemo(() => idMap(), [listUsersQuery.data]);
    const allUserIds = useMemo(() => usersList?.map((x) => x.id), [listUsersQuery.data]);

    const queryClient = useQueryClient();

    useEffect(() => {
        // manually setting data on error.
        // api returns error for empty state.
        queryClient.setQueryData(['permissions'], (oldData: any) => {
            return { ...oldData, shared_to_user_ids: [] };
        });
    }, [error, isError]);

    useEffect(() => {
        if (data?.shared_to_user_ids.length) {
            setSharedIds(data?.shared_to_user_ids);
        } else {
            setSharedIds([]);
        }
        if (data?.shared_to_user_ids.length && data?.shared_to_user_ids.length === allUserIds?.length) {
            setShareAll(true);
        } else {
            setShareAll(false);
        }
    }, [data, allUserIds]);

    const handleCheckAllChange = (checkedState: CheckedState) => {
        setShareAll(checkedState as boolean);
        if (checkedState) {
            setSharedIds(allUserIds as string[]);
        } else {
            setSharedIds([]);
        }
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
    };

    const isCheckboxChecked = (id: any) => {
        return sharedIds.includes(id);
    };

    const getColumns = () => {
        const columns: ColumnDef<any>[] = [
            {
                accessorKey: 'id',
                header: () => {
                    return (
                        <div className=''>
                            <Checkbox className='' checked={shareAll} onCheckedChange={handleCheckAllChange} />
                        </div>
                    );
                },
                cell: ({ row }) => (
                    <div className=''>
                        <Checkbox
                            className=''
                            checked={isCheckboxChecked(row.getValue('id'))}
                            onCheckedChange={() => handleCheckChange(row.getValue('id'))}
                        />
                    </div>
                ),
            },
            {
                accessorKey: 'name',
                header: () => {
                    return <span className='dark:text-neutral-light-1'>Name</span>;
                },
                cell: ({ row }) => {
                    return (
                        <div className='dark:text-neutral-light-1 text-nowrap text-black w-full'>
                            {row.getValue('name')}
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

    useEffect(() => {
        if (searchTerm.length) {
            const filtered = usersList?.filter((user) => user.name.toLowerCase().includes(searchTerm.toLowerCase()));
            setFilteredUsers(filtered as any[]);
        } else {
            setFilteredUsers([]);
        }
    }, [searchTerm]);

    return (
        <>
            {isLoading && <div>Loading ...</div>}
            {usersList?.length && (
                <div>
                    <>
                        <div className='flex-grow relative mb-2'>
                            <AppIcon.MagnifyingGlass size={16} className='absolute left-5 top-[50%] -mt-[8px]' />
                            <Input
                                type='text'
                                id='query-search'
                                placeholder='Search'
                                value={searchTerm}
                                className='w-full bg-transparent dark:bg-transparent rounded-none border-neutral-dark-5 border-t-0 border-x-0 pl-12'
                                onChange={(event: React.ChangeEvent<HTMLInputElement>) =>
                                    handleInput(event.target.value)
                                }
                            />
                        </div>
                        <DataTable
                            TableProps={{ className: '' }}
                            TableHeadProps={{
                                className: 'text-s font-bold first:!w-8 pl-3 first:pl-0 first:text-center',
                            }}
                            TableBodyProps={{ className: 'text-s font-roboto underline' }}
                            TableBodyRowProps={{ className: '' }}
                            TableCellProps={{ className: 'first:!w-8 pl-3 first:pl-0 first:text-center' }}
                            columns={getColumns()}
                            data={filteredUsers.length ? filteredUsers : usersList}
                        />
                    </>
                </div>
            )}

            {!usersList?.length && (
                <div className='flex flex-col  py-8'>There are currently no users on this account.</div>
            )}
        </>
    );
};

export default SavedQueryPermissions;
