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

import { FormControl, FormField, FormItem, Input, Skeleton } from '@bloodhoundenterprise/doodleui';
import { SeedTypeObjectId } from 'js-client-library';
import { FC, useContext } from 'react';
import { Control } from 'react-hook-form';
import { cn } from '../../../../utils';
import { Cypher } from '../../Cypher/Cypher';
import ObjectSelect from './ObjectSelect';
import RuleFormContext from './RuleFormContext';
import { SeedSelectionPreview } from './SeedSelectionPreview';
import { RuleFormInputs } from './types';

const SeedSelection: FC<{ control: Control<RuleFormInputs, any, RuleFormInputs> }> = ({ control }) => {
    const { seeds, ruleType, ruleQuery } = useContext(RuleFormContext);

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
                {ruleType === SeedTypeObjectId ? (
                    <ObjectSelect />
                ) : (
                    <Cypher preview={false} initialInput={firstSeed?.value} />
                )}
            </div>
            <SeedSelectionPreview seeds={seeds} ruleType={ruleType} />
        </>
    );
};

export default SeedSelection;
