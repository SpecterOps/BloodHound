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

import { isEmpty } from 'lodash';
import { useEffect, useState } from 'react';
import { isGraphResponse, useExploreGraph } from '../useExploreGraph';
import { useExploreParams } from '../useExploreParams';
import { useFeatureFlag } from '../useFeatureFlags';

// This should be able to detect when the Explore Table should display automatically and when NOT to display it.
// Auto display when current search is cypher and returned data contains nodes but not edges.
// And dont auto display if the auto display has been closed.
export const useExploreTableAutoDisplay = () => {
    const { data: graphData, isFetching } = useExploreGraph();
    const { searchType } = useExploreParams();
    const { data: featureFlag } = useFeatureFlag('explore_table_view');

    // Tracks if the current query has triggered the table.
    // Resets when the query fetches, which includes initial fetch and fetches from the cache
    const [hasTriggered, setHasTriggered] = useState(false);
    const [autoDisplayTable, setAutoDisplayTable] = useState(false);

    const isCypherSearch = searchType === 'cypher';
    const autoDisplayTableQueryCandidate = !!(
        isCypherSearch && // auto display only on cypher search
        !hasTriggered && // check that it hasnt already triggered for this query
        graphData && // type check the response for type safety
        isGraphResponse(graphData)
    );

    // Resets the trigger on every fetch of a query, even fetches from the cache.
    useEffect(() => {
        if (isFetching) {
            setHasTriggered(false);
        }
    }, [isFetching]);

    // checks if current query is a candidate to auto display the table
    // if it is, and the query is nodes only then call onAutoDisplayChange.
    useEffect(() => {
        if (autoDisplayTableQueryCandidate) {
            const emptyEdges = isEmpty(graphData.data.edges);
            const containsNodes = !isEmpty(graphData.data.nodes);

            const shouldAutoDisplay = emptyEdges && containsNodes;
            if (shouldAutoDisplay) {
                // set triggered to true so that we dont try to reopen the table after closing it.
                setHasTriggered(true);
            }
            // This will automatically open/close the table
            setAutoDisplayTable(shouldAutoDisplay);
        }
    }, [autoDisplayTableQueryCandidate, featureFlag?.enabled, graphData?.data]);

    return [autoDisplayTable, setAutoDisplayTable] as const;
};
