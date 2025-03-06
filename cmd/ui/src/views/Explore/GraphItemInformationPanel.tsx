import { isEdge, useExploreSelectedItem } from 'bh-shared-ui';
import EdgeInfoPane from './EdgeInfo/EdgeInfoPane';
// import EntityInfoPanel from './EntityInfo/EntityInfoPanel';

const GraphItemInformationPanel = () => {
    const { selectedItem, selectedItemQuery } = useExploreSelectedItem();

    if (!selectedItem) {
        return null;
    }

    // todo better error and loading state handling
    if (selectedItemQuery.isLoading) return null;

    if (selectedItemQuery.isError) return null;

    return isEdge(selectedItemQuery.data!) ? (
        <EdgeInfoPane selectedEdge={selectedItemQuery.data} />
    ) : // : <EntityInfoPanel />;
    undefined;
};

export default GraphItemInformationPanel;
