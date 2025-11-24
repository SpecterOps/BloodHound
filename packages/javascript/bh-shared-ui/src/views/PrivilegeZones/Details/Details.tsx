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

import { AssetGroupTagTypeLabel, AssetGroupTagTypeOwned, AssetGroupTagTypeZone } from 'js-client-library';
import { FC, useContext, useState } from 'react';
import { UseQueryResult } from 'react-query';
import { useHighestPrivilegeTagId, usePZPathParams } from '../../../hooks';
import {
    useSelectorMembersInfiniteQuery,
    useSelectorsInfiniteQuery,
    useTagMembersInfiniteQuery,
    useTagsQuery,
} from '../../../hooks/useAssetGroupTags';
import { useEnvironmentIdList } from '../../../hooks/useEnvironmentIdList';
import { privilegeZonesPath } from '../../../routes';
import { SortOrder } from '../../../types';
import { useAppNavigate } from '../../../utils';
import { PZEditButton } from '../PZEditButton';
import { PrivilegeZonesContext } from '../PrivilegeZonesContext';
import { PageDescription } from '../fragments';
import { MembersList } from './MembersList';
import SearchBar from './SearchBar';
import { SelectedDetails } from './SelectedDetails';
import { SelectorsList } from './SelectorsList';
import { TagList } from './TagList';

const getEditButtonState = (
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
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const {
        isLabelPage,
        zoneId = topTagId?.toString(),
        labelId,
        selectorId,
        memberId,
        tagId: defaultTagId,
        tagDetailsLink,
        ruleDetailsLink,
        objectDetailsLink,
    } = usePZPathParams();
    const tagId = !defaultTagId ? zoneId : defaultTagId;

    const [membersListSortOrder, setMembersListSortOrder] = useState<SortOrder>('asc');
    const [selectorsListSortOrder, setSelectorsListSortOrder] = useState<SortOrder>('asc');

    const environments = useEnvironmentIdList([{ path: `/${privilegeZonesPath}/*`, caseSensitive: false, end: false }]);

    const context = useContext(PrivilegeZonesContext);
    if (!context) {
        throw new Error('Details must be used within a PrivilegeZonesContext.Provider');
    }
    const { InfoHeader } = context;

    const zonesQuery = useTagsQuery({
        select: (tags) => tags.filter((tag) => tag.type === AssetGroupTagTypeZone),
        enabled: !!zoneId,
    });

    const labelsQuery = useTagsQuery({
        select: (tags) =>
            tags.filter((tag) => tag.type === AssetGroupTagTypeLabel || tag.type === AssetGroupTagTypeOwned),
        enabled: !!labelId,
    });

    const selectorsQuery = useSelectorsInfiniteQuery(tagId, selectorsListSortOrder, environments);
    const selectorMembersQuery = useSelectorMembersInfiniteQuery(tagId, selectorId, membersListSortOrder, environments);
    const tagMembersQuery = useTagMembersInfiniteQuery(tagId, membersListSortOrder, environments);

    if (!tagId) return null;
    return (
        <div className='h-full'>
            <PageDescription />
            <div className='flex mt-6'>
                <div className='flex flex-wrap basis-2/3 justify-between'>
                    {InfoHeader && <InfoHeader />}
                    <SearchBar />
                </div>
                <div className='basis-1/3 ml-8'>
                    <PZEditButton
                        showEditButton={!getEditButtonState(memberId, selectorsQuery, zonesQuery, labelsQuery)}
                    />
                </div>
            </div>
            <div className='flex gap-8 mt-4 h-full'>
                <div className='flex basis-2/3 bg-neutral-2 min-w-0 rounded-lg shadow-outer-1 h-fit'>
                    {isLabelPage ? (
                        <TagList
                            title='Labels'
                            listQuery={labelsQuery}
                            selected={tagId}
                            onSelect={(id) => {
                                navigate(tagDetailsLink(id, 'labels'));
                            }}
                        />
                    ) : (
                        <TagList
                            title='Zones'
                            listQuery={zonesQuery}
                            selected={tagId}
                            onSelect={(id) => {
                                navigate(tagDetailsLink(id, 'zones'));
                            }}
                        />
                    )}
                    <SelectorsList
                        listQuery={selectorsQuery}
                        selected={selectorId}
                        onSelect={(id) => {
                            navigate(ruleDetailsLink(tagId, id));
                        }}
                        sortOrder={selectorsListSortOrder}
                        onChangeSortOrder={setSelectorsListSortOrder}
                    />
                    {selectorId !== undefined ? (
                        <MembersList
                            listQuery={selectorMembersQuery}
                            selected={memberId}
                            onClick={(id) => {
                                navigate(objectDetailsLink(tagId, id, selectorId));
                            }}
                            sortOrder={membersListSortOrder}
                            onChangeSortOrder={setMembersListSortOrder}
                        />
                    ) : (
                        <MembersList
                            listQuery={tagMembersQuery}
                            selected={memberId}
                            onClick={(id) => {
                                navigate(objectDetailsLink(tagId, id));
                            }}
                            sortOrder={membersListSortOrder}
                            onChangeSortOrder={setMembersListSortOrder}
                        />
                    )}
                </div>
                <div className='flex basis-1/3 h-full'>
                    <SelectedDetails />
                </div>
            </div>
        </div>
    );
};

export default Details;
