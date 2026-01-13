// Copyright 2026 Specter Ops, Inc.
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

import { Button, Card, CardContent, CardHeader } from '@bloodhoundenterprise/doodleui';
import {
    NodeSourceSeed,
    SeedExpansionMethod,
    SeedExpansionMethodAll,
    SeedExpansionMethodChild,
    SeedExpansionMethodNone,
    SeedTypes,
    SeedTypesMap,
    SelectorSeedRequest,
} from 'js-client-library';
import { FC } from 'react';
import { useQuery } from 'react-query';
import VirtualizedNodeList from '../../../../components/VirtualizedNodeList';
import { useOwnedTagId } from '../../../../hooks/useAssetGroupTags';
import { usePZPathParams } from '../../../../hooks/usePZParams';
import { apiClient, cn } from '../../../../utils';

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

export const SeedSelectionPreview: FC<{ seeds: SelectorSeedRequest[]; ruleType: SeedTypes; exploreUrl?: string }> = ({
    seeds,
    ruleType,
    exploreUrl,
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

    const directObjects = sampleResults?.filter((objectItem) => objectItem.source === NodeSourceSeed);
    const expandedObjects = sampleResults?.filter((objectItem) => objectItem.source > NodeSourceSeed);
    const showViewInExploreButton = Boolean(directObjects?.length || expandedObjects?.length);

    return (
        <Card className='xl:max-w-[26rem] sm:w-96 md:w-96 lg:w-lg grow max-lg:mb-10 2xl:max-w-full min-h-[36rem]'>
            <CardHeader className='pl-6 first:py-6 text-xl font-bold'>
                <div className='flex justify-between items-center'>
                    <span>Sample Results</span>
                    <Button
                        asChild
                        variant='text'
                        disabled={!exploreUrl}
                        className={cn('font-normal', {
                            'pointer-events-none hidden': !showViewInExploreButton,
                        })}>
                        <a href={exploreUrl ? exploreUrl : undefined} target='_blank' rel='noreferrer'>
                            View in Explore
                        </a>
                    </Button>
                </div>
            </CardHeader>
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
                    displayText={`Enter ${SeedTypesMap[ruleType]} to see sample results`}
                />
            )}
        </Card>
    );
};

// const { seeds, ruleType, ruleQuery, dispatch } = useContext(RuleFormContext);
// const [cypherQueryForExploreUrl, setCypherQueryForExploreUrl] = useState('');
// const navigate = useAppNavigate();
// const { addNotification } = useNotifications();
// const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
// const deleteRuleMutation = useDeleteRule();
// const { tagId, ruleId = '', tagType } = usePZPathParams();
// const ownedId = useOwnedTagId();
// const expansion = getRuleExpansionMethod(tagId, tagType, ownedId?.toString());

// const exploreUrl = useMemo(
//     () =>
//         `/ui/explore?searchType=cypher&exploreSearchTab=cypher&cypherSearch=${encodeURIComponent(encodeCypherQuery(cypherQueryForExploreUrl))}`,
//     [cypherQueryForExploreUrl]
// );

// const previewQuery = useQuery({
//     queryKey: ['privilege-zones', 'preview-selectors', ruleType, seeds, expansion],
//     queryFn: async ({ signal }) => {
//         return apiClient
//             .assetGroupTagsPreviewSelectors({ seeds, expansion }, { signal })
//             .then((res) => res.data.data['members']);
//     },
//     retry: false,
//     refetchOnWindowFocus: false,
//     enabled: seeds.length > 0,
// });
// const { seeds, ruleType, ruleQuery } = useContext(RuleFormContext);

// const handleDeleteRule = useCallback(async () => {
//     try {
//         if (!tagId || !ruleId) throw new Error(`Missing required entity IDs; tagId: ${tagId} , ruleId: ${ruleId}`);

//         await deleteRuleMutation.mutateAsync({ tagId, ruleId });

//         addNotification('Rule was deleted successfully!', undefined, {
//             anchorOrigin: { vertical: 'top', horizontal: 'right' },
//         });

//         setDeleteDialogOpen(false);

//         navigate(`/${privilegeZonesPath}/${tagType}/${tagId}/${detailsPath}`);
//     } catch (error) {
//         handleError(error, 'deleting', 'rule', addNotification);
//     }
// }, [tagId, ruleId, navigate, deleteRuleMutation, addNotification, tagType]);
// const handleCancel = useCallback(() => setDeleteDialogOpen(false), []);

// <div>
//     <Card className='xl:max-w-[26rem] sm:w-96 md:w-96 lg:w-lg grow max-lg:mb-10 2xl:max-w-full min-h-[36rem]'>
//         <CardHeader className='pl-6  text-xl font-bold'>
//             <div className='flex justify-between items-center'>
//                 <span>Sample Results</span>
//                 <Button
//                     asChild
//                     variant='text'
//                     disabled={!cypherQueryForExploreUrl}
//                     className={cn('font-normal', {
//                         'pointer-events-none hidden': !previewQuery.data,
//                     })}>
//                     <a
//                         href={cypherQueryForExploreUrl ? exploreUrl : undefined}
//                         target='_blank'
//                         rel='noreferrer'>
//                         View in Explore
//                     </a>
//                 </Button>
//             </div>
//         </CardHeader>
//         <p className='px-6 pb-3'>
//             Enter {ruleType === SeedTypeObjectId ? 'Object ID' : 'Cypher'} to see sample
//         </p>

//         <CardContent className='pl-4'>
//             <div className='font-bold pl-2 mb-2'>
//                 <span>Type</span>
//                 <span className='ml-8'>Object Name</span>
//             </div>
//             <VirtualizedNodeList nodes={previewQuery.data ?? []} itemSize={46} heightScalar={10} />
//         </CardContent>
//     </Card>
//     <div className='flex justify-end gap-2 mt-6'>
//         <DeleteRuleButton
//             ruleId={ruleId}
//             ruleData={ruleQuery.data}
//             onClick={() => {
//                 setDeleteDialogOpen(true);
//             }}
//         />
//         <Button
//             data-testid='privilege-zones_save_rule-form_cancel-button'
//             variant={'secondary'}
//             onClick={() => navigate(-1)}>
//             Back
//         </Button>
//         <Button data-testid='privilege-zones_save_rule-form_save-button' variant={'primary'} type='submit'>
//             {ruleId === '' ? 'Create Rule' : 'Save Edits'}
//         </Button>
//     </div>
// </div>
// <DeleteConfirmationDialog
//     open={deleteDialogOpen}
//     itemName={ruleQuery.data?.name || 'Rule'}
//     itemType='rule'
//     onConfirm={handleDeleteRule}
//     onCancel={handleCancel}
// />
