import { Checkbox, ColumnDef, DataTable, Input } from '@bloodhoundenterprise/doodleui';
import { CheckedState } from '@radix-ui/react-checkbox';
import { User } from 'js-client-library';
import { useEffect, useMemo, useState } from 'react';
import { useQuery } from 'react-query';
import { AppIcon } from '../../../components';
import { useQueryPermissions } from '../../../hooks';
import { useSelf } from '../../../hooks/useSelf';
import { apiClient } from '../../../utils';

type SavedQueryPermissionsProps = {
    queryId?: number;
    sharedIds: string[];
    isPublic: boolean;
    setSharedIds: (ids: string[]) => void;
    setIsPublic: (isPublic: boolean) => void;
};
type ListUser = {
    name: string;
    id: string;
};

const SavedQueryPermissions: React.FC<SavedQueryPermissionsProps> = (props: SavedQueryPermissionsProps) => {
    const { isPublic, queryId, sharedIds, setSharedIds, setIsPublic } = props;
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [filteredUsers, setFilteredUsers] = useState<ListUser[]>([]);

    const { getSelfId } = useSelf();
    const { data: selfId } = getSelfId;

    const listUsersQuery = useQuery(['listUsers'], ({ signal }) =>
        apiClient.listUsers({ signal }).then((res) => res.data?.data?.users)
    );

    const { data, isLoading } = useQueryPermissions(queryId as number);

    function idMap() {
        return listUsersQuery.data
            ?.filter((user: User) => user.id !== selfId)
            .map((user: User) => {
                return {
                    name: user.principal_name,
                    id: user.id,
                };
            });
    }

    const usersList = useMemo(() => idMap(), [listUsersQuery.data]);
    const allUserIds = useMemo(() => usersList?.map((x) => x.id), [listUsersQuery.data]);

    useEffect(() => {
        if (data?.shared_to_user_ids.length) {
            setSharedIds([...data?.shared_to_user_ids]);
        } else {
            setSharedIds([]);
        }
        if (data?.public) {
            setIsPublic(true);
            setSharedIds(allUserIds as string[]);
        } else {
            setIsPublic(false);
        }
    }, [data, allUserIds]);

    const handleCheckAllChange = (checkedState: CheckedState) => {
        setIsPublic(checkedState as boolean);
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
            setFilteredUsers(filtered as ListUser[]);
        } else {
            setFilteredUsers([]);
        }
    }, [searchTerm]);

    return (
        <>
            {isLoading && <div>Loading ...</div>}
            {usersList?.length ? (
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
                        <DataTable
                            TableHeadProps={{
                                className: 'text-s font-bold first:!w-8 pl-3 first:pl-0 first:text-center',
                            }}
                            TableBodyProps={{ className: 'text-s font-roboto underline' }}
                            TableCellProps={{ className: 'first:!w-8 pl-3 first:pl-0 first:text-center' }}
                            columns={getColumns()}
                            data={filteredUsers.length ? filteredUsers : usersList}
                        />
                    </div>
                </div>
            ) : (
                <div className='flex flex-col  py-8'>There are currently no users on this account.</div>
            )}
        </>
    );
};

export default SavedQueryPermissions;
