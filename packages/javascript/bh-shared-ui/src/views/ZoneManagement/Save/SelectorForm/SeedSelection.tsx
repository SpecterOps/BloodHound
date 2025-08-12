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

import {
    Card,
    CardContent,
    CardHeader,
    FormControl,
    FormField,
    FormItem,
    Input,
    Skeleton,
} from '@bloodhoundenterprise/doodleui';
import { SeedTypeObjectId } from 'js-client-library';
import { FC, useContext } from 'react';
import { Control } from 'react-hook-form';
import { useQuery } from 'react-query';
import VirtualizedNodeList from '../../../../components/VirtualizedNodeList';
import { apiClient, cn } from '../../../../utils';
import { Cypher } from '../../Cypher/Cypher';
import ObjectSelect from './ObjectSelect';
import SelectorFormContext from './SelectorFormContext';
import { SelectorFormInputs } from './types';

const SeedSelection: FC<{ control: Control<SelectorFormInputs, any, SelectorFormInputs> }> = ({ control }) => {
    const { seeds, selectorType, selectorQuery } = useContext(SelectorFormContext);

    const previewQuery = useQuery({
        queryKey: ['zone-management', 'preview-selectors', selectorType, seeds],
        queryFn: async ({ signal }) => {
            return apiClient
                .assetGroupTagsPreviewSelectors({ seeds: seeds }, { signal })
                .then((res) => res.data.data['members']);
        },
        retry: false,
        refetchOnWindowFocus: false,
        enabled: seeds.length > 0,
    });

    if (selectorQuery.isLoading) return <Skeleton />;
    if (selectorQuery.isError) return <div>There was an error fetching the selector data</div>;

    const firstSeed = seeds.values().next().value;

    return (
        <>
            <div
                className={cn('w-full grow h-[36rem] md:w-96 xl:max-w-[36rem] 2xl:max-w-full', {
                    'md:w-60': selectorType === SeedTypeObjectId,
                })}>
                <FormField
                    control={control}
                    name='seeds'
                    render={({ field }) => (
                        <FormItem>
                            <FormControl>
                                <Input
                                    {...field}
                                    value={Array.from(seeds) as unknown as string}
                                    className='hidden w-0'
                                />
                            </FormControl>
                        </FormItem>
                    )}
                />
                {selectorType === SeedTypeObjectId ? (
                    <ObjectSelect />
                ) : (
                    <Cypher preview={false} initialInput={firstSeed?.value} />
                )}
            </div>
            <Card className='xl:max-w-[26rem] sm:w-96 md:w-96 lg:w-lg grow max-lg:mb-10 2xl:max-w-full min-h-[36rem]'>
                <CardHeader className='pl-6 first:py-6 text-xl font-bold'>Sample Results</CardHeader>
                <CardContent className='pl-4'>
                    <div className='font-bold pl-2 mb-2'>
                        <span>Type</span>
                        <span className='ml-8'>Object Name</span>
                    </div>
                    <VirtualizedNodeList nodes={previewQuery.data ?? []} itemSize={46} heightScalar={10} />
                </CardContent>
            </Card>
        </>
    );
};

export default SeedSelection;
