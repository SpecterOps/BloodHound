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
import { JOB_STATUS_MAP } from '../utils';

type Props = {
    statusOptions: string[];
    status?: number;
    onSelect: (value: string) => void;
};

export const StatusSelect: FC<Props> = ({ status = '', statusOptions, onSelect }) => {
    const STATUS_FILTERS = Object.entries(JOB_STATUS_MAP).filter(([, value]) => statusOptions.includes(value));

    return (
        <div className='flex flex-col gap-2'>
            <Label>Status</Label>

            <Select onValueChange={onSelect} value={status.toString()}>
                <SelectTrigger className='w-32' aria-label='Status Select'>
                    <SelectValue placeholder='Select' />
                </SelectTrigger>
                <SelectPortal>
                    <SelectContent>
                        <SelectItem className='italic' key='status-unselect' value='-none-'>
                            None
                        </SelectItem>
                        {STATUS_FILTERS.map(([id, value]) => (
                            <SelectItem key={`status-${id}`} value={id.toString()}>
                                {value}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </SelectPortal>
            </Select>
        </div>
    );
};
