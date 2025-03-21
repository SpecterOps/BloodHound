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
