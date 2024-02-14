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

import { faCode, faDirections, faMinus, faPlus, faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Collapse, Paper, Tab, Tabs, Theme, useMediaQuery, useTheme } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { Icon } from 'bh-shared-ui';
import React, { useState } from 'react';
import { CYPHER_SEARCH, PATHFINDING_SEARCH, PRIMARY_SEARCH } from 'src/ducks/searchbar/types';
import { useAppDispatch, useAppSelector } from 'src/store';
import CypherSearch from './CypherSearch';
import NodeSearch from './NodeSearch';
import PathfindingSearch from './PathfindingSearch';
import { tabChanged, primarySearch, pathfindingSearch, cypherSearch } from 'src/ducks/searchbar/actions';

const useStyles = makeStyles((theme) => ({
    menuButton: {
        borderRadius: theme.shape.borderRadius,
        borderColor: 'rgba(0,0,0,0.23)',
        color: 'black',
        height: '35px',
    },
    icon: {
        height: '40px',
        boxSizing: 'border-box',
        padding: theme.spacing(2),
        fontSize: theme.typography.fontSize,
        color: theme.palette.common.black,
    },
}));

const tabNameMap = {
    primary: 0,
    secondary: 1,
    cypher: 2,
    tierZero: 3,
};

interface ExploreSearchProps {
    handleColumns?: (isCypherEditorActive: boolean) => void;
}

const ExploreSearch = ({ handleColumns }: ExploreSearchProps) => {
    const classes = useStyles();
    const theme = useTheme();
    const matches = useMediaQuery(theme.breakpoints.down('md'));
    const dispatch = useAppDispatch();

    const tabKey = useAppSelector((state) => state.search.activeTab);
    const activeTab = tabNameMap[tabKey];

    const [showSearchWidget, setShowSearchWidget] = useState(true);

    const handleTabChange = (newTabIndex: number) => {
        switch (newTabIndex) {
            case 0:
                dispatch(primarySearch());
                return dispatch(tabChanged(PRIMARY_SEARCH));
            case 1:
                dispatch(pathfindingSearch());
                return dispatch(tabChanged(PATHFINDING_SEARCH));
            case 2:
                dispatch(cypherSearch());
                return dispatch(tabChanged(CYPHER_SEARCH));
        }

        const cypherTabIndex = 2;
        if (handleColumns) {
            handleColumns(newTabIndex === cypherTabIndex);
        }
    };

    return (
        <Box sx={{ pointerEvents: 'auto' }}>
            <Paper sx={{ height: '40px', display: 'flex', flexShrink: 4, gap: 1 }} elevation={0}>
                <Icon
                    className={classes.icon}
                    click={() => {
                        setShowSearchWidget((v) => !v);
                    }}>
                    <FontAwesomeIcon icon={showSearchWidget ? faMinus : faPlus} />
                </Icon>
                <Tabs
                    variant='fullWidth'
                    value={activeTab}
                    onChange={(e, newTabIdx) => handleTabChange(newTabIdx)}
                    onClick={() => setShowSearchWidget(true)}
                    sx={{
                        height: '40px',
                        minHeight: '40px',
                        display: 'flex',
                        justifyContent: 'space-around',
                        width: '100%',
                    }}
                    TabIndicatorProps={{
                        sx: { height: 3, backgroundColor: '#6798B9' },
                    }}>
                    {getTabsContent(theme, matches)}
                </Tabs>
            </Paper>

            <Collapse in={showSearchWidget}>
                <Paper sx={{ mt: 1, p: 1 }} elevation={0}>
                    <TabPanels tabs={[<NodeSearch />, <PathfindingSearch />, <CypherSearch />]} activeTab={activeTab} />
                </Paper>
            </Collapse>
        </Box>
    );
};

const getTabsContent = (theme: Theme, matches: boolean) => {
    const tabs = [
        {
            label: 'Search',
            icon: faSearch,
        },
        {
            label: 'Pathfinding',
            icon: faDirections,
        },
        {
            label: 'Cypher',
            icon: faCode,
        },
    ];

    return tabs.map(({ label, icon }) => (
        <Tab
            label={matches ? '' : label}
            key={label}
            icon={<FontAwesomeIcon icon={icon} />}
            iconPosition='start'
            title={label}
            sx={{
                height: '40px',
                minHeight: '40px',
                color: 'black',
                opacity: 1,
                padding: 0,
                flexGrow: 1,
                minWidth: theme.spacing(2),
            }}
        />
    ));
};

interface TabPanelsProps {
    tabs: React.ReactNode[];
    activeTab: number;
}

const TabPanels = ({ tabs, activeTab }: TabPanelsProps) => {
    return (
        <>
            {tabs.map((tab, index) => {
                return (
                    <Box role='tabpanel' key={index}>
                        {activeTab === index && tab}
                    </Box>
                );
            })}
        </>
    );
};

export default ExploreSearch;
