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
import { usePZPathParams } from '../../../hooks';
import { SelectedDetailsTabContent } from './SelectedDetailsTabContent';
import { DetailsTabOption, detailsTabOptions } from './utils';

type SelectedDetailsTabsProps = {
    currentDetailsTab: DetailsTabOption;
    onTabClick: (value: DetailsTabOption) => void;
};

export const SelectedDetailsTabs: FC<SelectedDetailsTabsProps> = ({ currentDetailsTab, onTabClick }) => {
    const { memberId, ruleId, tagTypeDisplay, tagId } = usePZPathParams();

    return (
        <>
            <Tabs
                defaultValue={currentDetailsTab}
                value={currentDetailsTab}
                className='w-full mb-4'
                onValueChange={(value) => {
                    onTabClick(value as DetailsTabOption);
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
                    currentDetailsTab={currentDetailsTab}
                    tagId={tagId}
                    ruleId={ruleId}
                    memberId={memberId}
                />
            </Suspense>
        </>
    );
};
