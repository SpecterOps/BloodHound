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

type Props = {
    /** Status options available in the select menu. Defaults to all if not provided. */
    statusOptions?: string[];

    /** Currently selected status. */
    status?: number;

    /** Callback for when a status is selected. */
    onSelect: (value: string) => void;
};

export const StatusSelect: FC<Props> = ({ status = '', statusOptions, onSelect }) => {
    const STATUS_FILTERS = Object.entries(JOB_STATUS_MAP).filter(([, value]) =>
        statusOptions ? statusOptions.includes(value) : true
    );

    return (
        <div className='flex flex-col gap-2'>
            <Label>Status</Label>

            <Select onValueChange={onSelect} value={status.toString()}>
                <SelectTrigger className='w-32' aria-label='Status Select'>
                    <SelectValue placeholder='Select' />
                </SelectTrigger>
                <SelectPortal>
                    <SelectContent>
                        <SelectItem className='italic' key='status-unselect' value='-none-'>
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
