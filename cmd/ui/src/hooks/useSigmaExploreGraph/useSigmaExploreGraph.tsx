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

import { isGraphResponse, transformToSigmaFlatGraphData, useExploreGraph } from 'bh-shared-ui';
import { GraphData, GraphResponse, SigmaFlatGraphData } from 'js-client-library';
import { useMemo } from 'react';

export const normalizeGraphDataForSigma = (
    graphData: GraphResponse | SigmaFlatGraphData | undefined
): SigmaFlatGraphData | undefined => {
    if (!graphData) return;

    if (isGraphResponse(graphData)) {
        return transformToSigmaFlatGraphData(graphData);
    } else {
        return graphData;
    }
};

export const useSigmaExploreGraph = (includeProperties: boolean) => {
    const graphQuery = useExploreGraph(includeProperties);
    const normalizedGraphData = useMemo(() => normalizeGraphDataForSigma(graphQuery.data), [graphQuery.data]);

    // return the full query so we can know loading/error state, and use react-query tools. But override the data field with the normalized value
    return {
        ...graphQuery,
        data: {
            nodes: normalizedGraphData,
            node_keys: (graphQuery?.data?.data as GraphData)?.node_keys,
        },
    };
};
