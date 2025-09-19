import {
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
} from '@bloodhoundenterprise/doodleui';
import { User } from 'js-client-library';
import type { FC } from 'react';
import { useQuery } from 'react-query';
import { apiClient } from '../../utils';

type Props = {
    user?: string;
    onSelect: (value: string) => void;
};

export const UserSelect: FC<Props> = ({ user = '', onSelect }) => {
    const { data: users } = useQuery({
        queryKey: ['users-minimal'],
        queryFn: ({ signal }) => apiClient.listUsersMinimal({ signal }).then((res) => res.data),
        select: (data) => data.data.users,
    });

    return (
        <div className='flex flex-col gap-2'>
            <Label>Users</Label>

            <Select onValueChange={onSelect} value={user}>
                <SelectTrigger className='w-32' aria-label='Client Select'>
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
