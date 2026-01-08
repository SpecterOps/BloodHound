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

import { Card, CardContent, CardHeader } from '@bloodhoundenterprise/doodleui';
import {
    SeedExpansionMethod,
    SeedExpansionMethodAll,
    SeedExpansionMethodChild,
    SeedExpansionMethodNone,
    SeedTypes,
    SelectorSeedRequest,
} from 'js-client-library';
import { FC, useMemo } from 'react';
import { useQuery } from 'react-query';
import VirtualizedNodeList from '../../../../components/VirtualizedNodeList';
import { useOwnedTagId } from '../../../../hooks'; //specify file
import { usePZPathParams } from '../../../../hooks/usePZParams';
import { apiClient } from '../../../../utils';

const getRuleExpansionMethod = (
    tagId: string,
    tagType: 'labels' | 'zones',
    ownedId: string | undefined
): SeedExpansionMethod => {
    // Owned is a specific tag type that does not undergo expansion
    if (tagId === ownedId) return SeedExpansionMethodNone;

    return tagType === 'zones' ? SeedExpansionMethodAll : SeedExpansionMethodChild;
};

const EmptySeedResults: FC<{ className: string; displayText: string }> = ({ className, displayText }) => {
    return <p className={className}>{displayText}</p>;
};

export const SeedSelectionPreview: FC<{ seeds: SelectorSeedRequest[]; ruleType: SeedTypes }> = ({
    seeds,
    ruleType,
}) => {
    const { tagType, tagId } = usePZPathParams();
    const ownedId = useOwnedTagId();

    const expansion = getRuleExpansionMethod(tagId, tagType, ownedId?.toString());

    const { data: sampleResults, isFetched: sampleResultsFetched } = useQuery({
        queryKey: ['privilege-zones', 'preview-selectors', ruleType, seeds, expansion],
        queryFn: async ({ signal }) => {
            return apiClient
                .assetGroupTagsPreviewSelectors({ seeds, expansion }, { signal })
                .then((res) => res.data.data['members']);
        },
        retry: false,
        refetchOnWindowFocus: false,
        enabled: seeds.length > 0,
    });

    const directObjects = useMemo(
        () => sampleResults?.filter((objectItem) => objectItem.source === 1),
        [sampleResults]
    );
    const expandedObjects = useMemo(
        () => sampleResults?.filter((objectItem) => objectItem.source > 1),
        [sampleResults]
    );

    const setRuleTypeDisplay = () => {
        switch (ruleType) {
            case 1:
                return 'Object ID';
            case 2:
                return 'Cypher';
            default:
                return '';
        }
    };
    return (
        <Card className='xl:max-w-[26rem] sm:w-96 md:w-96 lg:w-lg grow max-lg:mb-10 2xl:max-w-full min-h-[36rem]'>
            <CardHeader className='pl-6 first:py-6 text-xl font-bold'>Sample Results</CardHeader>
            {sampleResultsFetched ? (
                <>
                    <CardContent data-testid='pz-rule-preview__direct-objects-list' className='pl-4 '>
                        <div className='font-bold pl-2 mb-2'>Direct Objects</div>
                        {directObjects?.length ? (
                            <VirtualizedNodeList nodes={directObjects} itemSize={46} heightScalar={5} />
                        ) : (
                            <EmptySeedResults className='pl-2' displayText='No results found' />
                        )}
                    </CardContent>
                    <CardContent data-testid='pz-rule-preview__expanded-objects-list' className='pl-4 '>
                        <div className='font-bold pl-2 mb-2'>Expanded Objects</div>
                        {expandedObjects?.length ? (
                            <VirtualizedNodeList nodes={expandedObjects} itemSize={46} heightScalar={7} />
                        ) : (
                            <EmptySeedResults className='pl-2' displayText='No results found' />
                        )}
                    </CardContent>
                </>
            ) : (
                <EmptySeedResults
                    className='pl-6'
                    displayText={`Enter ${setRuleTypeDisplay()} to see sample results`}
                />
            )}
        </Card>
    );
};
