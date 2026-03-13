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
    Button,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from 'doodle-ui';
import { FC } from 'react';
import { UseFormReturn } from 'react-hook-form';
import { cn } from '../../../../utils';
import { AssetGroupTagHistoryFilters } from '../types';
import { actionMap } from '../utils';

const ActionField: FC<{
    form: UseFormReturn<AssetGroupTagHistoryFilters>;
}> = ({ form }) => {
    return (
        <FormField
            control={form.control}
            name='action'
            render={({ field }) => (
                <FormItem>
                    <FormLabel aria-labelledby='action'>Action</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value} defaultValue={field.value}>
                        <div className='flex gap-2'>
                            <FormControl className='w-11/12'>
                                <SelectTrigger>
                                    <SelectValue placeholder='Select' {...field} />
                                </SelectTrigger>
                            </FormControl>
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
                        <SelectContent>
                            {actionMap.map((action, index) => {
                                if (index === 0) return; // Do not render empty string item
                                return (
                                    <SelectItem key={actionMap[index].value} value={actionMap[index].value}>
                                        {actionMap[index].label}
                                    </SelectItem>
                                );
                            })}
                        </SelectContent>
                    </Select>
                </FormItem>
            )}
        />
    );
};

export default ActionField;
