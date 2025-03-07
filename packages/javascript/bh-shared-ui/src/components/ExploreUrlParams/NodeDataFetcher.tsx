import { FC, ReactNode } from 'react';
import { useExploreParams, useSearch } from '../../hooks';
import { EntityKinds } from '../../utils';

export type SelectedNode = {
    id: string;
    type: EntityKinds;
    name: string;
    graphId?: string;
};

export const NodeDataFetcher: FC<{ children: (SelectedNode: SelectedNode | null) => ReactNode }> = ({ children }) => {
    const { panelSelection } = useExploreParams();
    const nodeQueryParam = panelSelection || '';
    const { data: nodeInfoResponse } = useSearch(nodeQueryParam, undefined);
    const nodeInfoObject = nodeInfoResponse?.length && nodeInfoResponse?.at(0);
    const selectedNode = nodeInfoObject
        ? {
              id: nodeInfoObject.objectid,
              name: nodeInfoObject.name,
              type: nodeInfoObject.type as EntityKinds,
              graphId: '',
          }
        : null;

    return children(selectedNode);
};
