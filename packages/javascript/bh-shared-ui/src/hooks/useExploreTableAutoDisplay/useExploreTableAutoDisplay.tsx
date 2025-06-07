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

import { useEffect } from 'react';
// import { isGraphResponse, useExploreGraph } from '../useExploreGraph';
// import { useExploreParams } from '../useExploreParams';
// import { useFeatureFlag } from '../useFeatureFlags';

export const useExploreTableAutoDisplay = (setSelectedLayout: (layout: 'table') => void) => {
    // const { data: graphData } = useExploreGraph();
    // const { searchType } = useExploreParams();
    // const { data: featureFlag } = useFeatureFlag('explore_table_view');

    // const isCypherSearch = searchType === 'cypher';
    // const autoDisplayTableQueryCandidate = isCypherSearch && graphData && isGraphResponse(graphData);

    useEffect(() => {
        // if (autoDisplayTableQueryCandidate) {
        // const emptyEdges = isEmpty(graphData.data.edges);
        // const containsNodes = !isEmpty(graphData.data.nodes);

        // if (emptyEdges && containsNodes) {
        setSelectedLayout('table');
        // }
        // }
    }, [
        /* autoDisplayTableQueryCandidate, featureFlag?.enabled, graphData */
        setSelectedLayout,
    ]);
};
