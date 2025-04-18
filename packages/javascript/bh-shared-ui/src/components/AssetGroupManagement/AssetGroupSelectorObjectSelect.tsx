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
import { SeedTypeObjectId } from 'js-client-library';
import { SelectorSeedRequest } from 'js-client-library/dist/requests';
import { FC, useCallback, useState } from 'react';
import { SearchValue } from '../../store';
import ExploreSearchCombobox from '../ExploreSearchCombobox';
import NodeIcon from '../NodeIcon';

export type AssetGroupSelectedNode = SearchValue & { memberCount?: number };
export type AssetGroupSelectedNodes = AssetGroupSelectedNode[];

const mapSeeds = (seeds: SelectorSeedRequest[]): AssetGroupSelectedNodes => {
    return seeds.map((seed) => {
        return { objectid: seed.value };
    });
};

const AssetGroupSelectorObjectSelect: FC<{
    setSeeds: (seeds: SelectorSeedRequest[]) => void;
    seeds?: SelectorSeedRequest[];
}> = ({ setSeeds, seeds = [] }) => {
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [selectedNodes, setSelectedNodes] = useState<AssetGroupSelectedNodes>(mapSeeds(seeds));

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
        },
        [setSeeds]
    );

    return (
        <div>
            <Card className='rounded-lg'>
                <CardHeader className='px-6 first:pt-6 text-xl font-bold'>
                    Object Selector
                    <CardDescription className='pt-3 font-normal'>
                        Use the input field to add objects to the list
                    </CardDescription>
                </CardHeader>
                <CardContent className='pl-6'>
                    <div className='flex content-center mt-3'>
                        <div className='w-[18rem] mt-3'>
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
                            {selectedNodes.map((node) => (
                                <TableRow key={node.objectid} className='border-y-[1px] p-0 *:p-0 *:h-12'>
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
