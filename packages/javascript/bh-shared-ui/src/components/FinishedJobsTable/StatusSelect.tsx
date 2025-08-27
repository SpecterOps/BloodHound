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
import { JOB_STATUS_MAP } from '../../utils';

// 2 = Complete, 5 = Failed
const FILTERABLE_STATUSES = ['Complete', 'Failed'];
const STATUS_FILTERS = Object.entries(JOB_STATUS_MAP).filter(([, value]) => FILTERABLE_STATUSES.includes(value));

type Props = {
    status?: string;
    onSelect: (value: string) => void;
};

export const SELECT_NONE = '-none-';

export const StatusSelect: FC<Props> = ({ status = '', onSelect }) => {
    return (
        <div className='flex flex-col gap-2'>
            <Label htmlFor='status'>Status</Label>

            <Select onValueChange={onSelect} value={status}>
                <SelectTrigger className='w-32' aria-label='Status Select'>
                    <SelectValue placeholder='Select' />
                </SelectTrigger>
                <SelectPortal>
                    <SelectContent>
                        <SelectItem className='italic' key='status-unselect' value={SELECT_NONE}>
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
