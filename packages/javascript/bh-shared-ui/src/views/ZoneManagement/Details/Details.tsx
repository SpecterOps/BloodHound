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
import { AssetGroupTagMemberListItem, AssetGroupTagSelector } from 'js-client-library';
import { FC, useContext, useState } from 'react';
import { UseQueryResult, useInfiniteQuery, useQuery } from 'react-query';
import { Link, useLocation, useParams } from 'react-router-dom';
import { SortOrder } from '../../../types';
import { apiClient, useAppNavigate } from '../../../utils';
import { ZoneManagementContext } from '../ZoneManagementContext';
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

export const getEditButtonState = (memberId?: string, selectorsQuery?: UseQueryResult, tagsQuery?: UseQueryResult) => {
    return (
        !!memberId ||
        (selectorsQuery?.isLoading && tagsQuery?.isLoading) ||
        (selectorsQuery?.isError && tagsQuery?.isError)
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

    const tagsQuery = useQuery({
        queryKey: ['zone-management', 'tags'],
        queryFn: async () => {
            return apiClient
                .getAssetGroupTags({
                    params: {
                        counts: true,
                    },
                })
                .then((res) => {
                    return res.data.data['tags'];
                });
        },
    });

    const selectorsQuery = useInfiniteQuery<{
        items: AssetGroupTagSelector[];
        nextPageParam?: { skip: number; limit: number };
    }>({
        queryKey: ['zone-management', 'tags', tagId, 'selectors'],
        queryFn: async ({ pageParam = { skip: 0, limit: 25 } }) => {
            if (!tagId)
                return {
                    items: [],
                    nextPageParam: undefined,
                };
            return apiClient
                .getAssetGroupTagSelectors(tagId, {
                    params: {
                        skip: pageParam.skip,
                        limit: pageParam.limit,
                        counts: true,
                    },
                })
                .then((res) => {
                    const items = res.data.data['selectors'];
                    const hasMore = pageParam.skip + pageParam.limit < res.data.count;

                    return {
                        items,
                        nextPageParam: hasMore ? { skip: pageParam.skip + 25, limit: 25 } : undefined,
                    };
                });
        },
        getNextPageParam: (lastPage) => {
            return lastPage.nextPageParam;
        },
    });

    const membersQuery = useInfiniteQuery<{
        items: AssetGroupTagMemberListItem[];
        nextPageParam?: { skip: number; limit: number };
    }>({
        queryKey: ['zone-management', 'tagId', tagId, 'selectorId', selectorId, membersListSortOrder],
        queryFn: async ({ pageParam = { skip: 0, limit: 25 } }) => {
            const tag = tagId || 1;

            const sort_by = membersListSortOrder === 'asc' ? 'name' : '-name';

            if (selectorId) {
                return apiClient
                    .getAssetGroupTagSelectorMembers(tag, selectorId, pageParam.skip, pageParam.limit, sort_by)
                    .then((res) => {
                        const items = res.data.data['members'];
                        const hasMore = pageParam.skip + pageParam.limit < res.data.count;
                        return { items, nextPageParam: hasMore ? { skip: pageParam.skip + 25, limit: 25 } : undefined };
                    });
            } else {
                return apiClient.getAssetGroupTagMembers(tag, pageParam.skip, pageParam.limit, sort_by).then((res) => {
                    const items = res.data.data['members'];
                    const hasMore = pageParam.skip + pageParam.limit < res.data.count;
                    return { items, nextPageParam: hasMore ? { skip: pageParam.skip + 25, limit: 25 } : undefined };
                });
            }
        },
        getNextPageParam: (lastPage) => {
            return lastPage.nextPageParam;
        },
    });

    const showEditButton = !getEditButtonState(memberId, selectorsQuery, tagsQuery);

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
                    <DetailsList
                        title={location.pathname.includes('label') ? 'Labels' : 'Tiers'}
                        listQuery={tagsQuery}
                        selected={tagId}
                        onSelect={(id) => {
                            navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${id}`);
                        }}
                    />
                    <SelectorsList
                        listQuery={selectorsQuery}
                        selected={selectorId}
                        onSelect={(id) => {
                            navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/selector/${id}`);
                        }}
                    />
                    <MembersList
                        listQuery={membersQuery}
                        selected={memberId}
                        onClick={(id) => {
                            if (selectorId) {
                                navigate(
                                    `/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/selector/${selectorId}/member/${id}`
                                );
                            } else {
                                navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/member/${id}`);
                            }
                        }}
                        sortOrder={membersListSortOrder}
                        onChangeSortOrder={setMembersListSortOrder}
                    />
                </div>
                <div className='basis-1/3'>
                    <SelectedDetails />
                </div>
            </div>
        </div>
    );
};

export default Details;
