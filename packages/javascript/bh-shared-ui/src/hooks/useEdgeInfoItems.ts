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

import { GraphResponse } from 'js-client-library';
import { useQuery } from 'react-query';
import { useNavigate } from 'react-router-dom';
import { EdgeInfoProps, ExploreQueryParams, NodeResponse, ROUTE_EXPLORE, apiClient, createTypedSearchParams } from '..';
import { VirtualizedNodeListItem } from '../components/VirtualizedNodeList';

export const useEdgeInfoItems = ({
    sourceDBId,
    targetDBId,
    edgeName,
    endpoint,
    queryKey,
}: Partial<EdgeInfoProps> & { endpoint: string; queryKey: string }) => {
    const navigate = useNavigate();

    // To do: cleanup type of apiClient and create mapLiteral for endpoint
    const { data, isLoading, isError } = useQuery([queryKey, sourceDBId, targetDBId, edgeName], () =>
        (apiClient as any)[endpoint](sourceDBId!, targetDBId!, edgeName!).then((result: GraphResponse) => result.data)
    );

    const handleNodeClick = (item: any) => {
        const node = nodesArray[item];
        navigate({
            pathname: ROUTE_EXPLORE,
            search: createTypedSearchParams<ExploreQueryParams>({
                selectedItem: node.graphId,
                primarySearch: node.objectId,
                searchType: 'node',
            }),
        });
    };

    const nodesArray: VirtualizedNodeListItem[] = Object.entries((data?.data?.nodes as NodeResponse) || {}).map(
        ([graphId, node]) => ({
            name: node.label,
            objectId: node.objectId,
            graphId,
            kind: node.kind,
            onClick: handleNodeClick,
        })
    );
    return { isLoading, isError, nodesArray };
};
