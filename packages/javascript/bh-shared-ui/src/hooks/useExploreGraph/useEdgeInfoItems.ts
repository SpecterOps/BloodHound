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
import { EdgeInfoProps } from '../../components';
import { NormalizedNodeItem } from '../../components/VirtualizedNodeList';
import { apiClient } from '../../utils';
import { useExploreParams } from '../useExploreParams';

export enum EdgeInfoItems {
    relayTargets = 'relayTargets',
    composition = 'composition',
    aclInheritance = 'aclinheritance',
}

type EdgeInfoItemsArguments = Pick<EdgeInfoProps, 'sourceDBId' | 'targetDBId' | 'edgeName' | 'onNodeClick'>;

export type EdgeInfoItemsProps = EdgeInfoItemsArguments & {
    type: EdgeInfoItems;
};

type EdgeInfoItemOpts = {
    withProperties: boolean;
};

const queryConfig = {
    [EdgeInfoItems.relayTargets]: {
        endpoint: ({ sourceDBId, targetDBId, edgeName }: EdgeInfoItemsArguments) => {
            return apiClient.getRelayTargets(sourceDBId!, targetDBId!, edgeName!).then((result) => result.data);
        },
    },
    [EdgeInfoItems.composition]: {
        endpoint: ({ sourceDBId, targetDBId, edgeName }: EdgeInfoItemsArguments) => {
            return apiClient.getEdgeComposition(sourceDBId!, targetDBId!, edgeName!).then((result) => result.data);
        },
    },
    [EdgeInfoItems.aclInheritance]: {
        endpoint: ({ sourceDBId, targetDBId, edgeName }: EdgeInfoItemsArguments) => {
            return apiClient.getACLInheritance(sourceDBId!, targetDBId!, edgeName!).then((result) => result.data);
        },
    },
};

export const useEdgeInfoItems = (
    { sourceDBId, targetDBId, edgeName, type }: EdgeInfoItemsProps,
    opts?: EdgeInfoItemOpts
) => {
    const { setExploreParams } = useExploreParams();
    const { data, isLoading, isError } = useQuery(
        [type, sourceDBId, targetDBId, edgeName],
        () => queryConfig[type].endpoint({ sourceDBId, targetDBId, edgeName }),
        { enabled: !!(Number.isInteger(sourceDBId) && Number.isInteger(targetDBId) && edgeName) }
    );

    const handleNodeClick = (item: number) => {
        const node = nodesArray[item];

        setExploreParams({
            primarySearch: node.objectId,
            searchType: 'node',
            exploreSearchTab: 'node',
        });
    };

    const nodesArray: NormalizedNodeItem[] = Object.entries(data?.data?.nodes || {}).map(([graphId, node]) => ({
        name: node.label,
        objectId: node.objectId,
        graphId,
        kind: node.kind,
        ...(opts?.withProperties && { properties: node.properties }),
        onClick: handleNodeClick,
    }));
    return { isLoading, isError, nodesArray };
};
