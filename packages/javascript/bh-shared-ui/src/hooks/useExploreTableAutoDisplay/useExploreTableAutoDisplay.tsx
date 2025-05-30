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

import isEmpty from 'lodash/isEmpty';
import { useEffect } from 'react';
import { DEV_TABLE_VIEW } from '../../constants';
import { isGraphResponse, useExploreGraph } from '../useExploreGraph';
import { useExploreParams } from '../useExploreParams';

export const useExploreTableAutoDisplay = (setSelectedLayout: (layout: 'table') => void) => {
    const { data: graphData } = useExploreGraph();
    const { searchType } = useExploreParams();

    const isCypherSearch = searchType === 'cypher';
    const autoDisplayTableQueryCandidate = isCypherSearch && graphData && isGraphResponse(graphData);

    useEffect(() => {
        if (DEV_TABLE_VIEW && autoDisplayTableQueryCandidate) {
            const emptyEdges = isEmpty(graphData.data.edges);
            const containsNodes = !isEmpty(graphData.data.nodes);

            if (emptyEdges && containsNodes) {
                setSelectedLayout('table');
            }
        }
    }, [autoDisplayTableQueryCandidate, graphData, setSelectedLayout]);
};
