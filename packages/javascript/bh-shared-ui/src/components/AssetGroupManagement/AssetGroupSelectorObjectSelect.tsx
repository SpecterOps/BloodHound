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
    CardTitle,
    Table,
    TableBody,
    TableCell,
    TableRow,
} from '@bloodhoundenterprise/doodleui';
import { FC, useCallback, useState } from 'react';
import { SearchValue } from '../../store';
import ExploreSearchCombobox from '../ExploreSearchCombobox';
import NodeIcon from '../NodeIcon';

const AssetGroupSelectorObjectSelect: FC<{
    selectedNodes: SearchValue[];
    onSelectNode: (node: SearchValue) => void;
}> = ({ selectedNodes, onSelectNode }) => {
    const [searchTerm, setSearchTerm] = useState<string>('');

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
                    <div className='w-[12rem]'>
                        <ExploreSearchCombobox
                            labelText='Input Field'
                            inputValue={searchTerm}
                            selectedItem={null}
                            handleNodeEdited={setSearchTerm}
                            handleNodeSelected={handleSelectedNode}
                            variant='standard'
                        />
                    </div>
                    <Table className='mt-5 w-full table-fixed'>
                        <TableBody className='last:border-b-[1px] border-neutral-light-5 dark:border-netural-dark-5'>
                            {selectedNodes.map((node) => (
                                <TableRow
                                    key={node.objectid}
                                    className='border-y-[1px] border-neutral-light-5 dark:border-netural-dark-5 p-0 *:p-0 *:h-12'>
                                    <TableCell className='text-center w-[84px]'>
                                        <NodeIcon nodeType={node.type || ''} />
                                    </TableCell>
                                    <TableCell className='mr-4 truncate'>{node.name || node.objectid}</TableCell>
                                    {/* TODO add member count  */}
                                    {/* <TableCell className='text-center px-2 w-[116px]'>777 Members</TableCell> */}
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
