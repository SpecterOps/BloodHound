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

import { Button } from '@bloodhoundenterprise/doodleui';
import { FC, useContext } from 'react';
import { UseQueryResult, useQuery } from 'react-query';
import { Link, useParams } from 'react-router-dom';
import { useHighestPrivilegeTagId } from '../../../hooks';
import { ROUTE_ZONE_MANAGEMENT_SUMMARY } from '../../../routes';
import { apiClient, useAppNavigate } from '../../../utils';
import { getSavePath } from '../Details/Details';
import { SelectedDetails } from '../Details/SelectedDetails';
import { ZoneManagementContext } from '../ZoneManagementContext';
import { getTagUrlValue } from '../utils';
import SummaryList from './SummaryList';

export const getEditButtonState = (memberId?: string, selectorsQuery?: UseQueryResult, tagsQuery?: UseQueryResult) => {
    return (
        !!memberId ||
        (selectorsQuery?.isLoading && tagsQuery?.isLoading) ||
        (selectorsQuery?.isError && tagsQuery?.isError)
    );
};

const Summary: FC = () => {
    const navigate = useAppNavigate();
    const topTagId = useHighestPrivilegeTagId()?.toString();
    const { tierId = topTagId, labelId, selectorId, memberId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;

    const context = useContext(ZoneManagementContext);
    if (!context) {
        throw new Error('Details must be used within a ZoneManagementContext.Provider');
    }
    const { InfoHeader } = context;

    const tagsQuery = useQuery({
        queryKey: ['zone-management', 'tags'],
        queryFn: async () => {
            return apiClient.getAssetGroupTags({ params: { counts: true } }).then((res) => {
                return res.data.data['tags'];
            });
        },
    });

    const selectorsQuery = useQuery({
        queryKey: ['zone-management', 'tags', tagId, 'selectors'],
        queryFn: async () => {
            if (!tagId) return [];
            return apiClient.getAssetGroupTagSelectors(tagId, { params: { counts: true } }).then((res) => {
                return res.data.data['selectors'];
            });
        },
    });

    const showEditButton = !getEditButtonState(memberId, selectorsQuery, tagsQuery);

    return (
        <div>
            <div className='flex mt-6 gap-8'>
                <InfoHeader />
                <div className='basis-1/3'>
                    {showEditButton && (
                        <Button asChild variant={'secondary'} disabled={showEditButton}>
                            <Link
                                data-testid='zone-management_edit-button'
                                to={getSavePath(tierId, labelId, selectorId)}>
                                Edit
                            </Link>
                        </Button>
                    )}
                </div>
            </div>
            <div className='flex gap-8 mt-6 w-full'>
                <div className='flex-1'>
                    <SummaryList
                        title={labelId ? 'Labels' : 'Tiers'}
                        listQuery={tagsQuery}
                        selected={tagId as string}
                        onSelect={(id) => {
                            navigate(
                                `/zone-management/${ROUTE_ZONE_MANAGEMENT_SUMMARY}/${getTagUrlValue(labelId)}/${id}`
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
