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
import { FC } from 'react';
import { UseQueryResult, useQuery } from 'react-query';
import { Link, useParams } from 'react-router-dom';
import { AppIcon, CreateMenu } from '../../../components';
import { ROUTE_TIER_MANAGEMENT_DETAILS } from '../../../routes';
import { apiClient, useAppNavigate } from '../../../utils';
import { DetailsList } from './DetailsList';
import { MembersList } from './MembersList';
import { SelectedDetails } from './SelectedDetails';

export const getEditPath = (tagId: string | undefined, selectorId: string | undefined) => {
    const editPath = '/tier-management/edit';

    if (selectorId && tagId) return `/tier-management/edit/tag/${tagId}/selector/${selectorId}`;

    if (!selectorId && tagId) return `/tier-management/edit/tag/${tagId}`;

    return editPath;
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
    const { tagId = '1', selectorId, memberId } = useParams();

    const tagsQuery = useQuery({
        queryKey: ['asset-group-tags'],
        queryFn: async () => {
            return apiClient.getAssetGroupTags({ params: { counts: true } }).then((res) => {
                return res.data.data['tags'];
            });
        },
    });

    const selectorsQuery = useQuery({
        queryKey: ['asset-group-selectors', tagId],
        queryFn: async () => {
            if (!tagId) return [];
            return apiClient.getAssetGroupTagSelectors(tagId, { params: { counts: true } }).then((res) => {
                return res.data.data['selectors'];
            });
        },
    });

    const objectsQuery = useQuery({
        queryKey: ['asset-group-members', tagId, selectorId],
        queryFn: async () => {
            if (!tagId) return 0;

            if (!selectorId)
                return apiClient.getAssetGroupTagMembers(tagId, 0, 1).then((res) => {
                    return res.data.count;
                });

            return apiClient.getAssetGroupSelectorMembers(tagId, selectorId, 0, 1).then((res) => {
                return res.data.count;
            });
        },
    });

    const showEditButton = !getEditButtonState(memberId, selectorId, selectorsQuery, tagsQuery);

    return (
        <div>
            <div className='flex mt-6 gap-8'>
                <CreateMenu
                    createMenuTitle='Create Selector'
                    disabled={!tagId}
                    menuItems={[
                        {
                            title: 'Create Selector',
                            onClick: () => {
                                navigate(getEditPath(tagId, selectorId));
                            },
                        },
                    ]}
                />
                <div className='flex justify-around basis-2/3'>
                    <div className='flex justify-start gap-4 items-center basis-2/3 invisible'>
                        <div className='flex items-center align-middle'>
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
                    <div className='flex justify-start basis-1/3'>
                        <input type='text' placeholder='search' className='hidden' />
                    </div>
                </div>

                <div className='basis-1/3'>
                    {showEditButton && (
                        <Button asChild variant={'secondary'} disabled={showEditButton}>
                            <Link to={getEditPath(tagId, selectorId)}>Edit</Link>
                        </Button>
                    )}
                </div>
            </div>
            <div className='flex gap-8 mt-4 grow-1'>
                <div className='flex basis-2/3 bg-neutral-light-2 dark:bg-neutral-dark-2 rounded-lg shadow-outer-1 h-full *:grow-0 *:basis-1/3'>
                    <div>
                        <DetailsList
                            title='Tiers'
                            listQuery={tagsQuery}
                            selected={tagId}
                            onSelect={(id) => {
                                navigate(`/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}/tag/${id}`);
                            }}
                        />
                    </div>
                    <div className='border-neutral-light-3 dark:border-neutral-dark-3'>
                        <DetailsList
                            title='Selectors'
                            listQuery={selectorsQuery}
                            selected={selectorId}
                            onSelect={(id) => {
                                navigate(
                                    `/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}/tag/${tagId}/selector/${id}`
                                );
                            }}
                        />
                    </div>
                    <div>
                        <MembersList
                            itemCount={objectsQuery.data}
                            onClick={(id) => {
                                navigate(
                                    `/tier-management/${ROUTE_TIER_MANAGEMENT_DETAILS}/tag/${tagId}/selector/${selectorId}/member/${id}`
                                );
                            }}
                            selected={memberId}
                            selectedSelector={selectorId}
                            selectedTag={tagId}
                        />
                    </div>
                </div>
                <div className='basis-1/3'>
                    <SelectedDetails />
                </div>
            </div>
        </div>
    );
};

export default Details;
