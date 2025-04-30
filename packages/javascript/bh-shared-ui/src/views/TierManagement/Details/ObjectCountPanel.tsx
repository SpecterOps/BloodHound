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

import { Badge, Card } from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';
import { useQuery } from 'react-query';
import { apiClient } from '../../../utils';
import { itemSkeletons } from './utils';

const ObjectCountPanel: FC<{ tagId: string }> = ({ tagId }) => {
    const objectsCountQuery = useQuery(['asset-group-tags-count'], () => {
        return apiClient.getAssetGroupTagMembersCount(tagId).then((res) => {
            return res.data.data;
        });
    });

    if (objectsCountQuery.isLoading) {
        return itemSkeletons.map((skeleton, index) => {
            return skeleton('object-count', index);
        });
    }

    if (objectsCountQuery.isError) {
        return (
            <li className='border-neutral-light-3 dark:border-neutral-dark-3'>
                <span className='text-base'>There was an error fetching this data</span>
            </li>
        );
    }

    if (objectsCountQuery.isSuccess) {
        const { total_count, counts } = objectsCountQuery.data;

        return (
            <Card className='flex flex-col max-h-full px-6 py-6 select-none overflow-y-auto max-w-[32rem]'>
                <div className='flex justify-between'>
                    <p>Total Count</p>
                    <Badge label={total_count.toLocaleString()} />
                </div>
                {Object.entries(counts).map(([key, value]) => {
                    return (
                        <div className='flex justify-between mt-4' key={key}>
                            <p>{key}</p>
                            <Badge label={value.toLocaleString()} />
                        </div>
                    );
                })}
            </Card>
        );
    }
};

export default ObjectCountPanel;
