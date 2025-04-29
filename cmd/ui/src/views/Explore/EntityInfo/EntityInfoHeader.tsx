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

import { faAngleDoubleUp, faMinus, faPlus } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Typography } from '@mui/material';
import {
    Icon,
    NodeIcon,
    useExploreParams,
    useFeatureFlag,
    useHeaderStyles,
    useObjectInfoPanelContext,
} from 'bh-shared-ui';
import React from 'react';
import { useEntityInfoPanelContext } from 'src/views/Explore/EntityInfo/EntityInfoPanelContext';

export interface HeaderProps {
    expanded: boolean;
    name: string;
    onToggleExpanded: (expanded: boolean) => void;
    nodeType?: string;
}

const Header: React.FC<HeaderProps> = ({ name, nodeType, onToggleExpanded, expanded }) => {
    const styles = useHeaderStyles();
    const { data: backButtonFlag } = useFeatureFlag('back_button_support');
    const { collapseAllSections } = useEntityInfoPanelContext();
    const { setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();
    const { setExploreParams, expandedPanelSections } = useExploreParams();

    const handleCollapseAll = () => {
        setIsObjectInfoPanelOpen(false);
        if (backButtonFlag?.enabled) {
            if (expandedPanelSections?.length) {
                setExploreParams({
                    expandedPanelSections: [],
                });
            }
        } else {
            collapseAllSections();
        }
    };

    return (
        <Box className={styles.header}>
            <Icon
                className={styles.icon}
                click={() => {
                    onToggleExpanded(!expanded);
                }}>
                <FontAwesomeIcon icon={expanded ? faMinus : faPlus} />
            </Icon>

            {nodeType && <NodeIcon nodeType={nodeType} />}

            <Typography
                data-testid='explore_entity-information-panel_header-text'
                variant='h6'
                noWrap
                className={styles.headerText}>
                {name}
            </Typography>

            <Icon
                tip='Collapse All'
                click={handleCollapseAll}
                className={styles.icon}
                data-testid='explore_entity-information-panel_button-collapse-all'>
                <FontAwesomeIcon icon={faAngleDoubleUp} />
            </Icon>
        </Box>
    );
};

export default Header;
