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
import { AssetGroupTagTypeLabel, AssetGroupTagTypeOwned, AssetGroupTagTypeTier } from 'js-client-library';
import { FC, useContext, useState } from 'react';
import { UseQueryResult } from 'react-query';
import { Link, useLocation, useParams } from 'react-router-dom';
import { SortOrder } from '../../../types';
import { useAppNavigate } from '../../../utils';
import { ZoneManagementContext } from '../ZoneManagementContext';
import {
    useSelectorMembersInfiniteQuery,
    useSelectorsInfiniteQuery,
    useTagMembersInfiniteQuery,
    useTagsQuery,
} from '../hooks';
import { TIER_ZERO_ID, getTagUrlValue } from '../utils';
import { DetailsList } from './DetailsList';
import { MembersList } from './MembersList';
import { SelectedDetails } from './SelectedDetails';
import { SelectorsList } from './SelectorsList';

export const getSavePath = (
    tierId: string | undefined,
    labelId: string | undefined,
    selectorId: string | undefined
) => {
    const savePath = '/zone-management/save';

    if (selectorId && labelId) return `/zone-management/save/label/${labelId}/selector/${selectorId}`;
    if (selectorId && tierId) return `/zone-management/save/tier/${tierId}/selector/${selectorId}`;

    if (!selectorId && labelId) return `/zone-management/save/label/${labelId}`;
    if (!selectorId && tierId) return `/zone-management/save/tier/${tierId}`;

    return savePath;
};

export const getEditButtonState = (
    memberId?: string,
    selectorsQuery?: UseQueryResult,
    tiersQuery?: UseQueryResult,
    labelsQuery?: UseQueryResult
) => {
    return (
        !!memberId ||
        (selectorsQuery?.isLoading && tiersQuery?.isLoading && labelsQuery?.isLoading) ||
        (selectorsQuery?.isError && tiersQuery?.isError && labelsQuery?.isError)
    );
};

const Details: FC = () => {
    const navigate = useAppNavigate();
    const location = useLocation();
    const { tierId = TIER_ZERO_ID, labelId, selectorId, memberId } = useParams();
    const [membersListSortOrder, setMembersListSortOrder] = useState<SortOrder>('asc');
    const tagId = labelId === undefined ? tierId : labelId;

    const context = useContext(ZoneManagementContext);
    if (!context) {
        throw new Error('Details must be used within a ZoneManagementContext.Provider');
    }
    const { InfoHeader } = context;

    const tiersQuery = useTagsQuery((tag) => tag.type === AssetGroupTagTypeTier);

    const labelsQuery = useTagsQuery(
        (tag) => tag.type === AssetGroupTagTypeLabel || tag.type === AssetGroupTagTypeOwned
    );

    const selectorsQuery = useSelectorsInfiniteQuery(tagId);

    const selectorMembersQuery = useSelectorMembersInfiniteQuery(tagId, selectorId, membersListSortOrder);

    const tagMembersQuery = useTagMembersInfiniteQuery(tagId, membersListSortOrder);

    const showEditButton = !getEditButtonState(memberId, selectorsQuery, tiersQuery, labelsQuery);

    return (
        <div>
            <div className='flex mt-6 gap-8'>
                {InfoHeader && <InfoHeader />}
                <div className='basis-1/3'>
                    {showEditButton && (
                        <Button asChild variant={'secondary'} disabled={showEditButton}>
                            <Link to={getSavePath(tierId, labelId, selectorId)}>Edit</Link>
                        </Button>
                    )}
                </div>
            </div>
            <div className='flex gap-8 mt-4'>
                <div className='flex basis-2/3 bg-neutral-light-2 dark:bg-neutral-dark-2 rounded-lg shadow-outer-1 *:w-1/3 h-full'>
                    {location.pathname.includes('label') ? (
                        <DetailsList
                            title={'Labels'}
                            listQuery={labelsQuery}
                            selected={tagId}
                            onSelect={(id) => {
                                navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${id}`);
                            }}
                        />
                    ) : (
                        <DetailsList
                            title={'Tiers'}
                            listQuery={tiersQuery}
                            selected={tagId}
                            onSelect={(id) => {
                                navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${id}`);
                            }}
                        />
                    )}

                    <SelectorsList
                        listQuery={selectorsQuery}
                        selected={selectorId}
                        onSelect={(id) => {
                            navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/selector/${id}`);
                        }}
                    />
                    {selectorId !== undefined ? (
                        <MembersList
                            listQuery={selectorMembersQuery}
                            selected={memberId}
                            onClick={(id) => {
                                navigate(
                                    `/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/selector/${selectorId}/member/${id}`
                                );
                            }}
                            sortOrder={membersListSortOrder}
                            onChangeSortOrder={setMembersListSortOrder}
                        />
                    ) : (
                        <MembersList
                            listQuery={tagMembersQuery}
                            selected={memberId}
                            onClick={(id) => {
                                navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/member/${id}`);
                            }}
                            sortOrder={membersListSortOrder}
                            onChangeSortOrder={setMembersListSortOrder}
                        />
                    )}
                </div>
                <div className='basis-1/3'>
                    <SelectedDetails />
                </div>
            </div>
        </div>
    );
};

export default Details;
