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
    CardTitle,
    Table,
    TableBody,
    TableCell,
    TableRow,
} from '@bloodhoundenterprise/doodleui';
import { faPencil, faTrashCan } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FC, useCallback, useState } from 'react';
import { SearchValue } from '../../store';
import ExploreSearchCombobox from '../ExploreSearchCombobox';
import NodeIcon from '../NodeIcon';

const AssetGroupSelectorObjectSelect: FC<{
    selectedNodes: (SearchValue & { memberCount?: number })[];
    onSelectNode: (node: SearchValue & { memberCount?: number }) => void;
    onDeleteNode: (nodeObjectId: string) => void;
}> = ({ selectedNodes, onSelectNode, onDeleteNode }) => {
    const [searchTerm, setSearchTerm] = useState<string>('');
    const [showDeleteIcons, setShowDeleteIcons] = useState<boolean>(false);

    const handleSelectedNode = useCallback(
        (node: SearchValue) => {
            onSelectNode(node);
            setSearchTerm('');
        },
        [onSelectNode]
    );

    return (
        <div className='max-w-2xl'>
            <Card className='mt-5'>
                <CardHeader>
                    <CardTitle className='text-md'>Object Selector </CardTitle>
                </CardHeader>
                <CardContent>
                    <p className='text-sm'>
                        Use the input field to add objects and the edit button to remove objects from the list
                    </p>
                    <div className='flex content-center'>
                        <div className='w-[12rem] mt-3'>
                            <ExploreSearchCombobox
                                labelText='Search Objects To Add'
                                inputValue={searchTerm}
                                selectedItem={null}
                                handleNodeEdited={setSearchTerm}
                                handleNodeSelected={handleSelectedNode}
                                variant='standard'
                            />
                        </div>
                        <Button
                            className='rounded-full ml-5 mt-1'
                            variant={'icon'}
                            onClick={() => setShowDeleteIcons((prev) => !prev)}
                            aria-label='Edit selected objects'>
                            <FontAwesomeIcon icon={faPencil} size='lg' />
                        </Button>
                    </div>
                    <Table className='mt-5 w-full table-fixed'>
                        <TableBody className='first:border-t-[1px] last:border-b-[1px] border-neutral-light-5 dark:border-netural-dark-5'>
                            {selectedNodes.map((node) => (
                                <TableRow key={node.objectid} className='border-y-[1px] p-0 *:p-0 *:h-12'>
                                    {showDeleteIcons && (
                                        <TableCell className='*:p-0 text-center w-[30px]'>
                                            <Button
                                                variant={'text'}
                                                onClick={() => onDeleteNode(node.objectid)}
                                                aria-label='Remove object'>
                                                <FontAwesomeIcon icon={faTrashCan} />
                                            </Button>
                                        </TableCell>
                                    )}
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
