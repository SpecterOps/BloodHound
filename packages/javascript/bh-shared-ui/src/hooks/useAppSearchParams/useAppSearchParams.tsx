// Copyright 2023 Specter Ops, Inc.
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

import { useSearchParams } from 'react-router-dom';
import { ActiveDirectoryNodeKind, AzureNodeKind } from '../../graphSchema';

type UrlParamState = {
    graphQueryType: 'primary' | 'secondary' | 'cypher'; // TODO: there are types for this and we should use them
    primaryQuery: string; // search and pathfinding from
    secondaryQuery: string; // pathfinding to
    graphLayout: 'sequential' | 'standard';
    selectedNode: string; // TODO: should be node objectId
    selectedNodeType: AzureNodeKind | ActiveDirectoryNodeKind;
    omittedEdges: string[];
};

type UrlParamStateKeys = keyof UrlParamState;

export function useAppSearchParams() {
    const [search, setSearch] = useSearchParams();

    const getParam = <param extends UrlParamStateKeys>(key: param, fallback?: UrlParamState[param]) => {
        const encoded = search.get(key);
        if (!encoded) return fallback;

        return decodeURIComponent(encoded) as UrlParamState[param];
    };

    const setParam = (key: keyof UrlParamState, value: string) => {
        search.set(key, encodeURIComponent(value));
        setSearch(search);
    };

    const graphQueryType = getParam('graphQueryType', 'primary');
    const setGraphQueryType = (type: UrlParamState['graphQueryType']) => setParam('graphQueryType', type);

    const graphQuery = getParam('primaryQuery');
    const setGraphQuery = (primaryQuery: UrlParamState['primaryQuery']) => setParam('primaryQuery', primaryQuery);

    const secondaryQuery = getParam('secondaryQuery');
    const setSecondaryQuery = (secondaryQuery: UrlParamState['secondaryQuery']) =>
        setParam('secondaryQuery', secondaryQuery);

    const graphLayout = getParam('graphLayout');
    const setGraphLayout = (graphLayout: UrlParamState['graphLayout']) => setParam('graphLayout', graphLayout);

    const selectedNode = getParam('selectedNode');
    const setSelectedNode = (selectedNode: UrlParamState['selectedNode']) => setParam('selectedNode', selectedNode);

    const selectedNodeType = getParam('selectedNodeType');
    const setSelectedNodeType = (selectedNodeType: UrlParamState['selectedNodeType']) =>
        setParam('selectedNodeType', selectedNodeType);

    const omittedEdges = getParam('omittedEdges');
    const setOmittedEdges = (omittedEdges: UrlParamState['omittedEdges'][0]) => setParam('omittedEdges', omittedEdges);

    return {
        graphQueryType,
        setGraphQueryType,
        graphQuery,
        setGraphQuery,
        secondaryQuery,
        setSecondaryQuery,
        graphLayout,
        setGraphLayout,
        selectedNode,
        setSelectedNode,
        selectedNodeType,
        setSelectedNodeType,
        omittedEdges,
        setOmittedEdges,
    };
}
