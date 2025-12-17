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
import { FC, Suspense } from 'react';
import { usePZPathParams } from '../../../../hooks';
import { DetailsTabOption, TabObjectOption, TabRuleOption, TabTagOption } from '../utils';
import { SelectedDetailsTabContent } from './SelectedDetailsTabContent';
import { useSelectedDetailsTabsContext } from './SelectedDetailsTabsContext';

export const SelectedDetailsTabs: FC = () => {
    const { memberId, ruleId, tagTypeDisplay, tagId } = usePZPathParams();
    const { selectedDetailsTab, setSelectedDetailsTab } = useSelectedDetailsTabsContext();

    return (
        <>
            <Tabs
                value={selectedDetailsTab}
                className='w-full pb-4'
                onValueChange={(value) => {
                    setSelectedDetailsTab(value as DetailsTabOption);
                }}>
                <TabsList className='w-full flex justify-start'>
                    <TabsTrigger value={TabTagOption}>{tagTypeDisplay}</TabsTrigger>
                    <TabsTrigger disabled={!ruleId} value={TabRuleOption}>
                        Rule
                    </TabsTrigger>
                    <TabsTrigger disabled={!memberId} value={TabObjectOption}>
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
                <SelectedDetailsTabContent
                    currentDetailsTab={selectedDetailsTab}
                    tagId={tagId}
                    ruleId={ruleId}
                    memberId={memberId}
                />
            </Suspense>
        </>
    );
};
