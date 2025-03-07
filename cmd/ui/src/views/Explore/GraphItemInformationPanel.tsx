import { isEdge, isNode, useExploreSelectedItem } from 'bh-shared-ui';
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

    if (isEdge(selectedItemQuery.data!)) return <EdgeInfoPane selectedEdge={selectedItemQuery.data} />;

    if (isNode(selectedItemQuery.data!)) return <EntityInfoPanel selectedNode={selectedItemQuery.data} />;
};

export default GraphItemInformationPanel;
