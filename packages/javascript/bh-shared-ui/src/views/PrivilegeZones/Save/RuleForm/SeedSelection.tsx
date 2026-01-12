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
    Select,
    SelectContent,
    SelectItem,
    SelectPortal,
    SelectTrigger,
    SelectValue,
    Skeleton,
} from '@bloodhoundenterprise/doodleui';
import {
    SeedExpansionMethod,
    SeedExpansionMethodAll,
    SeedExpansionMethodChild,
    SeedExpansionMethodNone,
    SeedTypeCypher,
    SeedTypeObjectId,
    SeedTypesMap,
} from 'js-client-library';
import { FC, useCallback, useContext, useMemo, useState } from 'react';
import { Control } from 'react-hook-form';
import { useQuery } from 'react-query';
import { DeleteConfirmationDialog } from '../../../../components';
import VirtualizedNodeList from '../../../../components/VirtualizedNodeList';
import { encodeCypherQuery, useDeleteRule, useOwnedTagId, usePZPathParams } from '../../../../hooks';
import { useNotifications } from '../../../../providers';
import { detailsPath, privilegeZonesPath } from '../../../../routes';
import { apiClient, cn, useAppNavigate } from '../../../../utils';
import PrivilegeZonesCypherEditor from '../../PrivilegeZonesCypherEditor';
import { handleError } from '../utils';
import DeleteRuleButton from './DeleteRuleButton';
import ObjectSelect from './ObjectSelect';
import RuleFormContext from './RuleFormContext';
import { RuleFormInputs } from './types';

const getRuleExpansionMethod = (
    tagId: string,
    tagType: 'labels' | 'zones',
    ownedId: string | undefined
): SeedExpansionMethod => {
    // Owned is a specific tag type that does not undergo expansion
    if (tagId === ownedId) return SeedExpansionMethodNone;

    return tagType === 'zones' ? SeedExpansionMethodAll : SeedExpansionMethodChild;
};

const SeedSelection: FC<{ control: Control<RuleFormInputs, any, RuleFormInputs> }> = ({ control }) => {
    const { seeds, ruleType, ruleQuery, dispatch } = useContext(RuleFormContext);
    const [cypherQueryForExploreUrl, setCypherQueryForExploreUrl] = useState('');
    const navigate = useAppNavigate();
    const { addNotification } = useNotifications();
    const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
    const deleteRuleMutation = useDeleteRule();
    const { tagId, ruleId = '', tagType } = usePZPathParams();
    const ownedId = useOwnedTagId();
    const expansion = getRuleExpansionMethod(tagId, tagType, ownedId?.toString());

    const exploreUrl = useMemo(
        () =>
            `/ui/explore?searchType=cypher&exploreSearchTab=cypher&cypherSearch=${encodeURIComponent(encodeCypherQuery(cypherQueryForExploreUrl))}`,
        [cypherQueryForExploreUrl]
    );

    const previewQuery = useQuery({
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

    const handleDeleteRule = useCallback(async () => {
        try {
            if (!tagId || !ruleId) throw new Error(`Missing required entity IDs; tagId: ${tagId} , ruleId: ${ruleId}`);

            await deleteRuleMutation.mutateAsync({ tagId, ruleId });

            addNotification('Rule was deleted successfully!', undefined, {
                anchorOrigin: { vertical: 'top', horizontal: 'right' },
            });

            setDeleteDialogOpen(false);

            navigate(`/${privilegeZonesPath}/${tagType}/${tagId}/${detailsPath}`);
        } catch (error) {
            handleError(error, 'deleting', 'rule', addNotification);
        }
    }, [tagId, ruleId, navigate, deleteRuleMutation, addNotification, tagType]);
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
                        preview={false}
                        initialInput={firstSeed?.value}
                        onChange={setCypherQueryForExploreUrl}
                    />
                )}
            </div>
            <div>
                <Card className='xl:max-w-[26rem] sm:w-96 md:w-96 lg:w-lg grow max-lg:mb-10 2xl:max-w-full min-h-[36rem]'>
                    <CardHeader className='pl-6  text-xl font-bold'>
                        <div className='flex justify-between items-center'>
                            <span>Sample Results</span>
                            <Button
                                asChild
                                variant='text'
                                disabled={!cypherQueryForExploreUrl}
                                className={cn('font-normal', {
                                    'pointer-events-none hidden': !previewQuery.data,
                                })}>
                                <a
                                    href={cypherQueryForExploreUrl ? exploreUrl : undefined}
                                    target='_blank'
                                    rel='noreferrer'>
                                    View in Explore
                                </a>
                            </Button>
                        </div>
                    </CardHeader>
                    <p className='px-6 pb-3'>
                        Enter {ruleType === SeedTypeObjectId ? 'Object ID' : 'Cypher'} to see sample
                    </p>

                    <CardContent className='pl-4'>
                        <div className='font-bold pl-2 mb-2'>
                            <span>Type</span>
                            <span className='ml-8'>Object Name</span>
                        </div>
                        <VirtualizedNodeList nodes={previewQuery.data ?? []} itemSize={46} heightScalar={10} />
                    </CardContent>
                </Card>
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
                        Back
                    </Button>
                    <Button data-testid='privilege-zones_save_rule-form_save-button' variant={'primary'} type='submit'>
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
