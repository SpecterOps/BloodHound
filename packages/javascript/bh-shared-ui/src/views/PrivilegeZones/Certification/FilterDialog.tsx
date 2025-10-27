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
    DatePicker,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    DialogTrigger,
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
    Skeleton,
    VisuallyHidden,
} from '@bloodhoundenterprise/doodleui';
import { faClose } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { AssetGroupTagCertificationRecord, AssetGroupTagTypeZone } from 'js-client-library';
import { DateTime } from 'luxon';
import { FC, useCallback, useEffect, useState } from 'react';
import { UseFormReturn, useForm } from 'react-hook-form';
import { AppIcon, MaskedInput } from '../../../components';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../../graphSchema';
import { useGetUsersMinimal } from '../../../hooks/useGetUsers';
import { END_DATE, LuxonFormat, START_DATE, cn } from '../../../utils';
import TagIdField from '../History/FilterDialog/TagIdField';
import { fromDate, getStartAndEndDateTimes, toDate, validateFormDates } from '../utils';
import { defaultFilterValues } from './constants';
import { FilterFormValues } from './types';

interface FilterDialogProps {
    filters: FilterFormValues;
    onApplyFilters: (filters: FilterFormValues) => void;
    data: AssetGroupTagCertificationRecord[];
}

//TODO: we need to consolidate this into one universal shared component but separating in the interest of time
const FilterDialog: FC<FilterDialogProps> = ({ filters, onApplyFilters }) => {
    const bloodHoundUsersQuery = useGetUsersMinimal();
    const [open, setOpen] = useState(false);

    const form = useForm<FilterFormValues>({
        defaultValues: defaultFilterValues,
    });

    const allObjectTypes = [...Object.values(AzureNodeKind), ...Object.values(ActiveDirectoryNodeKind)];

    const validateDates = useCallback(
        (startDate: DateTime, endDate: DateTime) => validateFormDates(form, startDate, endDate)(),
        [form]
    );

    const handleConfirm = useCallback(() => {
        const values = form.getValues();
        const { startDate, endDate } = getStartAndEndDateTimes(values[START_DATE], values[END_DATE]);

        if (validateDates(startDate, endDate)) {
            onApplyFilters(values); // parent table only updated on Confirm
            closeDialog();
        }
    }, [form, onApplyFilters, validateDates]);

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
                    variant='text'
                    data-testid='certifications_filter_dialog'
                    onClick={() => {
                        setOpen((prev) => !prev);
                    }}>
                    <AppIcon.FilterOutline size={22} />
                </Button>
            </DialogTrigger>
            <DialogPortal>
                <DialogContent>
                    <Form {...form}>
                        <form className='flex flex-col gap-4'>
                            <DialogTitle className='flex justify-between items-center'>
                                <span className='text-xl'>Additional Filters</span>
                                <Button
                                    variant='text'
                                    onClick={() => {
                                        form.reset(defaultFilterValues);
                                    }}
                                    className='font-normal p-2'>
                                    Clear All
                                </Button>
                            </DialogTitle>
                            <VisuallyHidden asChild>
                                <DialogDescription>Filter Privilege Zone Certifications</DialogDescription>
                            </VisuallyHidden>
                            <TagIdField
                                fieldLabel='Zone'
                                form={form as unknown as UseFormReturn}
                                tagSelect={(data) => data.filter((tag) => tag.type === AssetGroupTagTypeZone)}
                            />
                            <FormField
                                name='objectType'
                                control={form.control}
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Object Type</FormLabel>
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <div className='flex gap-2'>
                                                <FormControl>
                                                    <SelectTrigger>
                                                        <SelectValue placeholder='Select Object Type' />
                                                    </SelectTrigger>
                                                </FormControl>
                                                <SelectPortal>
                                                    <SelectContent>
                                                        {allObjectTypes.map((objType) => (
                                                            <SelectItem key={objType} value={objType}>
                                                                {objType}
                                                            </SelectItem>
                                                        ))}
                                                    </SelectContent>
                                                </SelectPortal>
                                                <Button
                                                    variant={'text'}
                                                    disabled={!field.value}
                                                    className={cn('w-1/12 p-0', { invisible: !field.value })}
                                                    onClick={() => {
                                                        form.setValue(field.name, '');
                                                    }}>
                                                    <FontAwesomeIcon icon={faClose} />
                                                </Button>
                                            </div>
                                        </Select>
                                    </FormItem>
                                )}
                            />
                            <FormField
                                name='approvedBy'
                                control={form.control}
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Approved By</FormLabel>
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <div className='flex gap-2'>
                                                <FormControl>
                                                    {bloodHoundUsersQuery.isLoading ? (
                                                        <Skeleton className='h-10 w-24' />
                                                    ) : (
                                                        <SelectTrigger>
                                                            <SelectValue placeholder='Select User' />
                                                        </SelectTrigger>
                                                    )}
                                                </FormControl>
                                                <SelectPortal>
                                                    <SelectContent>
                                                        {bloodHoundUsersQuery.data?.map((user) => {
                                                            if (!user.email_address) return null;
                                                            return (
                                                                <SelectItem
                                                                    key={user.id}
                                                                    value={user.email_address || user.id}>
                                                                    {user.email_address}
                                                                </SelectItem>
                                                            );
                                                        })}
                                                    </SelectContent>
                                                </SelectPortal>
                                                <Button
                                                    variant={'text'}
                                                    disabled={!field.value}
                                                    className={cn('w-1/12 p-0', { invisible: !field.value })}
                                                    onClick={() => {
                                                        form.setValue(field.name, '');
                                                    }}>
                                                    <FontAwesomeIcon icon={faClose} />
                                                </Button>
                                            </div>
                                        </Select>
                                    </FormItem>
                                )}
                            />
                            <div className='flex gap-6'>
                                <FormField
                                    name='start-date'
                                    control={form.control}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>Start Date</FormLabel>
                                            <div className='flex gap-2'>
                                                <DatePicker
                                                    {...field}
                                                    InputElement={MaskedInput}
                                                    calendarProps={{
                                                        mode: 'single',
                                                        fromDate,
                                                        toDate,
                                                        selected: DateTime.fromFormat(
                                                            field.value,
                                                            LuxonFormat.ISO_8601
                                                        ).toJSDate(),
                                                        onSelect: (date) => {
                                                            form.setValue(
                                                                field.name,
                                                                date
                                                                    ? DateTime.fromJSDate(date).toFormat(
                                                                          LuxonFormat.ISO_8601
                                                                      )
                                                                    : ''
                                                            );
                                                        },
                                                        disabled: (date) =>
                                                            DateTime.fromJSDate(date) > DateTime.local(),
                                                    }}
                                                />
                                                <Button
                                                    variant={'text'}
                                                    disabled={!field.value}
                                                    className={cn('w-1/12 p-0', { invisible: !field.value })}
                                                    onClick={() => {
                                                        form.setValue(field.name, '');
                                                        form.clearErrors();
                                                    }}>
                                                    <FontAwesomeIcon icon={faClose} />
                                                </Button>
                                            </div>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    name='end-date'
                                    control={form.control}
                                    render={({ field }) => (
                                        <FormItem>
                                            <FormLabel>End Date</FormLabel>
                                            <div className='flex gap-2'>
                                                <DatePicker
                                                    {...field}
                                                    InputElement={MaskedInput}
                                                    calendarProps={{
                                                        mode: 'single',
                                                        fromDate,
                                                        toDate,
                                                        selected: DateTime.fromFormat(
                                                            field.value,
                                                            LuxonFormat.ISO_8601
                                                        ).toJSDate(),
                                                        onSelect: (date) => {
                                                            form.setValue(
                                                                field.name,
                                                                date
                                                                    ? DateTime.fromJSDate(date).toFormat(
                                                                          LuxonFormat.ISO_8601
                                                                      )
                                                                    : ''
                                                            );
                                                        },
                                                        disabled: (date) =>
                                                            DateTime.fromJSDate(date) > DateTime.local(),
                                                    }}
                                                />
                                                <Button
                                                    variant={'text'}
                                                    disabled={!field.value}
                                                    className={cn('w-1/12 p-0', { invisible: !field.value })}
                                                    onClick={() => {
                                                        form.setValue(field.name, '');
                                                        form.clearErrors();
                                                    }}>
                                                    <FontAwesomeIcon icon={faClose} />
                                                </Button>
                                            </div>{' '}
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            <DialogActions>
                                <Button variant='text' onClick={closeDialog}>
                                    Cancel
                                </Button>
                                <Button variant='text' onClick={handleConfirm}>
                                    Confirm
                                </Button>
                            </DialogActions>
                        </form>
                    </Form>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default FilterDialog;
