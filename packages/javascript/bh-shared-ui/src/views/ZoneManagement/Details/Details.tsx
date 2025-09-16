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
import { AssetGroupTagTypeLabel, AssetGroupTagTypeOwned, AssetGroupTagTypeZone } from 'js-client-library';
import { FC, useContext, useState } from 'react';
import { UseQueryResult } from 'react-query';
import { useParams } from 'react-router-dom';
import { AppLink } from '../../../components/Navigation';
import { useHighestPrivilegeTagId } from '../../../hooks';
import {
    useSelectorMembersInfiniteQuery,
    useSelectorsInfiniteQuery,
    useTagMembersInfiniteQuery,
    useTagsQuery,
} from '../../../hooks/useAssetGroupTags';
import { useEnvironmentIdList } from '../../../hooks/useEnvironmentIdList';
import {
    detailsPath,
    labelsPath,
    membersPath,
    privilegeZonesPath,
    savePath,
    selectorsPath,
    zonesPath,
} from '../../../routes';
import { SortOrder } from '../../../types';
import { useAppNavigate } from '../../../utils';
import { useTagFormUtils } from '../Save/TagForm/utils';
import { ZoneManagementContext } from '../ZoneManagementContext';
import { getTagUrlValue } from '../utils';
import { MembersList } from './MembersList';
import SearchBar from './SearchBar';
import { SelectedDetails } from './SelectedDetails';
import { SelectorsList } from './SelectorsList';
import { TagList } from './TagList';

export const getSavePath = (
    zoneId: string | undefined,
    labelId: string | undefined,
    selectorId: string | undefined
) => {
    const tagType = !labelId ? zonesPath : labelsPath;
    let tagPathId = '';

    if (zoneId || labelId) {
        tagPathId = tagType === 'zones' ? zoneId ?? '' : labelId ?? '';
    }

    if (tagPathId === '') return;

    const selectorIdPath = selectorId ? `${selectorsPath}/${selectorId}/${savePath}` : savePath;

    return `/${privilegeZonesPath}/${tagType}/${tagPathId}/${selectorIdPath}`;
};

export const getEditButtonState = (
    memberId?: string,
    selectorsQuery?: UseQueryResult,
    zonesQuery?: UseQueryResult,
    labelsQuery?: UseQueryResult
) => {
    return (
        !!memberId ||
        (selectorsQuery?.isLoading && zonesQuery?.isLoading && labelsQuery?.isLoading) ||
        (selectorsQuery?.isError && zonesQuery?.isError && labelsQuery?.isError)
    );
};

const Details: FC = () => {
    const navigate = useAppNavigate();
    const { isLabelLocation } = useTagFormUtils();
    const [membersListSortOrder, setMembersListSortOrder] = useState<SortOrder>('asc');
    const [selectorsListSortOrder, setSelectorsListSortOrder] = useState<SortOrder>('asc');
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const { zoneId = topTagId?.toString(), labelId, selectorId, memberId } = useParams();
    const tagType = !labelId ? zonesPath : labelsPath;
    const environments = useEnvironmentIdList([
        {
            path: `/${privilegeZonesPath}/${zonesPath}/${tagType}/${detailsPath}`,
            caseSensitive: false,
            end: false,
        },
    ]);

    const tagId = labelId === undefined ? zoneId : labelId;

    const context = useContext(ZoneManagementContext);
    if (!context) {
        throw new Error('Details must be used within a ZoneManagementContext.Provider');
    }
    const { InfoHeader } = context;

    const zonesQuery = useTagsQuery({ select: (tags) => tags.filter((tag) => tag.type === AssetGroupTagTypeZone) });

    const labelsQuery = useTagsQuery({
        select: (tags) =>
            tags.filter((tag) => tag.type === AssetGroupTagTypeLabel || tag.type === AssetGroupTagTypeOwned),
    });

    const selectorsQuery = useSelectorsInfiniteQuery(tagId, selectorsListSortOrder, environments);

    const selectorMembersQuery = useSelectorMembersInfiniteQuery(tagId, selectorId, membersListSortOrder, environments);

    const tagMembersQuery = useTagMembersInfiniteQuery(tagId, membersListSortOrder, environments);

    const showEditButton = !getEditButtonState(memberId, selectorsQuery, zonesQuery, labelsQuery);

    const saveLink = getSavePath(zoneId, labelId, selectorId);

    return (
        <div className='h-full'>
            <div className='flex mt-6'>
                <div className='w-1/3'>{InfoHeader && <InfoHeader />}</div>
                <div className='w-1/3 flex justify-end'>
                    <SearchBar />
                </div>
                <div className='w-1/3 ml-8'>
                    {showEditButton && (
                        <Button
                            asChild={showEditButton || !saveLink}
                            variant={'secondary'}
                            disabled={showEditButton || !saveLink}>
                            <AppLink to={saveLink || ''}>Edit</AppLink>
                        </Button>
                    )}
                </div>
            </div>
            <div className='flex gap-8 mt-4 h-full'>
                <div className='flex basis-2/3 bg-neutral-light-2 dark:bg-neutral-dark-2 rounded-lg shadow-outer-1 *:w-1/3 h-fit'>
                    {isLabelLocation ? (
                        <TagList
                            title={'Labels'}
                            listQuery={labelsQuery}
                            selected={tagId}
                            onSelect={(id) => {
                                navigate(`/${privilegeZonesPath}/${getTagUrlValue(labelId)}/${id}/${detailsPath}`);
                            }}
                        />
                    ) : (
                        <TagList
                            title={'Zones'}
                            listQuery={zonesQuery}
                            selected={tagId}
                            onSelect={(id) => {
                                navigate(`/${privilegeZonesPath}/${getTagUrlValue(labelId)}/${id}/${detailsPath}`);
                            }}
                        />
                    )}

                    <SelectorsList
                        listQuery={selectorsQuery}
                        onChangeSortOrder={setSelectorsListSortOrder}
                        onSelect={(id) => {
                            navigate(
                                `/${privilegeZonesPath}/${getTagUrlValue(labelId)}/${tagId}/${selectorsPath}/${id}/${detailsPath}`
                            );
                        }}
                        selected={selectorId}
                        sortOrder={selectorsListSortOrder}
                    />
                    {selectorId !== undefined ? (
                        <MembersList
                            listQuery={selectorMembersQuery}
                            selected={memberId}
                            onClick={(id) => {
                                navigate(
                                    `/${privilegeZonesPath}/${getTagUrlValue(labelId)}/${tagId}/${selectorsPath}/${selectorId}/${membersPath}/${id}/${detailsPath}`
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
                                navigate(
                                    `/${privilegeZonesPath}/${getTagUrlValue(labelId)}/${tagId}/${membersPath}/${id}/${detailsPath}`
                                );
                            }}
                            sortOrder={membersListSortOrder}
                            onChangeSortOrder={setMembersListSortOrder}
                        />
                    )}
                </div>
                <div className='basis-1/3 h-full'>
                    <SelectedDetails />
                </div>
            </div>
        </div>
    );
};

export default Details;
