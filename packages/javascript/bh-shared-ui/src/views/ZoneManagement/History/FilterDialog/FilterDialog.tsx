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
    DialogPortal,
    DialogTitle,
    Form,
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
    Skeleton,
    VisuallyHidden,
} from '@bloodhoundenterprise/doodleui';
import { DateTime } from 'luxon';
import { FC, HTMLAttributes, useCallback, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { DatePicker } from '../../../../components';
import { useTagsQuery } from '../../../../hooks';
import { useBloodHoundUsers } from '../../../../hooks/useBloodHoundUsers';
import { LuxonFormat, cn } from '../../../../utils';

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

const toDate = DateTime.local();
const fromDate = toDate.minus({ years: 1 });

const FilterDialog: FC<{
    open: boolean;
    handleClose: () => void;
    setFilters: (filters: AssetGroupTagHistoryFilters) => void;
    filters?: AssetGroupTagHistoryFilters;
}> = ({ open, filters = defaultValues, handleClose, setFilters }) => {
    const tagsQuery = useTagsQuery();
    const bloodHoundUsersQuery = useBloodHoundUsers();

    const form = useForm<AssetGroupTagHistoryFilters>({ defaultValues });

    const selectClasses: (field: keyof AssetGroupTagHistoryFilters) => HTMLAttributes<HTMLSelectElement>['className'] =
        useCallback(
            (fieldName: keyof AssetGroupTagHistoryFilters) => {
                return cn(
                    'rounded-md border border-neutral-dark-5 dark:border-neutral-light-5 px-3 py-2 text-sm ring-offset-secondary dark:ring-offset-secondary-variant-2 focus-visible:border-none focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-secondary dark:focus-visible:ring-secondary-variant-2 focus-visible:ring-offset-2 hover:border-2 bg-neutral-1',
                    { 'w-24': form.getValues(fieldName) === '' }
                );
            },
            [form]
        );

    useEffect(() => {
        form.reset(filters);
    }, [form, filters]);

    return (
        <Dialog open={open}>
            <DialogPortal>
                <DialogContent className='flex flex-col gap-4'>
                    <Form {...form}>
                        <form className='flex flex-col gap-4'>
                            <DialogTitle className='flex justify-between items-center'>
                                <span>Filter</span>
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
                                                <SelectTrigger className={selectClasses(field.name)}>
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
                                        <FormLabel aria-labelledby='tag'>Tier/Label</FormLabel>
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
                                                    <SelectTrigger className={selectClasses(field.name)}>
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
                                                    <SelectTrigger className={selectClasses(field.name)}>
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
                            <FormField
                                name='start-date'
                                control={form.control}
                                render={({ field }) => (
                                    <FormItem className='w-40'>
                                        <FormLabel htmlFor={field.name}>Start Date</FormLabel>
                                        <FormControl>
                                            <DatePicker
                                                {...field}
                                                error={form.formState.errors['start-date']?.message}
                                                clearError={() => form.clearErrors(field.name)}
                                                setValue={(date: string) =>
                                                    form.setValue('start-date', date, { shouldDirty: true })
                                                }
                                                fromDate={fromDate}
                                                toDate={toDate}
                                                disabled={(date: Date) => {
                                                    return DateTime.fromJSDate(date) > DateTime.local();
                                                }}
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />
                            <FormField
                                name='end-date'
                                control={form.control}
                                render={({ field }) => (
                                    <FormItem className='w-40'>
                                        <FormLabel htmlFor={field.name}>End Date</FormLabel>
                                        <FormControl>
                                            <DatePicker
                                                {...field}
                                                error={form.formState.errors['end-date']?.message}
                                                clearError={() => form.clearErrors(field.name)}
                                                setValue={(date: string) =>
                                                    form.setValue('end-date', date, { shouldDirty: true })
                                                }
                                                fromDate={fromDate}
                                                toDate={toDate}
                                                disabled={
                                                    form.getValues('start-date')
                                                        ? (date: Date) =>
                                                              DateTime.fromFormat(
                                                                  form.getValues('start-date'),
                                                                  LuxonFormat.ISO_8601
                                                              ) > DateTime.fromJSDate(date)
                                                        : undefined
                                                }
                                            />
                                        </FormControl>
                                    </FormItem>
                                )}
                            />
                            <DialogActions>
                                <Button variant={'text'} onClick={handleClose} className='p-2'>
                                    Cancel
                                </Button>
                                <Button
                                    variant={'text'}
                                    className='text-primary dark:text-secondary-variant-2 p-2'
                                    onClick={() => {
                                        setFilters({ ...form.getValues() });
                                    }}>
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
