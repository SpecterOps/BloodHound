// Copyright 2024 Specter Ops, Inc.
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

import { Skeleton } from '@bloodhoundenterprise/doodleui';
import { Checkbox, FormControlLabel } from '@mui/material';
import { type OptionsObject } from 'notistack';
import { type FC } from 'react';
import { useQuery } from 'react-query';
import { useNotifications } from '../../providers/NotificationProvider/hooks';
import { apiClient } from '../../utils/api';

const ERROR = {
    key: 'database-management-source-kind',
    message: 'An error occurred while loading source kinds. Deleting graph data is diabled. Try refreshing the page.',
    options: {
        persist: true,
        anchorOrigin: { vertical: 'top', horizontal: 'right' },
    } as OptionsObject,
};

const useSourceKindsQuery = () => {
    const { addNotification } = useNotifications();

    return useQuery({
        queryKey: ['source-kinds'],
        queryFn: ({ signal }) => apiClient.getSourceKinds({ signal }).then((res) => res.data.data.kinds),
        onError: () => addNotification(ERROR.message, ERROR.key, ERROR.options),
    });
};

// Displayed while source kinds are loading
const LOADING_CHECKBOXES = (
    <>
        <div role='status' className='pl-5 flex items-center'>
            <Checkbox disabled />
            <Skeleton className='h-4 w-[200px]' />
        </div>
        <div role='status' className='pl-5 flex items-center'>
            <Checkbox disabled />
            <Skeleton className='h-4 w-[200px]' />
        </div>
        <div role='status' className='pl-5 flex items-center'>
            <Checkbox disabled />
            <Skeleton className='h-4 w-[200px]' />
        </div>
    </>
);

// The default source kind names are replaced with friendlier ones
const KIND_LABEL_MAP: Record<string, string> = {
    Base: 'Active Directory',
    AZBase: 'Azure',
};

export const SourceKindsCheckboxes: FC<{
    checked: number[];
    disabled?: boolean;
    onChange: (checked: number[]) => void;
}> = ({ checked, disabled = true, onChange }) => {
    const { data: sourceKinds, isLoading, isSuccess } = useSourceKindsQuery();

    // Feature disabled is passed in prop or if query fails
    const isDisabled = disabled || !isSuccess;
    let amountChecked = 'none';

    if (isSuccess && checked.length > 0) {
        amountChecked = sourceKinds.length === checked.length ? 'all' : 'some';
    }

    // If all boxes are checked, they are all unchecked; other wise all boxes are checked
    const toggleAllChecked = () => {
        if (sourceKinds) {
            onChange(['none', 'some'].includes(amountChecked) ? sourceKinds.map((item) => item.id) : []);
        }
    };

    // Toggle a source kind on or off, then update the set of checked boxes
    const toggleSourceKind = (id: number) => () => {
        const newChecked = checked.includes(id) ? checked.filter((item) => item !== id) : [...checked, id];
        onChange(newChecked);
    };

    return (
        <div className='flex flex-col' data-testid='source-kinds-checkboxes'>
            <FormControlLabel
                label='All graph data'
                control={
                    <Checkbox
                        checked={amountChecked === 'all'}
                        disabled={isDisabled}
                        indeterminate={amountChecked === 'some'}
                        name='All GraphData'
                        onChange={toggleAllChecked}
                    />
                }
            />

            {isLoading && LOADING_CHECKBOXES}

            {isSuccess &&
                sourceKinds.map((item) => (
                    <FormControlLabel
                        className='pl-8'
                        control={
                            <Checkbox
                                checked={checked.includes(item.id)}
                                onChange={toggleSourceKind(item.id)}
                                name={item.name}
                                disabled={isDisabled}
                            />
                        }
                        label={(KIND_LABEL_MAP[item.name] ?? item.name) + ' data'}
                        key={item.id}
                    />
                ))}
        </div>
    );
};
