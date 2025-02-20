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
import { Box, Paper, Tab, Tabs, useMediaQuery, useTheme } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { ExploreQueryParams, Icon, useExploreParams } from 'bh-shared-ui';
import React, { useState } from 'react';
import CypherSearch from './CypherSearch';
import NodeSearch from './NodeSearch';
import PathfindingSearch from './PathfindingSearch';

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
        color: theme.palette.color.primary,
    },
    tab: {
        height: '40px',
        minHeight: '40px',
        color: theme.palette.primary.main,
        opacity: 1,
        padding: 0,
        flexGrow: 1,
        minWidth: theme.spacing(2),
    },
}));

const tabMap = {
    node: 0,
    pathfinding: 1,
    cypher: 2,
} as const;

const getInitialTab = (searchType: ExploreQueryParams['searchType']) => {
    if (searchType && searchType in tabMap) return searchType as keyof typeof tabMap;
    return 'node';
};

interface ExploreSearchProps {
    onTabChange?: (tab: string) => void;
}

const ExploreSearch = ({ onTabChange = () => {} }: ExploreSearchProps) => {
    /* Hooks */
    const classes = useStyles();

    const theme = useTheme();

    const matches = useMediaQuery(theme.breakpoints.down('md'));

    const { searchType, setExploreParams } = useExploreParams();

    const [activeTab, setActiveTab] = useState<keyof typeof tabMap>(getInitialTab(searchType));

    const [showSearchWidget, setShowSearchWidget] = useState(true);

    /* Event Handlers */
    const handleTabChange = (newTabIndex: number) => {
        switch (newTabIndex) {
            case 0:
                onTabChange('search');
                setActiveTab('node');
                setExploreParams({ searchType: 'node' });
                break;
            case 1:
                onTabChange('pathfinding');
                setActiveTab('pathfinding');
                setExploreParams({ searchType: 'pathfinding' });
                break;
            case 2:
                onTabChange('cypher');
                setActiveTab('cypher');
                setExploreParams({ searchType: 'cypher' });
                break;
        }
    };

    return (
        <Box
            sx={{
                height: '100%',
                display: 'flex',
                flexDirection: 'column',
                minHeight: 0,
                gap: 1,
            }}>
            <Paper
                sx={{
                    height: '40px',
                    display: 'flex',
                    flexShrink: 0,
                    gap: 1,
                    backgroundColor: theme.palette.neutral.secondary,
                    borderRadius: '8px',
                    pointerEvents: 'auto',
                }}
                elevation={0}>
                <Icon
                    className={classes.icon}
                    click={() => {
                        setShowSearchWidget((v) => !v);
                    }}>
                    <FontAwesomeIcon icon={showSearchWidget ? faMinus : faPlus} />
                </Icon>
                <Tabs
                    variant='fullWidth'
                    value={tabMap[activeTab]}
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
                    {getTabsContent(classes.tab, matches)}
                </Tabs>
            </Paper>

            <Box
                display={showSearchWidget ? 'flex' : 'none'}
                sx={{
                    minHeight: 0,
                    flexDirection: 'column',
                }}>
                <Paper
                    sx={{
                        p: 1,
                        backgroundColor: theme.palette.neutral.secondary,
                        borderRadius: '8px',
                        boxSizing: 'border-box',
                        height: '100%',
                        pointerEvents: 'auto',
                    }}
                    elevation={0}>
                    <TabPanels
                        tabs={[
                            // This linting rule is disabled because the elements in this array do not require a key prop.
                            /* eslint-disable react/jsx-key */
                            <NodeSearch />,
                            <PathfindingSearch />,
                            <CypherSearch />,
                            /* eslint-enable react/jsx-key */
                        ]}
                        activeTab={tabMap[activeTab]}
                    />
                </Paper>
            </Box>
        </Box>
    );
};

const getTabsContent = (className: string, matches: boolean) => {
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
            className={className}
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
                if (activeTab === index) {
                    return (
                        <Box role='tabpanel' key={index} height='100%'>
                            {tab}
                        </Box>
                    );
                } else {
                    return null;
                }
            })}
        </>
    );
};

export default ExploreSearch;
