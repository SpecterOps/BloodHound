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
} from 'doodle-ui';
import { DateTime } from 'luxon';
import { FC, useCallback, useEffect } from 'react';
import { ErrorOption, useForm } from 'react-hook-form';
import { MaskedInput } from '../../../../components';
import { useTagsQuery } from '../../../../hooks';
import { useBloodHoundUsers } from '../../../../hooks/useBloodHoundUsers';
import { CustomRangeError, END_DATE, LuxonFormat, START_DATE } from '../../../../utils';

const actionOptions = [
    '', // Empty string added to list for adhering to `(typeof actionOptions)[number]` type
    'Certified by user',
    'Certified',
    'Added to Selector',
    'Modified',
    'Created',
    'Deleted',
] as const;

interface AssetGroupTagHistoryFilters {
    action: (typeof actionOptions)[number];
    tag: string;
    madeBy: string;
    ['start-date']: string;
    ['end-date']: string;
}

const defaultValues = { action: actionOptions[0], tag: '', madeBy: '', 'start-date': '', 'end-date': '' };

const toDate = DateTime.local().toJSDate();
const fromDate = DateTime.fromJSDate(toDate).minus({ years: 1 }).toJSDate();

const FilterDialog: FC<{
    open: boolean;
    handleClose: () => void;
    setFilters: (filters: AssetGroupTagHistoryFilters) => void;
    filters?: AssetGroupTagHistoryFilters;
}> = ({ open, filters = defaultValues, handleClose, setFilters }) => {
    const tagsQuery = useTagsQuery();
    const bloodHoundUsersQuery = useBloodHoundUsers();

    const form = useForm<AssetGroupTagHistoryFilters>({ defaultValues });

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

        // Allow partial filtering of records; Do not block if neither date is filled
        if (!start && !end) {
            setFilters({ ...form.getValues() });
            return;
        }

        const startDate = DateTime.fromFormat(start, LuxonFormat.ISO_8601);
        const endDate = DateTime.fromFormat(end, LuxonFormat.ISO_8601);

        // Otherwise validate that both dates are chosen for a valid range
        if (validateDateFields(startDate, endDate)) {
            setFilters({ ...form.getValues() });
        }
    }, [form, setFilters, validateDateFields]);

    useEffect(() => {
        form.reset(filters);
    }, [form, filters]);

    return (
        <Dialog open={open}>
            <DialogPortal>
                <DialogContent>
                    <Form {...form}>
                        <form className='flex flex-col gap-4'>
                            <DialogTitle className='flex justify-between items-center'>
                                <span className='text-xl'>Filter</span>
                                <Button
                                    variant={'text'}
                                    onClick={() => form.reset(defaultValues)}
                                    className='font-normal p-2'>
                                    Clear All
                                </Button>
                            </DialogTitle>
                            <VisuallyHidden asChild>
                                <DialogDescription>Filter Privilege Zone History</DialogDescription>
                            </VisuallyHidden>
                            <FormField
                                control={form.control}
                                name='action'
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel aria-labelledby='action'>Action</FormLabel>
                                        <Select
                                            onValueChange={field.onChange}
                                            value={field.value}
                                            defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder='Select' {...field} />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectPortal>
                                                <SelectContent>
                                                    {actionOptions.map((action, index) => {
                                                        if (index === 0) return; // Do not render empty string item
                                                        return (
                                                            <SelectItem key={action} value={action}>
                                                                {action}
                                                            </SelectItem>
                                                        );
                                                    })}
                                                </SelectContent>
                                            </SelectPortal>
                                        </Select>
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name='tag'
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel aria-labelledby='tag'>Zone/Label</FormLabel>
                                        <Select
                                            onValueChange={field.onChange}
                                            value={field.value}
                                            defaultValue={field.value}>
                                            <FormControl>
                                                {tagsQuery.isError ? (
                                                    <span className='text-error'>
                                                        There was an error fetching this data. Please refresh the page
                                                        to try again.
                                                    </span>
                                                ) : (
                                                    <SelectTrigger>
                                                        <SelectValue placeholder='Select' />
                                                    </SelectTrigger>
                                                )}
                                            </FormControl>
                                            <SelectPortal>
                                                {tagsQuery.isLoading ? (
                                                    <Skeleton className='h-10 w-24' />
                                                ) : (
                                                    <SelectContent>
                                                        {tagsQuery.data?.map((tag) => (
                                                            <SelectItem key={tag.id} value={tag.name}>
                                                                {tag.name}
                                                            </SelectItem>
                                                        ))}
                                                    </SelectContent>
                                                )}
                                            </SelectPortal>
                                        </Select>
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name='madeBy'
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel aria-labelledby='madeBy'>Made By</FormLabel>
                                        <Select
                                            onValueChange={field.onChange}
                                            value={field.value}
                                            defaultValue={field.value}>
                                            <FormControl>
                                                {bloodHoundUsersQuery.isError ? (
                                                    <span className='text-error'>
                                                        There was an error fetching this data. Please refresh the page
                                                        to try again.
                                                    </span>
                                                ) : (
                                                    <SelectTrigger>
                                                        <SelectValue placeholder='Select' />
                                                    </SelectTrigger>
                                                )}
                                            </FormControl>
                                            <SelectPortal>
                                                {bloodHoundUsersQuery.isLoading ? (
                                                    <Skeleton className='h-10 w-24' />
                                                ) : (
                                                    <SelectContent>
                                                        <SelectItem value={'SYSTEM'}>SYSTEM</SelectItem>
                                                        {bloodHoundUsersQuery.data
                                                            ?.filter((user) => {
                                                                let hasPermission = false;
                                                                user.roles.forEach((role) => {
                                                                    role.permissions.forEach((permission) => {
                                                                        if (
                                                                            permission.name === 'ManageUsers' &&
                                                                            permission.authority === 'auth'
                                                                        )
                                                                            hasPermission = true;
                                                                    });
                                                                });

                                                                return hasPermission;
                                                            })
                                                            .sort((a, b) =>
                                                                (a.email_address || a.principal_name).localeCompare(
                                                                    b.email_address || b.principal_name
                                                                )
                                                            )
                                                            ?.map((user) => (
                                                                <SelectItem
                                                                    key={user.id}
                                                                    value={user.email_address || user.id}>
                                                                    {user.email_address || user.principal_name}
                                                                </SelectItem>
                                                            ))}
                                                    </SelectContent>
                                                )}
                                            </SelectPortal>
                                        </Select>
                                    </FormItem>
                                )}
                            />
                            <div className='flex gap-6'>
                                <FormField
                                    name='start-date'
                                    control={form.control}
                                    render={({ field }) => (
                                        <FormItem className='w-40 flex flex-col gap-2 justify-start'>
                                            <FormLabel htmlFor={field.name}>Start Date</FormLabel>
                                            <FormControl>
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
                                                        onSelect: (value: Date | undefined) => {
                                                            form.setValue(
                                                                field.name,
                                                                value
                                                                    ? DateTime.fromJSDate(value).toFormat(
                                                                          LuxonFormat.ISO_8601
                                                                      )
                                                                    : ''
                                                            );
                                                        },
                                                        disabled: (date: Date) => {
                                                            return DateTime.fromJSDate(date) > DateTime.local();
                                                        },
                                                    }}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                                <FormField
                                    name='end-date'
                                    control={form.control}
                                    render={({ field }) => (
                                        <FormItem className='w-40 flex flex-col gap-2 justify-start'>
                                            <FormLabel htmlFor={field.name}>End Date</FormLabel>
                                            <FormControl>
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
                                                        onSelect: (value: Date | undefined) => {
                                                            form.setValue(
                                                                field.name,
                                                                value
                                                                    ? DateTime.fromJSDate(value).toFormat(
                                                                          LuxonFormat.ISO_8601
                                                                      )
                                                                    : ''
                                                            );
                                                        },
                                                        disabled: (date: Date) => {
                                                            return DateTime.fromJSDate(date) > DateTime.local();
                                                        },
                                                    }}
                                                />
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            <DialogActions>
                                <Button variant={'text'} onClick={handleClose} className='p-2'>
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
            </DialogPortal>
        </Dialog>
    );
};

export default FilterDialog;
