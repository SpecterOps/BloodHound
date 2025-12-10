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

import { Tabs, TabsList, TabsTrigger } from '@bloodhoundenterprise/doodleui';
import { CircularProgress } from '@mui/material';
import { AssetGroupTagTypeLabel, AssetGroupTagTypeOwned, AssetGroupTagTypeZone } from 'js-client-library';
import { FC, Suspense, useContext, useState } from 'react';
import { UseQueryResult } from 'react-query';
import { useHighestPrivilegeTagId, usePZPathParams } from '../../../hooks';
import {
    useRuleMembersInfiniteQuery,
    useRulesInfiniteQuery,
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
import { RulesList } from './RulesList';
import SearchBar from './SearchBar';
import { SelectedDetailsV2 } from './SelectedDetailsV2';
import { TagList } from './TagList';

const getEditButtonState = (
    memberId?: string,
    rulesQuery?: UseQueryResult,
    zonesQuery?: UseQueryResult,
    labelsQuery?: UseQueryResult
) => {
    return (
        !!memberId ||
        (rulesQuery?.isLoading && zonesQuery?.isLoading && labelsQuery?.isLoading) ||
        (rulesQuery?.isError && zonesQuery?.isError && labelsQuery?.isError)
    );
};

const Details: FC = () => {
    const navigate = useAppNavigate();
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const {
        isLabelPage,
        zoneId = topTagId?.toString(),
        labelId,
        ruleId,
        memberId,
        tagId: defaultTagId,
        tagDetailsLink,
        ruleDetailsLink,
        objectDetailsLink,
        tagTypeDisplay,
    } = usePZPathParams();
    const tagId = !defaultTagId ? zoneId : defaultTagId;

    const getFromPathParams = () => {
        if (memberId) return '3';
        if (ruleId) return '2';
        return '1';
    };

    const [currentTab, setCurrentTab] = useState(getFromPathParams()); // placeholder

    const [membersListSortOrder, setMembersListSortOrder] = useState<SortOrder>('asc');
    const [rulesListSortOrder, setRulesListSortOrder] = useState<SortOrder>('asc');

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

    const rulesQuery = useRulesInfiniteQuery(tagId, rulesListSortOrder, environments);
    const ruleMembersQuery = useRuleMembersInfiniteQuery(tagId, ruleId, membersListSortOrder, environments);
    const tagMembersQuery = useTagMembersInfiniteQuery(tagId, membersListSortOrder, environments);

    const handleSetTab = (value: string) => {
        setCurrentTab(value);
    };

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
                    <PZEditButton showEditButton={!getEditButtonState(memberId, rulesQuery, zonesQuery, labelsQuery)} />
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
                                handleSetTab('1');
                                navigate(tagDetailsLink(id, 'labels'));
                            }}
                        />
                    ) : (
                        <TagList
                            title='Zones'
                            listQuery={zonesQuery}
                            selected={tagId}
                            onSelect={(id) => {
                                handleSetTab('1');
                                navigate(tagDetailsLink(id, 'zones'));
                            }}
                        />
                    )}
                    <RulesList
                        listQuery={rulesQuery}
                        selected={ruleId}
                        onSelect={(id) => {
                            handleSetTab('2');
                            navigate(ruleDetailsLink(tagId, id));
                        }}
                        sortOrder={rulesListSortOrder}
                        onChangeSortOrder={setRulesListSortOrder}
                    />
                    {ruleId !== undefined ? (
                        <MembersList
                            listQuery={ruleMembersQuery}
                            selected={memberId}
                            onClick={(id) => {
                                handleSetTab('3');
                                navigate(objectDetailsLink(tagId, id, ruleId));
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
                <div className='flex flex-col w-[400px]'>
                    <Tabs
                        defaultValue={getFromPathParams()}
                        value={currentTab}
                        className='w-full mb-4'
                        onValueChange={(value) => {
                            setCurrentTab(value);
                        }}>
                        <TabsList className='w-full flex justify-start'>
                            <TabsTrigger value={'1'}>{tagTypeDisplay}</TabsTrigger>
                            <TabsTrigger disabled={!ruleId} value={'2'}>
                                Rule
                            </TabsTrigger>
                            <TabsTrigger disabled={!memberId} value={'3'}>
                                Object
                            </TabsTrigger>
                        </TabsList>
                    </Tabs>
                    <Suspense
                        fallback={
                            <div className='absolute inset-0 flex items-center justify-center'>
                                <CircularProgress color='primary' size={80} />
                            </div>
                        }>
                        <SelectedDetailsV2 currentTab={currentTab} />
                    </Suspense>
                </div>
            </div>
        </div>
    );
};

export default Details;
