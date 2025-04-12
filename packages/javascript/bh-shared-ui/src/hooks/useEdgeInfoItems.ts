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

import { useQuery } from 'react-query';
import { useNavigate } from 'react-router-dom';
import {
    EdgeInfoProps,
    ExploreQueryParams,
    MappedStringLiteral,
    ROUTE_EXPLORE,
    apiClient,
    createTypedSearchParams,
    useFeatureFlag,
} from '..';
import { VirtualizedNodeListItem } from '../components/VirtualizedNodeList';

type EdgeInfoTypes = 'relayTargets' | 'composition';
type EdgeInfoEndpoints = 'getRelayTargets' | 'getEdgeComposition';
type EdgeInfoQueryKeys = 'relayTargets' | 'edgeComposition';

export const EdgeInfoType = {
    relayTargets: 'relayTargets',
    composition: 'composition',
} satisfies MappedStringLiteral<EdgeInfoTypes, EdgeInfoTypes>;

export const EdgeInfoEndpoints = {
    relayTargets: 'getRelayTargets',
    composition: 'getEdgeComposition',
} satisfies MappedStringLiteral<EdgeInfoTypes, EdgeInfoEndpoints>;

export const EdgeInfoQueryKeys = {
    relayTargets: 'relayTargets',
    composition: 'edgeComposition',
} satisfies MappedStringLiteral<EdgeInfoTypes, EdgeInfoQueryKeys>;

export const useEdgeInfoItems = ({
    sourceDBId,
    targetDBId,
    edgeName,
    type,
}: Pick<EdgeInfoProps, 'sourceDBId' | 'targetDBId' | 'edgeName'> & {
    type: EdgeInfoTypes;
}) => {
    const navigate = useNavigate();
    const { data: backButtonflag } = useFeatureFlag('back_button_support');

    const { data, isLoading, isError } = useQuery([EdgeInfoQueryKeys[type], sourceDBId, targetDBId, edgeName], () =>
        apiClient[EdgeInfoEndpoints[type]](sourceDBId!, targetDBId!, edgeName!).then((result) => result.data)
    );

    const handleNodeClick = (item: any) => {
        const node = nodesArray[item];
        if (backButtonflag?.enabled) {
            navigate({
                pathname: ROUTE_EXPLORE,
                search: createTypedSearchParams<ExploreQueryParams>({
                    selectedItem: node.graphId,
                    primarySearch: node.objectId,
                    searchType: 'node',
                }),
            });
        }
    };

    const nodesArray: VirtualizedNodeListItem[] = Object.entries(data?.data?.nodes || {}).map(([graphId, node]) => ({
        name: node.label,
        objectId: node.objectId,
        graphId,
        kind: node.kind,
        onClick: handleNodeClick,
    }));
    return { isLoading, isError, nodesArray };
};
