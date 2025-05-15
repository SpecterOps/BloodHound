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
    CardDescription,
    CardHeader,
    Table,
    TableBody,
    TableCell,
    TableRow,
} from '@bloodhoundenterprise/doodleui';
import { faTrashCan } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { GraphNodes, SeedTypeObjectId, SelectorSeedRequest } from 'js-client-library';
import { FC, useCallback, useContext, useEffect, useState } from 'react';
import { useQuery } from 'react-query';
import { useParams } from 'react-router-dom';
import { SearchValue } from '../../store';
import { apiClient, cn } from '../../utils';
import SelectorFormContext from '../../views/TierManagement/Save/SelectorForm/SelectorFormContext';
import ExploreSearchCombobox from '../ExploreSearchCombobox';
import NodeIcon from '../NodeIcon';

export type AssetGroupSelectedNode = SearchValue & { memberCount?: number };
export type AssetGroupSelectedNodes = AssetGroupSelectedNode[];

const mapSeeds = (nodes: GraphNodes | undefined): AssetGroupSelectedNodes => {
    if (nodes === undefined) return [];
    return Object.values(nodes).map((node) => {
        return { objectid: node.objectId, name: node.label, type: node.kind };
    });
};

const AssetGroupSelectorObjectSelect: FC<{ seeds: SelectorSeedRequest[] }> = ({ seeds }) => {
    const { tagId = '', selectorId = '' } = useParams();

    const { setSeeds, setResults } = useContext(SelectorFormContext);

    const [searchTerm, setSearchTerm] = useState<string>('');
    const [stalePreview, setStalePreview] = useState(false);
    const [selectedNodes, setSelectedNodes] = useState<AssetGroupSelectedNodes>([]);

    const previewQuery = useQuery({
        queryKey: ['tier-management', 'preview-selectors', SeedTypeObjectId],
        queryFn: ({ signal }) => {
            if (selectedNodes.length === 0) return [];

            const seeds = selectedNodes.map((seed) => {
                return {
                    type: SeedTypeObjectId,
                    value: seed.objectid,
                };
            });

            return apiClient
                .assetGroupTagsPreviewSelectors({ seeds: [...seeds] }, { signal })
                .then((res) => res.data.data['members']);
        },
        retry: false,
        refetchOnWindowFocus: false,
    });

    const seedsQuery = useQuery({
        queryKey: ['tier-management', 'tags', tagId, 'selectors', selectorId, 'seeds'],
        queryFn: async ({ signal }) => {
            const seedsList = seeds.map((seed) => {
                return `"${seed.value}"`;
            });

            const query = `match(n) where n.objectid in [${seedsList?.join(',')}] return n`;

            return apiClient.cypherSearch(query, { signal }).then((res) => {
                setSelectedNodes(mapSeeds(res.data.data.nodes));
                return res.data.data;
            });
        },
        enabled: seeds.length !== 0,
        refetchOnWindowFocus: false,
    });

    const handleRun = useCallback(() => {
        previewQuery.refetch();
        setStalePreview(false);
    }, [previewQuery]);

    useEffect(() => {
        if (!previewQuery.data) return;

        setResults(previewQuery.data || []);
    }, [previewQuery.data, setResults]);

    useEffect(() => {
        previewQuery.refetch();
    }, [seedsQuery.data]);

    const handleSelectedNode = useCallback(
        (node: SearchValue) => {
            setSelectedNodes((prev) => {
                if (
                    prev.find((iteratedNode) => {
                        return iteratedNode.objectid === node.objectid;
                    })
                ) {
                    return prev;
                }

                const updatedNodes = [...prev, node];

                const seeds = updatedNodes.map((node) => {
                    return { type: SeedTypeObjectId, value: node.objectid };
                });

                setSeeds(seeds);

                return updatedNodes;
            });

            setSearchTerm('');
            setStalePreview(true);
        },
        [setSeeds]
    );

    const handleDeleteNode = useCallback(
        (node: SearchValue) => {
            setSelectedNodes((prev) => {
                const filteredNodes = prev.filter((n) => {
                    return n.objectid !== node.objectid;
                });

                const seeds = filteredNodes.map((node) => {
                    return { type: SeedTypeObjectId, value: node.objectid };
                });

                setSeeds(seeds);

                return filteredNodes;
            });
            setStalePreview(true);
        },
        [setSeeds]
    );

    return (
        <div>
            <Card className='rounded-lg'>
                <CardHeader className='px-6 first:pt-6 text-xl font-bold'>
                    <div className='flex justify-between'>
                        <span>Object Selector</span>
                        <Button
                            variant='text'
                            className={cn(
                                'p-0 text-sm text-primary font-bold dark:text-secondary-variant-2 hover:no-underline',
                                {
                                    'animate-pulse': stalePreview,
                                }
                            )}
                            onClick={handleRun}>
                            Update Sample Results
                        </Button>
                    </div>
                    <CardDescription className='pt-3 font-normal'>
                        Use the input field to add objects to the list
                    </CardDescription>
                </CardHeader>
                <CardContent className='pl-6'>
                    <div className='flex content-center mt-3'>
                        <div className='w-2xs mt-3'>
                            <ExploreSearchCombobox
                                labelText='Search Objects To Add'
                                inputValue={searchTerm}
                                selectedItem={null}
                                handleNodeEdited={setSearchTerm}
                                handleNodeSelected={handleSelectedNode}
                                variant='standard'
                            />
                        </div>
                    </div>
                    <Table className='mt-5 w-full table-fixed'>
                        <TableBody className='first:border-t-[1px] last:border-b-[1px] border-neutral-light-5 dark:border-netural-dark-5'>
                            {selectedNodes.map((node, index) => (
                                <TableRow key={node.objectid + index} className='border-y p-0 *:p-0 *:h-12'>
                                    <TableCell className='*:p-0 text-center w-[30px]'>
                                        <Button
                                            variant={'text'}
                                            onClick={() => handleDeleteNode(node)}
                                            aria-label='Remove object'>
                                            <FontAwesomeIcon icon={faTrashCan} />
                                        </Button>
                                    </TableCell>
                                    <TableCell className='text-center w-[84px]'>
                                        <NodeIcon nodeType={node.type || ''} />
                                    </TableCell>
                                    <TableCell className='mr-4 truncate'>{node.name || node.objectid}</TableCell>
                                    {node.memberCount && (
                                        <TableCell className='text-center px-2 w-[116px]'>
                                            {node.memberCount} Members
                                        </TableCell>
                                    )}
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </div>
    );
};

export default AssetGroupSelectorObjectSelect;
