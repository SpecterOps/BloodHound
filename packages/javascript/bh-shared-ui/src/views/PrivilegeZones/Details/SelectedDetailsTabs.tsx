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
import { FC, Suspense, useState } from 'react';
import { usePZPathParams } from '../../../hooks';
import { SelectedDetailsTabContent } from './SelectedDetailsTabContent';
import { DetailsTabOption, detailsTabOptions } from './utils';

const selectedDetailsTabFromPathParams = (memberId?: string, ruleId?: string) => {
    if (memberId) return detailsTabOptions[2];
    if (ruleId && !memberId) return detailsTabOptions[1];
    return detailsTabOptions[0];
};

export const SelectedDetailsTabs: FC = () => {
    const { memberId, ruleId, tagTypeDisplay, tagId } = usePZPathParams();

    const [clickedTab, setClickedTab] = useState<DetailsTabOption>();

    const listChosenTab = selectedDetailsTabFromPathParams(memberId, ruleId);
    const currentSelectedTab = clickedTab || listChosenTab;

    console.log(tagTypeDisplay);

    return (
        <>
            <Tabs
                defaultValue={detailsTabOptions[0]}
                value={currentSelectedTab}
                className='w-full mb-4'
                onValueChange={(value) => {
                    setClickedTab(value as DetailsTabOption);
                }}>
                <TabsList className='w-full flex justify-start'>
                    <TabsTrigger value={detailsTabOptions[0]}>{tagTypeDisplay}</TabsTrigger>
                    <TabsTrigger disabled={!ruleId} value={detailsTabOptions[1]}>
                        Rule
                    </TabsTrigger>
                    <TabsTrigger disabled={!memberId} value={detailsTabOptions[2]}>
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
                    currentDetailsTab={currentSelectedTab}
                    tagId={tagId}
                    ruleId={ruleId}
                    memberId={memberId}
                />
            </Suspense>
        </>
    );
};
