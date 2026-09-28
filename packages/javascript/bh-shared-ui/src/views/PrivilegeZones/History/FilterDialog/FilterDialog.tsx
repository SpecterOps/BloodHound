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
import {
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogTitle,
    DialogTrigger,
    Form,
    IconButton,
    TextButton,
    Tooltip,
    VisuallyHidden,
} from 'doodle-ui';
import { DateTime } from 'luxon';
import { FC, useCallback, useEffect, useState } from 'react';
import { UseFormReturn, useForm } from 'react-hook-form';
import { getStartAndEndDateTimes, validateFormDates } from '../..';
import { END_DATE, START_DATE } from '../../../..';
import { AppIcon } from '../../../../components';
import { EndDateField, StartDateField, TagIdField } from '../../Filters';
import { useHistoryTableContext } from '../HistoryTableContext';
import { AssetGroupTagHistoryFilters } from '../types';
import { DEFAULT_FILTER_VALUE } from '../utils';
import ActionField from './ActionField';
import MadeByField from './MadeByField';

const FilterDialog: FC<{
    setFilters: (filters: AssetGroupTagHistoryFilters) => void;
    filters?: AssetGroupTagHistoryFilters;
}> = ({ filters = DEFAULT_FILTER_VALUE, setFilters = () => {} }) => {
    const { clearSelected } = useHistoryTableContext();
    const [open, setOpen] = useState(false);

    const form = useForm<AssetGroupTagHistoryFilters>({ defaultValues: DEFAULT_FILTER_VALUE });

    const validateDateFields = useCallback(
        (startDate: DateTime, endDate: DateTime) => validateFormDates(form, startDate, endDate)(),
        [form]
    );

    const handleConfirm = useCallback(() => {
        const values = form.getValues();
        const { startDate, endDate } = getStartAndEndDateTimes(values[START_DATE], values[END_DATE]);

        // Prevent setting invalid dates before applying filters, e.g., bogus date like 9999/99/99 or a range where the start date is after the end date
        if (validateDateFields(startDate, endDate)) {
            setFilters({ ...form.getValues() });
            clearSelected();
            closeDialog();
        }
    }, [form, setFilters, validateDateFields, clearSelected]);

    const closeDialog = () => setOpen(false);

    useEffect(() => {
        form.reset(filters);
    }, [form, filters]);

    return (
        <Dialog
            open={open}
            onOpenChange={(open) => {
                setOpen(open);
            }}>
            <DialogTrigger asChild>
                <IconButton
                    data-testid='privilege-zones_history_filter-button'
                    className='ml-4'
                    size={24}
                    aria-label='Filter'
                    onClick={() => {
                        setOpen((prev) => !prev);
                    }}>
                    <Tooltip tooltip='Filters'>
                        <AppIcon.FilterOutline />
                    </Tooltip>
                </IconButton>
            </DialogTrigger>

            <DialogContent>
                <Form {...form}>
                    <form className='flex flex-col gap-4 m-1'>
                        <DialogTitle className='flex justify-between items-center'>
                            <span className='text-xl'>Filter</span>
                            <TextButton
                                fontColor='primary'
                                onClick={() => form.reset(DEFAULT_FILTER_VALUE)}
                                className='font-bold'>
                                Clear All
                            </TextButton>
                        </DialogTitle>
                        <VisuallyHidden asChild>
                            <DialogDescription>Filter Privilege Zone History</DialogDescription>
                        </VisuallyHidden>

                        <ActionField form={form} />

                        <TagIdField form={form as unknown as UseFormReturn} fieldLabel='Zone/Label' />

                        <MadeByField form={form} />

                        <div className='flex gap-6'>
                            <StartDateField form={form as unknown as UseFormReturn} />
                            <EndDateField form={form as unknown as UseFormReturn} />
                        </div>

                        <DialogActions className='gap-2'>
                            <TextButton onClick={closeDialog}>Cancel</TextButton>

                            <TextButton
                                fontColor='primary'
                                data-testid='file_ingest_log-filter_dialog_confirm'
                                onClick={handleConfirm}>
                                Confirm
                            </TextButton>
                        </DialogActions>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
};

export default FilterDialog;
