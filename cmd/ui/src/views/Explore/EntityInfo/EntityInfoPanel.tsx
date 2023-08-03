// Copyright 2023 Specter Ops, Inc.
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

import { Paper } from '@mui/material';
import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import usePreviousValue from 'src/hooks/usePreviousValue';
import { AppState } from 'src/store';
import EntityInfoContent from './EntityInfoContent';
import Header from './EntityInfoHeader';
import { EntityInfoPanelContextProvider } from './EntityInfoPanelContextProvider';
import { useEntityInfoPanelContext } from './EntityInfoPanelContext';
import usePaneStyles from 'src/views/Explore/InfoStyles/Pane';

const EntityInfoPanel: React.FC = () => {
    const styles = usePaneStyles();
    const [expanded, setExpanded] = useState(true);
    const selectedNode = useSelector((state: AppState) => state.entityinfo.selectedNode);
    const { setExpandedSections } = useEntityInfoPanelContext();
    const previousSelectedNode = usePreviousValue(selectedNode);

    useEffect(() => {
        if (previousSelectedNode?.id !== selectedNode?.id) {
            setExpandedSections({ 'Object Information': true });
        }
    }, [setExpandedSections, previousSelectedNode, selectedNode]);

    if (selectedNode === null) {
        return (
            <div className={styles.container} data-testid='explore_entity-information-panel'>
                <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                    <Header
                        expanded={expanded}
                        onToggleExpanded={(expanded) => {
                            setExpanded(expanded);
                        }}
                    />
                </Paper>
                <Paper
                    elevation={0}
                    classes={{ root: styles.contentPaperRoot }}
                    style={{
                        display: expanded ? 'initial' : 'none',
                    }}>
                    No information to display.
                </Paper>
            </div>
        );
    }

    return (
        <div className={styles.container} data-testid='explore_entity-information-panel'>
            <Paper elevation={0} classes={{ root: styles.headerPaperRoot }}>
                <Header
                    name={selectedNode?.name}
                    nodeType={selectedNode?.type}
                    expanded={expanded}
                    onToggleExpanded={(expanded) => {
                        setExpanded(expanded);
                    }}
                />
            </Paper>
            <Paper
                elevation={0}
                classes={{ root: styles.contentPaperRoot }}
                style={{
                    display: expanded ? 'initial' : 'none',
                }}>
                <EntityInfoContent id={selectedNode.id} nodeType={selectedNode.type} />
            </Paper>
        </div>
    );
};

const WrappedEntityInfoPanel = () => (
    <EntityInfoPanelContextProvider>
        <EntityInfoPanel />
    </EntityInfoPanelContextProvider>
);

export default WrappedEntityInfoPanel;
