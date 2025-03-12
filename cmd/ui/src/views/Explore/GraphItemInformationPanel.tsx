import { SxProps, useTheme } from '@mui/material';
import { EntityKinds, isEdge, isNode, useExploreSelectedItem } from 'bh-shared-ui';
import EdgeInfoPane from './EdgeInfo/EdgeInfoPane';
import EntityInfoPanel from './EntityInfo/EntityInfoPanel';

const GraphItemInformationPanel = () => {
    const { selectedItem, selectedItemQuery } = useExploreSelectedItem();

    const theme = useTheme();

    const infoPaneStyles: SxProps = {
        bottom: 0,
        top: 0,
        marginBottom: theme.spacing(2),
        marginTop: theme.spacing(2),
        maxWidth: theme.spacing(50),
        position: 'absolute',
        right: theme.spacing(2),
        width: theme.spacing(50),
    };

    if (!selectedItem || selectedItemQuery.isLoading) {
        return <EntityInfoPanel sx={infoPaneStyles} selectedNode={null} />;
    }

    // todo better error and loading state handling
    if (selectedItemQuery.isLoading) return null;

    if (selectedItemQuery.isError) {
        return (
            <EntityInfoPanel
                sx={infoPaneStyles}
                selectedNode={{ graphId: selectedItem, id: '', name: 'Unknown', type: 'Unknown' as EntityKinds }}
            />
        );
    }

    if (isEdge(selectedItemQuery.data!)) {
        const selectedEdge = {
            id: selectedItem as string,
            name: selectedItemQuery.data.label || '',
            data: selectedItemQuery.data.properties || {},
            sourceNode: {
                id: selectedItemQuery.data.source,
                objectId: selectedItemQuery.data.sourceNode.objectId,
                name: selectedItemQuery.data.sourceNode.label,
                type: selectedItemQuery.data.sourceNode.kind,
            },
            targetNode: {
                id: selectedItemQuery.data.target,
                objectId: selectedItemQuery.data.targetNode.objectId,
                name: selectedItemQuery.data.targetNode.label,
                type: selectedItemQuery.data.targetNode.kind,
            },
        };
        return <EdgeInfoPane sx={infoPaneStyles} selectedEdge={selectedEdge} />;
    }

    if (isNode(selectedItemQuery.data!)) {
        const selectedNode = {
            graphId: selectedItem,
            id: selectedItemQuery.data.objectId,
            name: selectedItemQuery.data.label,
            type: selectedItemQuery.data.kind as EntityKinds,
        };
        return <EntityInfoPanel sx={infoPaneStyles} selectedNode={selectedNode} />;
    }
};

export default GraphItemInformationPanel;
