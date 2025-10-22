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
    DialogClose,
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
import { AssetGroupTagCertificationRecord, ExtendedCertificationFilters } from 'js-client-library';
import { DateTime } from 'luxon';
import { FC, useCallback, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { AppIcon, MaskedInput } from '../../../components';
import { useBloodHoundUsers } from '../../../hooks/useBloodHoundUsers';
import { CustomRangeError, END_DATE, LuxonFormat, START_DATE } from '../../../utils';

type FilterFormValues = Partial<ExtendedCertificationFilters> & {
    'start-date'?: string;
    'end-date'?: string;
};

interface FilterDialogProps {
    filters: Partial<ExtendedCertificationFilters>;
    setFilters: (filters: ExtendedCertificationFilters) => void;
    onApplyFilters: (filters: Partial<ExtendedCertificationFilters>) => void;
    data: AssetGroupTagCertificationRecord[];
}

const defaultValues: FilterFormValues = { objectType: '', approvedBy: '', 'start-date': '', 'end-date': '' };

const toDate = DateTime.local().toJSDate();
const fromDate = DateTime.fromJSDate(toDate).minus({ years: 1 }).toJSDate();

//TODO: we need to consolidate this into one universal shared component but separating in the interest of time
const FilterDialog: FC<FilterDialogProps> = ({ filters, onApplyFilters, data }) => {
    const bloodHoundUsersQuery = useBloodHoundUsers();

    const form = useForm<FilterFormValues>({
        defaultValues: filters,
    });

    const validateDates = useCallback(
        (startDate?: DateTime, endDate?: DateTime) => {
            form.clearErrors();
            const errors: { name: typeof START_DATE | typeof END_DATE; error: any }[] = [];

            if (startDate && !startDate.isValid)
                errors.push({ name: START_DATE, error: { message: CustomRangeError.INVALID_DATE } });
            if (endDate && !endDate.isValid)
                errors.push({ name: END_DATE, error: { message: CustomRangeError.INVALID_DATE } });

            if (startDate && endDate && startDate > endDate) {
                errors.push({ name: START_DATE, error: { message: CustomRangeError.INVALID_RANGE_START } });
                errors.push({ name: END_DATE, error: { message: CustomRangeError.INVALID_RANGE_END } });
            }

            errors.forEach((e) => form.setError(e.name, e.error));

            return errors.length === 0;
        },
        [form]
    );

    const handleConfirm = useCallback(() => {
        const values = form.getValues();
        if (
            validateDates(
                values.startDate ? DateTime.fromFormat(values.startDate, LuxonFormat.ISO_8601) : undefined,
                values.endDate ? DateTime.fromFormat(values.endDate, LuxonFormat.ISO_8601) : undefined
            )
        ) {
            onApplyFilters(values); // parent table only updated on Confirm
        }
    }, [form, onApplyFilters, validateDates]);

    //filter out duplicate member types
    const uniqueObjects = data.filter(
        (item, index, self) => index === self.findIndex((t) => t.primary_kind === item.primary_kind)
    );

    useEffect(() => {
        form.reset(filters);
    }, [form, filters]);
    return (
        <Dialog>
            <DialogTrigger asChild>
                <Button variant='text' data-testid='certifications_filter_dialog'>
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
                                        form.reset(defaultValues);
                                    }}
                                    className='font-normal p-2'>
                                    Clear All
                                </Button>
                            </DialogTitle>
                            <VisuallyHidden asChild>
                                <DialogDescription>Filter Privilege Zone Certifications</DialogDescription>
                            </VisuallyHidden>
                            <FormField
                                name='objectType'
                                control={form.control}
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Object Type</FormLabel>
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder='Select Object Type' />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectPortal>
                                                <SelectContent>
                                                    {uniqueObjects.map((obj) => (
                                                        <SelectItem key={obj.primary_kind} value={obj.primary_kind}>
                                                            {obj.primary_kind}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
                                            </SelectPortal>
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
                                                    {bloodHoundUsersQuery.data?.map((user) => (
                                                        <SelectItem key={user.id} value={user.email_address || user.id}>
                                                            {user.email_address || user.principal_name}
                                                        </SelectItem>
                                                    ))}
                                                </SelectContent>
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
                                        <FormItem>
                                            <FormLabel>Start Date</FormLabel>
                                            <FormControl>
                                                <DatePicker
                                                    {...field}
                                                    InputElement={MaskedInput}
                                                    calendarProps={{
                                                        mode: 'single',
                                                        fromDate,
                                                        toDate,
                                                        selected: field.value
                                                            ? DateTime.fromFormat(
                                                                  field.value,
                                                                  LuxonFormat.ISO_8601
                                                              ).toJSDate()
                                                            : undefined,
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
                                            </FormControl>
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
                                            <FormControl>
                                                <DatePicker
                                                    {...field}
                                                    InputElement={MaskedInput}
                                                    calendarProps={{
                                                        mode: 'single',
                                                        fromDate,
                                                        toDate,
                                                        selected: field.value
                                                            ? DateTime.fromFormat(
                                                                  field.value,
                                                                  LuxonFormat.ISO_8601
                                                              ).toJSDate()
                                                            : undefined,
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
                                            </FormControl>
                                            <FormMessage />
                                        </FormItem>
                                    )}
                                />
                            </div>

                            <DialogActions>
                                <DialogClose asChild>
                                    <Button variant='text'>Cancel</Button>
                                </DialogClose>
                                <DialogClose asChild>
                                    <Button variant='text' onClick={handleConfirm}>
                                        Confirm
                                    </Button>
                                </DialogClose>
                            </DialogActions>
                        </form>
                    </Form>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default FilterDialog;
