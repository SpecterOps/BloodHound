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
import { faAngleDoubleUp, faRemove } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Typography } from '@mui/material';
import React from 'react';
import Icon from '../../components/Icon';
import NodeIcon from '../../components/NodeIcon/NodeIcon';
import { useExploreParams, useExploreSelectedItem } from '../../hooks';
import { EntityKinds } from '../../utils/content';
import { useHeaderStyles } from '../../views/Explore/InfoStyles';
import { useObjectInfoPanelContext } from '../../views/Explore/providers';

export interface HeaderProps {
    name: string;
    nodeType?: EntityKinds;
}

const Header: React.FC<HeaderProps> = ({ name, nodeType }) => {
    const styles = useHeaderStyles();
    const { setIsObjectInfoPanelOpen } = useObjectInfoPanelContext();
    const { setExploreParams, expandedPanelSections } = useExploreParams();
    const { clearSelectedItem } = useExploreSelectedItem();

    const handleCollapseAll = () => {
        setIsObjectInfoPanelOpen(false);

        if (expandedPanelSections?.length) {
            setExploreParams({
                expandedPanelSections: [],
            });
        }
    };

    return (
        <Box className={styles.header}>
            <Icon className={styles.icon} click={clearSelectedItem}>
                <FontAwesomeIcon icon={faRemove} />
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
