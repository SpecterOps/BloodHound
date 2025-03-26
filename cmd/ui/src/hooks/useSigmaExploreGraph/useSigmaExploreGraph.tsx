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

import { ExploreQueryParams, transformToFlatGraphResponse, useExploreGraph, useExploreParams } from 'bh-shared-ui';
import { FlatGraphResponse, GraphResponse } from 'js-client-library';
import { useMemo } from 'react';

export const normalizeGraphDataToSigma = (
    graphData: GraphResponse | FlatGraphResponse | undefined,
    searchType: ExploreQueryParams['searchType']
): FlatGraphResponse => {
    if (!graphData) return {};
    switch (searchType) {
        case 'node':
        case 'relationship': {
            return graphData as FlatGraphResponse;
        }
        case 'cypher':
        case 'composition':
        case 'pathfinding': {
            return transformToFlatGraphResponse(graphData as GraphResponse);
        }
    }
    return {};
};

export const useSigmaExploreGraph = () => {
    const { searchType } = useExploreParams();
    const graphState = useExploreGraph();
    const normalizedGraphData = useMemo(
        () => normalizeGraphDataToSigma(graphState.data, searchType),
        [graphState.data, searchType]
    );
    return { ...graphState, data: normalizedGraphData };
};
