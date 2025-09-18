import {
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
} from '@bloodhoundenterprise/doodleui';
import type { FC } from 'react';

type Props = {
    client?: string;
    onSelect: (value: string) => void;
};

export const UserSelect: FC<Props> = ({ client = '', onSelect }) => {
    const clients = { data: [{ id: 'test', name: 'test' }] };
    return (
        <div className='flex flex-col gap-2'>
            <Label>Users</Label>

            <Select onValueChange={onSelect} value={client}>
                <SelectTrigger className='w-32' aria-label='Client Select'>
                    <SelectValue placeholder='Select' />
                </SelectTrigger>
                <SelectPortal>
                    <SelectContent>
                        <SelectItem className='italic' key='client-unselect' value='-none-'>
                            None
                        </SelectItem>
                        {clients?.data.map((item) => (
                            <SelectItem key={`client-${item.id}`} value={item.id}>
                                {item.name}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </SelectPortal>
            </Select>
        </div>
    );
};
