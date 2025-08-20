import {
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
} from '@bloodhoundenterprise/doodleui';
import type { FC } from 'react';
import type { Control, FieldValues } from 'react-hook-form';
import { JOB_STATUS_MAP } from '../../utils';

export const StatusSelect: FC<{ control: Control<FieldValues, any, FieldValues> }> = ({ control }) => {
    return (
        <FormField
            control={control}
            name='status'
            render={({ field }) => (
                <FormItem>
                    <FormLabel>Status</FormLabel>

                    <FormControl>
                        <Select onValueChange={field.onChange}>
                            <SelectTrigger className='w-32'>
                                <SelectValue placeholder='Select' />
                            </SelectTrigger>
                            <SelectPortal>
                                <SelectContent>
                                    {Object.entries(JOB_STATUS_MAP).map(([key, value]) => (
                                        <SelectItem key={key} value={key}>
                                            {value.label}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </SelectPortal>
                        </Select>
                    </FormControl>
                </FormItem>
            )}
        />
    );
};
