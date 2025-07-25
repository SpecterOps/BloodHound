import { CheckedState } from '@radix-ui/react-checkbox';
import { useEffect, useMemo, useState } from 'react';
import { useQuery } from 'react-query';
import { AppIcon } from '../../../components';
import { useDeleteQueryPermissions, useQueryPermissions, useUpdateQueryPermissions } from '../../../hooks';

import { apiClient } from '../../../utils';

import { Checkbox, ColumnDef, DataTable, Input } from '@bloodhoundenterprise/doodleui';
type SavedQueryPermissionsProps = {
    queryId: number;
};

const SavedQueryPermissions: React.FC<SavedQueryPermissionsProps> = (props: SavedQueryPermissionsProps) => {
    const { queryId } = props;
    // const [sharedIds, setSharedIds] = useState<string[]>([]);
    const [shareAll, setShareAll] = useState<boolean>(false);
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [filteredUsers, setFilteredUsers] = useState<any[]>([]);

    const updateQueryPermissionsMutation = useUpdateQueryPermissions();

    const getSelf = useQuery(['getSelf'], ({ signal }) => apiClient.getSelf({ signal }).then((res) => res.data.data));

    const listUsersQuery = useQuery(['listUsers'], ({ signal }) =>
        apiClient.listUsers({ signal }).then((res) => res.data?.data?.users)
    );

    // const { data, error, isLoading } = useQuery<any, any>({
    //     queryFn: () => getQueryPermissions(queryId),
    //     // enabled: false,
    //     refetchOnWindowFocus: false,
    //     retry: false,
    // });

    const { data, error, isLoading } = useQueryPermissions(queryId);
    const { shared_to_user_ids: sharedIds } = data || [];

    const deletePermissionsMutation = useDeleteQueryPermissions();

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

    useEffect(() => {
        if (error) {
            console.log('error');
            // console.log(error.response.data.errors);
            // // console.log(error);
            // console.log(
            //     error.response.data.errors.some((item: any) =>
            //         item.message.includes('no query permissions exist for saved query')
            //     )
            // );
            // setSharedIds([]);
        }
    }, [error]);

    useEffect(() => {
        if (data) {
            console.log('data');
            console.log(data);
            // console.log(data.data.shared_to_user_ids);
            // setSharedIds(data.shared_to_user_ids);
        }
    }, [data]);

    const handleCheckAllChange = (checkedState: CheckedState) => {
        if (checkedState) {
            updateQueryPermissionsMutation.mutate({
                id: queryId,
                payload: {
                    user_ids: allUserIds as string[],
                    public: false,
                },
            });
        } else {
            deletePermissionsMutation.mutate({
                id: queryId,
                payload: {
                    user_ids: allUserIds as string[],
                },
            });
        }
    };

    const handleCheckChange = (sharedUserId: string) => {
        if (sharedIds?.includes(sharedUserId)) {
            //delete

            deletePermissionsMutation.mutate({
                id: queryId,
                payload: {
                    user_ids: [sharedUserId],
                },
            });
        } else {
            //add
            updateQueryPermissionsMutation.mutate({
                id: queryId,
                payload: {
                    user_ids: [...sharedIds, sharedUserId],
                    public: false,
                },
            });
        }
    };

    const getColumns = () => {
        const columns: ColumnDef<any>[] = [
            {
                accessorKey: 'id',
                header: () => {
                    return (
                        <div className='min-w-12 max-w-12'>
                            <Checkbox className='ml-4' checked={shareAll} onCheckedChange={handleCheckAllChange} />
                        </div>
                    );
                },
                cell: ({ row }) => (
                    <>
                        <div className='min-w-12 max-w-12'>
                            <Checkbox
                                className='ml-4'
                                checked={sharedIds?.includes(row.getValue('id'))}
                                onCheckedChange={() => handleCheckChange(row.getValue('id'))}
                            />
                        </div>
                    </>
                ),
            },
            {
                accessorKey: 'name',
                header: () => {
                    return <span className='dark:text-neutral-light-1'>Name</span>;
                },
                cell: ({ row }) => {
                    return (
                        <div className='dark:text-neutral-light-1 text-nowrap text-black w-full min-w-36 max-w-36'>
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
        <div>
            {isLoading && <div>Loading ...</div>}
            {usersList && (
                <>
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
                    <DataTable
                        TableProps={{ className: '' }}
                        TableHeadProps={{ className: 'text-s font-bold' }}
                        TableBodyProps={{ className: 'text-s font-roboto underline' }}
                        TableBodyRowProps={{ className: '' }}
                        columns={getColumns()}
                        data={filteredUsers.length ? filteredUsers : usersList}
                    />
                </>
            )}
        </div>
    );
};

export default SavedQueryPermissions;
