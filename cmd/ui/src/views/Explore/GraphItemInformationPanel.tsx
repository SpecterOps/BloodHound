import { EntityKinds, isEdge, isNode, useExploreSelectedItem } from 'bh-shared-ui';
import EdgeInfoPane from './EdgeInfo/EdgeInfoPane';
import EntityInfoPanel from './EntityInfo/EntityInfoPanel';

const GraphItemInformationPanel = () => {
    const { selectedItem, selectedItemQuery } = useExploreSelectedItem();

    if (!selectedItem) {
        return null;
    }

    // todo better error and loading state handling
    if (selectedItemQuery.isLoading) return null;

    if (selectedItemQuery.isError) return null;

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
        return <EdgeInfoPane selectedEdge={selectedEdge} />;
    }

    if (isNode(selectedItemQuery.data!)) {
        const selectedNode = {
            graphId: selectedItem,
            id: selectedItemQuery.data.objectId,
            name: selectedItemQuery.data.label,
            type: selectedItemQuery.data.kind as EntityKinds,
        };
        return <EntityInfoPanel selectedNode={selectedNode} />;
    }
};

export default GraphItemInformationPanel;
