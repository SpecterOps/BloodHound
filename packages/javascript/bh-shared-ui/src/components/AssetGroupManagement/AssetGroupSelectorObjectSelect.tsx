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
import { Box } from '@mui/material';
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
        <Box maxWidth='sm'>
            <Card className='mt-5'>
                <CardHeader>
                    <CardTitle className='text-md'>Object Selector </CardTitle>
                </CardHeader>
                <CardContent>
                    <Box maxWidth='30%'>
                        <ExploreSearchCombobox
                            labelText='Input Field'
                            inputValue={searchTerm}
                            selectedItem={null}
                            handleNodeEdited={setSearchTerm}
                            handleNodeSelected={handleSelectedNode}
                        />
                    </Box>
                    <Table className='mt-5'>
                        <TableBody>
                            {selectedNodes.map((node) => (
                                <TableRow key={node.objectid} className='border-t-1'>
                                    <TableCell>
                                        <NodeIcon nodeType={node.type || ''} />
                                    </TableCell>
                                    <TableCell>{node.name || node.objectid}</TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </CardContent>
            </Card>
        </Box>
    );
};

export default AssetGroupSelectorObjectSelect;
