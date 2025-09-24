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
    Tooltip,
} from '@bloodhoundenterprise/doodleui';
import { faTrashCan } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { FC, useContext, useState } from 'react';
import ExploreSearchCombobox from '../../../../components/ExploreSearchCombobox';
import NodeIcon from '../../../../components/NodeIcon';
import { SearchValue } from '../../../Explore';
import SelectorFormContext from './SelectorFormContext';

const ObjectSelect: FC = () => {
    const { selectedObjects, dispatch } = useContext(SelectorFormContext);
    const [searchTerm, setSearchTerm] = useState<string>('');

    const handleSelectedNode = (node: SearchValue) => {
        dispatch({ type: 'add-selected-object', node: node });
        setSearchTerm('');
    };

    const handleDeleteNode = (node: SearchValue) => {
        dispatch({ type: 'remove-selected-object', node: node });
    };

    return (
        <Card className='rounded-lg min-h-[36rem] h-[36rem]'>
            <CardHeader className='px-6 first:pt-6 text-xl font-bold'>
                <div className='flex justify-between'>
                    <span>Object Selector</span>
                </div>
                <CardDescription className='pt-3 font-normal'>
                    Use the input field to add objects to the list
                </CardDescription>
            </CardHeader>
            <CardContent className='px-6'>
                <div className='flex content-center mt-3'>
                    <div className='w-full my-2'>
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
                <div className='overflow-auto h-[400px]'>
                    <Table className='mt-5 w-full table-fixed' role='table' aria-label='Selected Objects'>
                        <TableBody className='first:border-t-[1px] last:border-b-[1px] border-neutral-light-5 dark:border-netural-dark-5'>
                            {selectedObjects.map((node, index) => (
                                <TableRow key={node.objectid + index} className='border-y p-0 *:p-0 *:h-12'>
                                    <TableCell className='*:p-0 text-center w-[30px]'>
                                        <Button
                                            variant={'text'}
                                            onClick={() => handleDeleteNode(node)}
                                            aria-label={`Remove ${node.name || node.objectid}`}>
                                            <FontAwesomeIcon icon={faTrashCan} />
                                        </Button>
                                    </TableCell>
                                    <TableCell className='text-center w-[60px]'>
                                        <NodeIcon nodeType={node.type || ''} />
                                    </TableCell>
                                    <TableCell className='mr-4 truncate'>
                                        <Tooltip
                                            tooltip={node.name || node.objectid}
                                            contentProps={{ className: 'max-w-80 dark:bg-neutral-dark-5 border-0' }}>
                                            <div className='truncate ml-2'>{node.name || node.objectid}</div>
                                        </Tooltip>
                                    </TableCell>
                                    {node.memberCount && (
                                        <TableCell className='text-center px-2 w-[116px]'>
                                            {node.memberCount} Members
                                        </TableCell>
                                    )}
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            </CardContent>
        </Card>
    );
};

export default ObjectSelect;
