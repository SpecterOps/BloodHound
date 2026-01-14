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
    Button,
    Card,
    CardContent,
    CardHeader,
    FormControl,
    FormField,
    FormItem,
    Input,
    Label,
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
    Skeleton,
} from '@bloodhoundenterprise/doodleui';
import { SeedTypeCypher, SeedTypeObjectId, SeedTypesMap } from 'js-client-library';
import { FC, useCallback, useContext, useMemo, useState } from 'react';
import { Control } from 'react-hook-form';
import { DeleteConfirmationDialog } from '../../../../components';
import { encodeCypherQuery, useDeleteRule, usePZPathParams } from '../../../../hooks';
import { useNotifications } from '../../../../providers';
import { cn, useAppNavigate } from '../../../../utils';
import PrivilegeZonesCypherEditor from '../../PrivilegeZonesCypherEditor';
import { handleError } from '../utils';
import DeleteRuleButton from './DeleteRuleButton';
import ObjectSelect from './ObjectSelect';
import RuleFormContext from './RuleFormContext';
import { SeedSelectionPreview } from './SeedSelectionPreview';
import { RuleFormInputs } from './types';

const SeedSelection: FC<{
    control: Control<RuleFormInputs, any, RuleFormInputs>;
}> = ({ control }) => {
    const [cypherQueryForExploreUrl, setCypherQueryForExploreUrl] = useState('');
    const { ruleId = '', tagId, tagDetailsLink } = usePZPathParams();
    const { seeds, ruleType, ruleQuery, dispatch } = useContext(RuleFormContext);
    const [stalePreview, setStalePreview] = useState(false);

    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const navigate = useAppNavigate();
    const deleteRuleMutation = useDeleteRule();
    const { addNotification } = useNotifications();
    const exploreUrl = useMemo(
        () =>
            cypherQueryForExploreUrl
                ? `/ui/explore?searchType=cypher&exploreSearchTab=cypher&cypherSearch=${encodeURIComponent(encodeCypherQuery(cypherQueryForExploreUrl))}`
                : undefined,
        [cypherQueryForExploreUrl]
    );

    const handleDeleteRule = useCallback(async () => {
        try {
            if (!tagId || !ruleId) throw new Error(`Missing required entity IDs; tagId: ${tagId} , ruleId: ${ruleId}`);

            await deleteRuleMutation.mutateAsync({ tagId, ruleId });

            addNotification('Rule was deleted successfully!', undefined, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });

            setDeleteDialogOpen(false);

            navigate(tagDetailsLink(tagId));
        } catch (error) {
            handleError(error, 'deleting', 'rule', addNotification);
        }
    }, [tagId, ruleId, navigate, deleteRuleMutation, addNotification, tagDetailsLink]);
    const handleCancel = useCallback(() => setDeleteDialogOpen(false), []);

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
                    <CardHeader className='text-xl font-bold'>
                        <Label className='text-base font-bold' htmlFor='rule-seed-type-select'>
                            Rule Type
                        </Label>
                    </CardHeader>
                    <CardContent>
                        <Select
                            data-testid='privilege-zones_save_rule-form_type-select'
                            value={ruleType.toString()}
                            onValueChange={(value: string) => {
                                if (value === SeedTypeObjectId.toString()) {
                                    dispatch({ type: 'set-rule-type', ruleType: SeedTypeObjectId });
                                } else if (value === SeedTypeCypher.toString()) {
                                    dispatch({ type: 'set-rule-type', ruleType: SeedTypeCypher });
                                    dispatch({ type: 'set-seeds', seeds: [] });

                                    setStalePreview(true);
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
                        stalePreview={stalePreview}
                        setStalePreview={setStalePreview}
                    />
                )}
            </div>
            <div>
                <SeedSelectionPreview exploreUrl={exploreUrl} seeds={seeds} ruleType={ruleType} />
                <div className='flex justify-end gap-2 mt-6'>
                    <DeleteRuleButton
                        ruleId={ruleId}
                        ruleData={ruleQuery.data}
                        onClick={() => {
                            setDeleteDialogOpen(true);
                        }}
                    />
                    <Button
                        data-testid='privilege-zones_save_rule-form_cancel-button'
                        variant={'secondary'}
                        onClick={() => navigate(-1)}>
                        Cancel
                    </Button>
                    <Button data-testid='privilege-zones_save_rule-form_save-button' variant='secondary' type='submit'>
                        {ruleId === '' ? 'Create Rule' : 'Save Edits'}
                    </Button>
                </div>
            </div>
            <DeleteConfirmationDialog
                open={deleteDialogOpen}
                itemName={ruleQuery.data?.name || 'Rule'}
                itemType='rule'
                onConfirm={handleDeleteRule}
                onCancel={handleCancel}
            />
        </>
    );
};

export default SeedSelection;
