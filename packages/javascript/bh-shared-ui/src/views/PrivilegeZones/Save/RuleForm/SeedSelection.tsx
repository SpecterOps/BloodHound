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
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
    Skeleton,
} from '@bloodhoundenterprise/doodleui';
import { SeedTypeCypher, SeedTypeObjectId, SeedTypesMap } from 'js-client-library';
import { FC, useContext, useMemo, useState } from 'react';
import { Control } from 'react-hook-form';
import { encodeCypherQuery } from '../../../../hooks';
import { cn } from '../../../../utils';
import PrivilegeZonesCypherEditor from '../../PrivilegeZonesCypherEditor';
import ObjectSelect from './ObjectSelect';
import RuleFormContext from './RuleFormContext';
import { SeedSelectionPreview } from './SeedSelectionPreview';
import { RuleFormInputs } from './types';

const SeedSelection: FC<{ control: Control<RuleFormInputs, any, RuleFormInputs> }> = ({ control }) => {
    const [cypherQueryForExploreUrl, setCypherQueryForExploreUrl] = useState('');
    const { seeds, ruleType, ruleQuery, dispatch } = useContext(RuleFormContext);

    const exploreUrl = useMemo(
        () =>
            cypherQueryForExploreUrl
                ? `/ui/explore?searchType=cypher&exploreSearchTab=cypher&cypherSearch=${encodeURIComponent(encodeCypherQuery(cypherQueryForExploreUrl))}`
                : undefined,
        [cypherQueryForExploreUrl]
    );

    if (ruleQuery.isLoading) return <Skeleton />;
    if (ruleQuery.isError) return <div>There was an error fetching the rule data</div>;

    const firstSeed = seeds.values().next().value;

    return (
        <>
            <div
                className={cn('w-full grow h-[36rem] md:w-96 xl:max-w-[36rem] 2xl:max-w-full', {
                    'md:w-60': ruleType === SeedTypeObjectId,
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
                <Card className='mb-5 pl-4 px-4 py-2'>
                    <CardHeader className='text-xl font-bold'>Rule Type</CardHeader>
                    <CardContent>
                        <Select
                            data-testid='privilege-zones_save_rule-form_type-select'
                            value={ruleType.toString()}
                            onValueChange={(value: string) => {
                                if (value === SeedTypeObjectId.toString()) {
                                    dispatch({ type: 'set-rule-type', ruleType: SeedTypeObjectId });
                                } else if (value === SeedTypeCypher.toString()) {
                                    dispatch({ type: 'set-rule-type', ruleType: SeedTypeCypher });
                                }
                            }}>
                            <SelectTrigger aria-label='select rule seed type' id='rule-seed-type-select'>
                                <SelectValue placeholder='Choose a Rule Type' />
                            </SelectTrigger>
                            <SelectPortal>
                                <SelectContent>
                                    {Object.entries(SeedTypesMap).map(([seedType, displayValue]) => (
                                        <SelectItem key={seedType} value={seedType}>
                                            {displayValue}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </SelectPortal>
                        </Select>
                    </CardContent>
                </Card>
                {ruleType === SeedTypeObjectId ? (
                    <ObjectSelect />
                ) : (
                    <PrivilegeZonesCypherEditor
                        onChange={setCypherQueryForExploreUrl}
                        preview={false}
                        initialInput={firstSeed?.value}
                    />
                )}
            </div>
            <SeedSelectionPreview exploreUrl={exploreUrl} seeds={seeds} ruleType={ruleType} />
        </>
    );
};

export default SeedSelection;
