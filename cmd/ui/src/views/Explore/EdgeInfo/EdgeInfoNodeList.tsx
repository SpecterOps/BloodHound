import { InfiniteScrollingTable, apiClient } from 'bh-shared-ui';

export interface EdgeInfoNodeListProps {
    sourceId: number;
    targetId: number;
    edgeName: string;
}

const EdgeInfoNodeList: React.FC<EdgeInfoNodeListProps> = ({ sourceId, targetId, edgeName }) => {
    return (
        <InfiniteScrollingTable
            fetchDataCallback={({ skip, limit }) => {
                return apiClient.getEdgeDetails(sourceId, targetId, edgeName).then((result) => ({
                    data: Object.entries(result.data.data.nodes),
                    total: Object.keys(result.data.data.nodes).length,
                    skip,
                    limit,
                }));
            }}
        />
    );
};

export default EdgeInfoNodeList;
