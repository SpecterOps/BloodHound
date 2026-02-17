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

import { Alert, AlertTitle } from '@mui/material';
import { AssetGroupTagMemberListItem } from 'js-client-library';
import { FC } from 'react';
import { useHighestPrivilegeTagId, useObjectCounts, usePZPathParams, useRuleInfo } from '../../../hooks';
import { useAppNavigate } from '../../../utils';
import { usePZContext } from '../PrivilegeZonesContext';
import { PageDescription } from '../fragments';
import { ObjectTabValue } from '../utils';
import { ObjectsAccordion } from './ObjectsAccordion';
import { RulesAccordion } from './RulesAccordion';
import SearchBar from './SearchBar';
import { SelectedDetailsTabs } from './SelectedDetailsTabs';
import { useSelectedDetailsTabsContext } from './SelectedDetailsTabs/SelectedDetailsTabsContext';
import TagSelector from './TagSelector';

const Details: FC = () => {
    const { tagId: topTagId } = useHighestPrivilegeTagId();
    const { tagTypeDisplay, tagId: pathTagId, ruleId, memberId, objectDetailsLink } = usePZPathParams();
    const tagId = pathTagId ?? topTagId;

    const ruleQuery = useRuleInfo(tagId, ruleId);

    const { InfoHeader, EnvironmentSelector } = usePZContext();

    const objectCountsQuery = useObjectCounts();

    const { setSelectedDetailsTab } = useSelectedDetailsTabsContext();
    const navigate = useAppNavigate();

    if (!tagId)
        return (
            <Alert severity='error'>
                <AlertTitle>Missing Tag ID</AlertTitle>
                <p>We were unable to locate the Tag ID for loading this page. Please refresh the page and try again.</p>
            </Alert>
        );

    const handleObjectClick = (item: AssetGroupTagMemberListItem) => {
        setSelectedDetailsTab(ObjectTabValue);
        navigate(objectDetailsLink(tagId, item.id, ruleId));
    };

    return (
        <div className='h-full max-h-[75vh]'>
            <PageDescription />
            <div className='mt-6 w-2/3'>
                <InfoHeader />
            </div>
            <div className='flex gap-8 mt-4 h-full'>
                <div className='flex flex-col gap-2 basis-2/3 bg-neutral-2 pt-4 min-w-0 rounded shadow-outer-1 h-full'>
                    <h2 className='font-bold text-xl pl-4 pb-1'>{tagTypeDisplay} Details</h2>
                    <div className='flex flex-wrap justify-between w-full pb-4 border-b border-neutral-3 pl-4'>
                        <div className='flex gap-6 items-center'>
                            <TagSelector />
                            <EnvironmentSelector />
                        </div>
                        <SearchBar showTags={false} />
                    </div>
                    <div className='flex overflow-x-hidden max-lg:flex-col h-dvh'>
                        <div className='w-1/2 grow border-r border-neutral-3 max-lg:border-none max-lg:w-full overflow-y-auto'>
                            <RulesAccordion key={tagId} />
                        </div>
                        <div className='w-1/2 max-lg:w-full overflow-y-auto'>
                            {ruleQuery.data && ruleQuery.data.disabled_at !== null ? (
                                <Alert severity='warning' className='mx-8'>
                                    <AlertTitle>Rule is disabled</AlertTitle>
                                    <p>Enable this Rule to see Objects.</p>
                                </Alert>
                            ) : (
                                <ObjectsAccordion
                                    key={tagId + ruleId}
                                    kindCounts={objectCountsQuery.data?.counts || {}}
                                    totalCount={objectCountsQuery.data?.total_count || 0}
                                    tagId={tagId}
                                    ruleId={ruleId}
                                    objectId={memberId}
                                    onObjectClick={handleObjectClick}
                                />
                            )}
                        </div>
                    </div>
                </div>
                <SelectedDetailsTabs key={tagId} />
            </div>
        </div>
    );
};

export default Details;
