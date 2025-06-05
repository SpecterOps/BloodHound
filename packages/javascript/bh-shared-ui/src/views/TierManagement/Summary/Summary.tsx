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

import { FC } from 'react';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { ROUTE_TIER_MANAGEMENT_SUMMARY } from '../../../routes';
import { apiClient, useAppNavigate } from '../../../utils';
import { SelectedDetails } from '../Details/SelectedDetails';
import { TierActionBar } from '../fragments';
import { TIER_ZERO_ID, getTagUrlValue } from '../utils';
import SummaryList from './SummaryList';

const Summary: FC = () => {
    const navigate = useAppNavigate();
    const { tierId = TIER_ZERO_ID, labelId, selectorId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;

    const tagsQuery = useQuery({
        queryKey: ['tier-management', 'tags'],
        queryFn: async () => {
            return apiClient.getAssetGroupTags({ params: { counts: true } }).then((res) => {
                return res.data.data['tags'];
            });
        },
    });

    return (
        <div>
            <TierActionBar tierId={tagId} labelId={labelId} selectorId={selectorId} />
            <div className='flex gap-8 mt-4 w-full'>
                <div className='flex-1'>
                    <SummaryList
                        title={labelId ? 'Labels' : 'Tiers'}
                        listQuery={tagsQuery}
                        selected={tagId}
                        onSelect={(id) => {
                            navigate(
                                `/tier-management/${ROUTE_TIER_MANAGEMENT_SUMMARY}/${getTagUrlValue(labelId)}/${id}`
                            );
                        }}
                    />
                </div>
                <div className='basis-1/3'>
                    <SelectedDetails />
                </div>
            </div>
        </div>
    );
};

export default Summary;
