// Copyright 2024 Specter Ops, Inc.
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

    const getParam = <param extends UrlParamStateKeys>(key: param) => {
        const encoded = search.get(key);
        if (!encoded) return;

        return decodeURIComponent(encoded) as UrlParamState[param];
    };

    const setAppSearchParam = (key: keyof UrlParamState, value: string | string[]) => {
        if (typeof value === 'string') {
            search.set(key, encodeURIComponent(value));
            setSearch(search);
        } else {
            search.delete(key);
            value.forEach((value) => search.append(key, value));
            setSearch(search);
        }
    };

    const graphQueryType = getParam('graphQueryType');

    const graphQuery = getParam('primaryQuery');

    const secondaryQuery = getParam('secondaryQuery');

    const graphLayout = getParam('graphLayout');

    const selectedNode = getParam('selectedNode');

    const selectedNodeType = getParam('selectedNodeType');

    const omittedEdges = getParam('omittedEdges');

    return {
        setAppSearchParam,
        graphQueryType,
        graphQuery,
        secondaryQuery,
        graphLayout,
        selectedNode,
        selectedNodeType,
        omittedEdges,
    };
}
