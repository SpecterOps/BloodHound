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

import { AssetGroupTag } from 'js-client-library';
import { FC, useContext } from 'react';
import { useQuery } from 'react-query';
import { useEnvironmentIdList, useHighestPrivilegeTagId, usePZPathParams } from '../../../hooks';
import { privilegeZonesPath } from '../../../routes';
import { apiClient } from '../../../utils/api';
import { useAppNavigate } from '../../../utils/searchParams';
import SearchBar from '../Details/SearchBar';
import { SelectedDetailsTabs } from '../Details/SelectedDetailsTabs';
import { PrivilegeZonesContext } from '../PrivilegeZonesContext';
import { ObjectsAccordion } from './ObjectsAccordion';
// IMPORTANT! BED-6836: Uncomment below when details list is ready and we want to set tab context on click of item
// import { useSelectedDetailsTabsContext } from './SelectedDetailsTabs/SelectedDetailsTabsContext';
// import { DetailsTabOption } from './utils';

const useTagObjectCounts = (tagId: string | undefined, environments: string[]) =>
    useQuery({
        queryKey: ['asset-group-tags-count', tagId, ...environments],
        queryFn: async ({ signal }) => {
            if (!tagId) return Promise.reject('No Tag ID available for tag counts request');

            return apiClient.getAssetGroupTagMembersCount(tagId, environments, { signal }).then((res) => res.data.data);
        },
        enabled: !!tagId,
    });

const useRuleObjectCounts = (tagId: string | undefined, ruleId: string | undefined, environments: string[]) =>
    useQuery({
        queryKey: ['asset-group-tags-count', tagId, 'rule', ruleId, ...environments],
        queryFn: async ({ signal }) => {
            if (!tagId) return Promise.reject('No Tag ID available for Rule counts request');
            if (!ruleId) return Promise.reject('No Rule ID available for Rule counts request');

            return apiClient
                .getAssetGroupTagRuleMembersCount(tagId, ruleId, environments, { signal })
                .then((res) => res.data.data);
        },
        enabled: !!tagId && !!ruleId,
    });

const useObjectCounts = () => {
    const { ruleId, tagId } = usePZPathParams();

    const environments = useEnvironmentIdList([{ path: `/${privilegeZonesPath}/*`, caseSensitive: false, end: false }]);

    const tagCounts = useTagObjectCounts(tagId, environments);
    const ruleCounts = useRuleObjectCounts(tagId, ruleId, environments);

    if (ruleId) return ruleCounts;
    return tagCounts;
};

const Details: FC = () => {
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const { zoneId = topTagId?.toString(), tagTypeDisplay, tagId: defaultTagId, tagDetailsLink } = usePZPathParams();
    const tagId = !defaultTagId ? zoneId : defaultTagId;
    // IMPORTANT! BED-6836: Uncomment below when details list is ready and we want to set tab context on click of item
    // const { setSelectedDetailsTab } = useSelectedDetailsTabsContext();
    // Add Below function on click of each list item
    // const handleSelectedTab = (tabValue: DetailsTabOption) => setSelectedDetailsTab(tabValue);

    const navigate = useAppNavigate();

    const context = useContext(PrivilegeZonesContext);
    if (!context) {
        throw new Error('Details must be used within a PrivilegeZonesContext.Provider');
    }
    const { InfoHeader, ZoneSelector } = context;

    const objectCountsQuery = useObjectCounts();

    if (!tagId) return null;

    const handleZoneClick = (zone: AssetGroupTag) => {
        navigate(tagDetailsLink(zone.id));
    };

    return (
        <div className='h-full'>
            <div className='flex mt-6'>
                <div className='flex-wrap-reverse basis-2/3 justify-between items-center'>
                    <InfoHeader />
                </div>
            </div>
            <div className='flex gap-8 mt-4 h-full'>
                <div className='flex flex-col gap-2 basis-2/3 bg-neutral-2 py-4 min-w-0 rounded shadow-outer-1 h-full'>
                    <h2 className='font-bold text-xl pl-4 pb-1'>{tagTypeDisplay} Details</h2>
                    <div className='flex justify-between w-full pb-4 border-b border-neutral-3'>
                        <div className='flex gap-6 pl-4'>
                            {ZoneSelector && <ZoneSelector onZoneClick={handleZoneClick} />}
                            <div className='flex items-center px-4 rounded h-10 border-contrast border'>
                                TITANCORP.LOCAL
                            </div>
                        </div>
                        <SearchBar showTags={false} />
                    </div>
                    <div className='flex overflow-auto max-lg:flex-col'>
                        <div className='basis-1/2'>
                            <ul className='h-dvh overflow-y-scroll'></ul>
                        </div>
                        <div className='basis-1/2'>
                            <ObjectsAccordion
                                kindCounts={objectCountsQuery.data?.counts || {}}
                                totalCount={objectCountsQuery.data?.total_count || 0}
                            />
                        </div>
                    </div>
                </div>
                <SelectedDetailsTabs />
            </div>
        </div>
    );
};

export default Details;
