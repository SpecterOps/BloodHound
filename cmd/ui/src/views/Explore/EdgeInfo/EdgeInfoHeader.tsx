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
import { Typography } from '@mui/material';
import { Icon, collapseAllSections } from 'bh-shared-ui';
import React from 'react';
import { useAppDispatch } from 'src/store';
import { useHeaderStyles } from 'bh-shared-ui';

interface HeaderProps {
    name: string;
    expanded: boolean;
    onToggleExpanded: (expanded: boolean) => void;
}

const Header: React.FC<HeaderProps> = ({ name = 'None Selected', onToggleExpanded, expanded }) => {
    const styles = useHeaderStyles();
    const dispatch = useAppDispatch();

    return (
        <div className={styles.header}>
            <Icon
                className={styles.icon}
                click={() => {
                    onToggleExpanded(!expanded);
                }}>
                <FontAwesomeIcon icon={expanded ? faMinus : faPlus} />
            </Icon>

            <Typography
                data-testid='explore_edge-information-pane_header-text'
                variant='h6'
                noWrap
                className={styles.headerText}>
                {name}
            </Typography>

            <Icon
                tip='Collapse All'
                click={() => {
                    dispatch(collapseAllSections());
                }}
                className={styles.icon}
                data-testid='explore_edge-information-pane_button-collapse-all'>
                <FontAwesomeIcon icon={faAngleDoubleUp} />
            </Icon>
        </div>
    );
};

export default Header;
