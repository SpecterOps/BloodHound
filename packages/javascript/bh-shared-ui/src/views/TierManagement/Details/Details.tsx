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
import { AssetGroupTagSelectorsListItem, AssetGroupTagsListItem } from 'js-client-library';
import { FC } from 'react';
import { UseQueryResult, useQuery } from 'react-query';
import { Link, useParams, useLocation } from 'react-router-dom';
import { AppIcon } from '../../../components';
import { ROUTE_TIER_MANAGEMENT_DETAILS } from '../../../routes';
import { apiClient, useAppNavigate } from '../../../utils';
import { TIER_ZERO_ID, getTagUrlValue } from '../utils';
import { DetailsList } from './DetailsList';
import { MembersList } from './MembersList';
import { SelectedDetails } from './SelectedDetails';

const getSavePath = (tierId: string | undefined, labelId: string | undefined, selectorId: string | undefined) => {
    const savePath = '/tier-management/save';

    if (selectorId && labelId) return `/tier-management/save/label/${labelId}/selector/${selectorId}`;
    if (selectorId && tierId) return `/tier-management/save/tier/${tierId}/selector/${selectorId}`;

    if (!selectorId && labelId) return `/tier-management/save/label/${labelId}`;
    if (!selectorId && tierId) return `/tier-management/save/tier/${tierId}`;

    return savePath;
};

const getItemCount = (
    tagId: string | undefined,
    tagsQuery: UseQueryResult<AssetGroupTagsListItem[]>,
    selectorId: string | undefined,
    selectorsQuery: UseQueryResult<AssetGroupTagSelectorsListItem[]>
) => {
    if (selectorId !== undefined) {
        const selectedSelector = selectorsQuery.data?.find((selector) => {
            return selectorId === selector.id.toString();
        });
        return selectedSelector?.counts?.members || 0;
    } else if (tagId !== undefined) {
        const selectedTag = tagsQuery.data?.find((tag) => {
            return tagId === tag.id.toString();
        });
        return selectedTag?.counts?.members || 0;
    } else {
        return 0;
    }
};

export const getEditButtonState = (
    memberId: string | undefined,
    selectorId: string | undefined,
    selectorsQuery: UseQueryResult,
    tagsQuery: UseQueryResult
) => {
    return (
        !!memberId ||
        !selectorId ||
        (selectorsQuery.isLoading && tagsQuery.isLoading) ||
        (selectorsQuery.isError && tagsQuery.isError)
    );
};

const Details: FC = () => {
    const navigate = useAppNavigate();
    const { tierId = TIER_ZERO_ID, labelId, selectorId, memberId } = useParams();
    const tagId = labelId === undefined ? tierId : labelId;
    const location = useLocation();
    const value = location.pathname.includes('tier/') ? 'Tier' : 'Label';

    const tagsQuery = useQuery({
        queryKey: ['tier-management', 'tags'],
        queryFn: async () => {
            return apiClient.getAssetGroupTags({ params: { counts: true } }).then((res) => {
                return res.data.data['tags'];
            });
        },
    });

    const selectorsQuery = useQuery({
        queryKey: ['tier-management', 'tags', tagId, 'selectors'],
        queryFn: async () => {
            if (!tagId) return [];
            return apiClient.getAssetGroupTagSelectors(tagId, { params: { counts: true } }).then((res) => {
                return res.data.data['selectors'];
            });
        },
    });

    const showEditButton = !getEditButtonState(memberId, selectorId, selectorsQuery, tagsQuery);

    return (
        <div>
            <div className='flex mt-6 gap-8'>
                <div className='flex justify-around basis-2/3'>
                    <div className='flex justify-start gap-4 items-center basis-2/3'>
                        <div className='flex items-center align-middle gap-4'>
                            <Button
                                onClick={() => {
                                    if (value === 'Label') {
                                        navigate(`/tier-management/save/label`)
                                    }
                                    if (value === 'Tier') {
                                        navigate(`/tier-management/save/tier`)
                                    }
                                }}>
                                Create {value}
                            </Button>
                            <Button
                                variant="secondary"
                                onClick={() => {
                                    navigate(`/tier-management/save/${getTagUrlValue(labelId)}/${tagId}/selector`)
                                }}>
                                Create Selector
                            </Button>
                            <div className='hidden'>
                                <div>
                                    <AppIcon.Info className='mr-4 ml-2' size={24} />
                                </div>
                                <span>
                                    To create additional tiers{' '}
                                    <Button
                                        variant='text'
                                        asChild
                                        className='p-0 text-base text-secondary dark:text-secondary-variant-2'>
                                        <a href='#'>contact sales</a>
                                    </Button>{' '}
                                    in order to upgrade for multi-tier analysis.
                                </span>
                            </div>
                        </div>
                    </div>
                    <div className='flex justify-start basis-1/3'>
                        <input type='text' placeholder='search' className='hidden' />
                    </div>
                </div>

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
                        title={labelId ? 'Labels' : 'Tiers'}
                        listQuery={tagsQuery}
                        selected={tagId}
                        onSelect={(id) => {
                            navigate(
                                `/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}/${getTagUrlValue(labelId)}/${id}`
                            );
                        }}
                    />
                    <DetailsList
                        title='Selectors'
                        listQuery={selectorsQuery}
                        selected={selectorId}
                        onSelect={(id) => {
                            navigate(
                                `/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}/${getTagUrlValue(labelId)}/${tagId}/selector/${id}`
                            );
                        }}
                    />
                    <MembersList
                        itemCount={getItemCount(tagId, tagsQuery, selectorId, selectorsQuery)}
                        onClick={(id) => {
                            navigate(
                                `/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}/${getTagUrlValue(labelId)}/${tagId}/selector/${selectorId}/member/${id}`
                            );
                        }}
                        selected={memberId}
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
