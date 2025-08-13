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

import { SxProps, useTheme } from '@mui/material';
import {
    DeepSniffInfoPane,
    EdgeInfoPane,
    EntityInfoDataTable,
    EntityInfoDataTableGraphed,
    EntityInfoPanel,
    EntityKinds,
    isEdge,
    isNode,
    useExploreGraph,
    useExploreSelectedItem,
} from 'bh-shared-ui';
import { useEffect, useState } from 'react';

const GraphItemInformationPanel = () => {
    const { selectedItem, selectedItemQuery } = useExploreSelectedItem();
    // Must call hooks unconditionally at top level
    const graphQuery = useExploreGraph();
    const isDeepSniff = Boolean((graphQuery.data as any)?.deepSniff);
    const [dismissedDeepSniff, setDismissedDeepSniff] = useState(false);
    // Dismiss deep sniff when a selection is made
    useEffect(() => {
        if (selectedItem) setDismissedDeepSniff(true);
    }, [selectedItem]);
    // Reset dismissal when a new deep sniff result arrives
    useEffect(() => {
        if (isDeepSniff) setDismissedDeepSniff(false);
    }, [isDeepSniff]);

    const theme = useTheme();

    const infoPaneStyles: SxProps = {
        bottom: 0,
        top: 0,
        marginBottom: theme.spacing(2),
        marginTop: theme.spacing(2),
        maxWidth: theme.spacing(50),
        position: 'absolute',
        right: theme.spacing(2),
        width: theme.spacing(50),
    };
    // Show Deep Sniff pane (unless user dismissed it by selecting an item)
    const deepSniffActive = isDeepSniff && !dismissedDeepSniff;

    if (!deepSniffActive && (!selectedItem || selectedItemQuery.isLoading)) {
        return <EntityInfoPanel sx={infoPaneStyles} selectedNode={null} DataTable={EntityInfoDataTable} />;
    }

    if (deepSniffActive) {
        return <DeepSniffInfoPane sx={infoPaneStyles} />;
    }

    if (selectedItemQuery.isError) {
        return (
            <EntityInfoPanel
                DataTable={EntityInfoDataTableGraphed}
                sx={infoPaneStyles}
                selectedNode={{
                    graphId: selectedItem ?? undefined,
                    id: '',
                    name: 'Unknown',
                    type: 'Unknown' as EntityKinds,
                }}
            />
        );
    }

    if (selectedItemQuery.data && isEdge(selectedItemQuery.data)) {
        const selectedEdge = {
            id: selectedItem as string,
            name: selectedItemQuery.data.label || '',
            data: selectedItemQuery.data.properties || {},
            sourceNode: {
                id: selectedItemQuery.data.source,
                objectId: selectedItemQuery.data.sourceNode.objectId,
                name: selectedItemQuery.data.sourceNode.label,
                type: selectedItemQuery.data.sourceNode.kind,
            },
            targetNode: {
                id: selectedItemQuery.data.target,
                objectId: selectedItemQuery.data.targetNode.objectId,
                name: selectedItemQuery.data.targetNode.label,
                type: selectedItemQuery.data.targetNode.kind,
            },
        };
        return <EdgeInfoPane sx={infoPaneStyles} selectedEdge={selectedEdge} />;
    }

    if (selectedItemQuery.data && isNode(selectedItemQuery.data)) {
        const selectedNode = {
            graphId: selectedItem ?? undefined,
            id: selectedItemQuery.data.objectId,
            name: selectedItemQuery.data.label,
            type: selectedItemQuery.data.kind as EntityKinds,
        };
        return (
            <EntityInfoPanel sx={infoPaneStyles} selectedNode={selectedNode} DataTable={EntityInfoDataTableGraphed} />
        );
    }

    // No selection: show default empty info panel
    if (!selectedItem) {
        return <EntityInfoPanel sx={infoPaneStyles} selectedNode={null} DataTable={EntityInfoDataTable} />;
    }

    // Default fallback
    return <EntityInfoPanel sx={infoPaneStyles} selectedNode={null} DataTable={EntityInfoDataTable} />;
};

export default GraphItemInformationPanel;
