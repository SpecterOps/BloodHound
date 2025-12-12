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
import { SelectedDetailsTabs } from './SelectedDetailsTabs';
import { TagList } from './TagList';
import { DetailsTabOption, detailsTabOptions } from './utils';

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
    } = usePZPathParams();
    const tagId = !defaultTagId ? zoneId : defaultTagId;

    // Need to know which side panel tab to pick on refresh
    const selectedDetailsTabFromPathParams = () => {
        if (memberId) return detailsTabOptions[3];
        if (ruleId) return detailsTabOptions[2];
        return detailsTabOptions[1];
    };

    // Keeps track of the list item tab but on first render set whatever is in the params to match the selected list
    const [currentDetailsTab, setCurrentDetailsTab] = useState<DetailsTabOption>(selectedDetailsTabFromPathParams());

    // Set side tab on click of a list
    const handleClickTagList = (list: DetailsTabOption) => {
        setCurrentDetailsTab(list);
    };

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
                                handleClickTagList(detailsTabOptions[1]);
                                navigate(tagDetailsLink(id, 'labels'));
                            }}
                        />
                    ) : (
                        <TagList
                            title='Zones'
                            listQuery={zonesQuery}
                            selected={tagId}
                            onSelect={(id) => {
                                handleClickTagList(detailsTabOptions[1]);
                                navigate(tagDetailsLink(id, 'zones'));
                            }}
                        />
                    )}
                    <RulesList
                        listQuery={rulesQuery}
                        selected={ruleId}
                        onSelect={(id) => {
                            handleClickTagList(detailsTabOptions[2]);
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
                                handleClickTagList(detailsTabOptions[3]);
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
                                handleClickTagList(detailsTabOptions[3]);
                                navigate(objectDetailsLink(tagId, id));
                            }}
                            sortOrder={membersListSortOrder}
                            onChangeSortOrder={setMembersListSortOrder}
                        />
                    )}
                </div>
                {/* IMPORTANT!!! Revert TW UTILITIES */}
                <div className='flex flex-col w-[400px]'>
                    {/* IMPORTANT!!!! Revert this to the selected details original and move this to new details  */}
                    <SelectedDetailsTabs
                        currentDetailsTab={currentDetailsTab}
                        onTabClick={(value: DetailsTabOption) => setCurrentDetailsTab(value)}
                    />
                </div>
            </div>
        </div>
    );
};

export default Details;
