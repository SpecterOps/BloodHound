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
import { AssetGroupTag, AssetGroupTagSelector } from 'js-client-library';
import { FC, useContext } from 'react';
import { UseQueryResult, useQuery } from 'react-query';
import { useLocation, useParams } from 'react-router-dom';
import { AppLink } from '../../../components/Navigation';
import { useHighestPrivilegeTag } from '../../../hooks';
import { apiClient, useAppNavigate } from '../../../utils';
import { ZoneManagementContext } from '../ZoneManagementContext';
import { getTagUrlValue } from '../utils';
import { DetailsList } from './DetailsList';
import { MembersList } from './MembersList';
import { SelectedDetails } from './SelectedDetails';

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

const getItemCount = (
    tagId: string | undefined,
    tagsQuery: UseQueryResult<AssetGroupTag[]>,
    selectorId: string | undefined,
    selectorsQuery: UseQueryResult<AssetGroupTagSelector[]>
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
    const topTagId = useHighestPrivilegeTag()?.id.toString();
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
                {InfoHeader && <InfoHeader />}
                <div className='basis-1/3'>
                    {showEditButton && (
                        <Button asChild variant={'secondary'} disabled={showEditButton}>
                            <AppLink to={getSavePath(tierId, labelId, selectorId)}>Edit</AppLink>
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
                    <DetailsList
                        title='Selectors'
                        listQuery={selectorsQuery}
                        selected={selectorId}
                        onSelect={(id) => {
                            navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/selector/${id}`);
                        }}
                    />
                    <MembersList
                        itemCount={getItemCount(tagId, tagsQuery, selectorId, selectorsQuery)}
                        onClick={(id) => {
                            if (selectorId) {
                                navigate(
                                    `/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/selector/${selectorId}/member/${id}`
                                );
                            } else {
                                navigate(`/zone-management/details/${getTagUrlValue(labelId)}/${tagId}/member/${id}`);
                            }
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
