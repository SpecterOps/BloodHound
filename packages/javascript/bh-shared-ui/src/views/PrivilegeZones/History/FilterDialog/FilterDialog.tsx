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
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogTitle,
    DialogTrigger,
    Form,
    Tooltip,
    VisuallyHidden,
} from '@bloodhoundenterprise/doodleui';
import { DateTime } from 'luxon';
import { FC, useCallback, useEffect, useState } from 'react';
import { ErrorOption, useForm } from 'react-hook-form';
import { AppIcon } from '../../../../components';
import { CustomRangeError, END_DATE, LuxonFormat, START_DATE } from '../../../../utils';
import { useHistoryTableContext } from '../HistoryTableContext';
import { AssetGroupTagHistoryFilters } from '../types';
import { DEFAULT_FILTER_VALUE } from '../utils';
import ActionField from './ActionField';
import { EndDateField, StartDateField } from './DateField';
import MadeByField from './MadeByField';
import TagIdField from './TagIdField';

const FilterDialog: FC<{
    setFilters: (filters: AssetGroupTagHistoryFilters) => void;
    filters?: AssetGroupTagHistoryFilters;
}> = ({ filters = DEFAULT_FILTER_VALUE, setFilters = () => {} }) => {
    const { clearSelected } = useHistoryTableContext();
    const [open, setOpen] = useState(false);

    const form = useForm<AssetGroupTagHistoryFilters>({ defaultValues: DEFAULT_FILTER_VALUE });

    const validateDateFields = useCallback(
        (startDate: DateTime, endDate: DateTime) => {
            form.clearErrors();
            const errors: { name: typeof START_DATE | typeof END_DATE; error: ErrorOption }[] = [];

            if (!startDate.isValid) {
                errors.push({ name: START_DATE, error: { message: CustomRangeError.INVALID_DATE } });
            }
            if (!endDate.isValid) {
                errors.push({ name: END_DATE, error: { message: CustomRangeError.INVALID_DATE } });
            }
            if (errors.length === 0 && startDate > endDate) {
                errors.push({ name: START_DATE, error: { message: CustomRangeError.INVALID_RANGE_START } });
                errors.push({ name: END_DATE, error: { message: CustomRangeError.INVALID_RANGE_END } });
            }

            if (errors.length > 0) {
                errors.forEach((error) => form.setError(error.name, error.error));
                return false;
            } else {
                form.clearErrors();
                return true;
            }
        },
        [form]
    );

    const handleConfirm = useCallback(() => {
        const start = form.getValues(START_DATE);
        const end = form.getValues(END_DATE);
        // If the start date is empty use the start of epoch time
        const startDate = start !== '' ? DateTime.fromFormat(start, LuxonFormat.ISO_8601) : DateTime.fromMillis(0);
        // Use the client time if the end date is empty
        const endDate = end !== '' ? DateTime.fromFormat(end, LuxonFormat.ISO_8601) : DateTime.now();

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
                <Button
                    data-testid='privilege-zones_history_filter-button'
                    variant='text'
                    aria-label='Filter'
                    onClick={() => {
                        setOpen((prev) => !prev);
                    }}>
                    <Tooltip tooltip='Filters'>
                        <span>
                            <AppIcon.FilterOutline size={22} />
                        </span>
                    </Tooltip>
                </Button>
            </DialogTrigger>

            <DialogContent>
                <Form {...form}>
                    <form className='flex flex-col gap-4'>
                        <DialogTitle className='flex justify-between items-center'>
                            <span className='text-xl'>Filter</span>
                            <Button
                                variant={'text'}
                                onClick={() => form.reset(DEFAULT_FILTER_VALUE)}
                                className='font-normal p-2'>
                                Clear All
                            </Button>
                        </DialogTitle>
                        <VisuallyHidden asChild>
                            <DialogDescription>Filter Privilege Zone History</DialogDescription>
                        </VisuallyHidden>

                        <ActionField form={form} />

                        <TagIdField form={form} />

                        <MadeByField form={form} />

                        <div className='flex gap-6'>
                            <StartDateField form={form} />
                            <EndDateField form={form} />
                        </div>

                        <DialogActions>
                            <Button variant={'text'} className='p-2' onClick={closeDialog}>
                                Cancel
                            </Button>
                            <Button
                                variant={'text'}
                                className='text-primary dark:text-secondary-variant-2 p-2'
                                onClick={handleConfirm}>
                                Confirm
                            </Button>
                        </DialogActions>
                    </form>
                </Form>
            </DialogContent>
        </Dialog>
    );
};

export default FilterDialog;
