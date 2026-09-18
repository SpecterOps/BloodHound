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
import { faClose } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
    Skeleton,
    TextButton,
} from 'doodle-ui';
import { FC } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { TagSelect, useTagsQuery } from '../../../hooks';
import { cn } from '../../../utils';

export const TagIdField: FC<{
    fieldLabel: string;
    form: UseFormReturn;
    tagSelect?: TagSelect;
}> = ({ form, tagSelect, fieldLabel }) => {
    const tagsQuery = useTagsQuery({ select: tagSelect });

    return (
        <FormField
            control={form.control}
            name='tagId'
            render={({ field }) => {
                const selectedTagName = tagsQuery.data?.find((tag) => tag.id.toString() === field.value)?.name;

                return (
                    <FormItem>
                        <FormLabel aria-labelledby='tag'>{fieldLabel}</FormLabel>
                        <Select onValueChange={field.onChange} value={field.value} defaultValue={field.value}>
                            <div className='flex min-w-0 gap-2'>
                                <FormControl className='w-11/12 min-w-0'>
                                    {tagsQuery.isError ? (
                                        <span className='text-error'>
                                            There was an error fetching this data. Please refresh the page to try again.
                                        </span>
                                    ) : (
                                        <SelectTrigger
                                            className='min-w-0 overflow-hidden gap-2 [&>span]:min-w-0 [&>span]:truncate [&>svg]:shrink-0'
                                            title={selectedTagName}>
                                            <SelectValue placeholder={`Select ${fieldLabel}`} />
                                        </SelectTrigger>
                                    )}
                                </FormControl>
                                <TextButton
                                    aria-label={`clear ${fieldLabel}`}
                                    disabled={!field.value}
                                    className={cn('w-1/12 p-0', { invisible: !field.value })}
                                    onClick={() => {
                                        form.setValue(field.name, '');
                                    }}>
                                    <FontAwesomeIcon icon={faClose} />
                                </TextButton>
                            </div>

                            {tagsQuery.isLoading ? (
                                <Skeleton className='h-10 w-24' />
                            ) : (
                                <SelectContent className='max-w-[var(--radix-select-trigger-width)]'>
                                    {tagsQuery.data?.map((tag) => (
                                        <SelectItem
                                            key={tag.id}
                                            value={tag.id.toString()}
                                            title={tag.name}
                                            className='min-w-0 [&>span]:truncate'>
                                            {tag.name}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            )}
                        </Select>
                    </FormItem>
                );
            }}
        />
    );
};
