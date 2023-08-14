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

import { faSearch, faDirections, faCode, faMinus, faPlus } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Box, Collapse, Paper, Tab, Tabs } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import React, { useState } from 'react';
import CypherInput from './CypherInput';
import NodeSearch from './NodeSearch';
import PathfindingSearch from './PathfindingSearch';
import { PRIMARY_SEARCH } from 'src/ducks/searchbar/types';
import { useSelector } from 'react-redux';
import { AppState } from 'src/store';
import { Icon } from 'bh-shared-ui';

const useStyles = makeStyles((theme) => ({
    menuButton: {
        minWidth: '35px',
        borderRadius: theme.shape.borderRadius,
        borderColor: 'rgba(0,0,0,0.23)',
        color: 'black',
        height: '35px',
        width: '35px',
    },
    icon: {
        height: '40px',
        boxSizing: 'border-box',
        padding: theme.spacing(2),
        fontSize: theme.typography.fontSize,
        color: theme.palette.common.black,
    },
}));

const ExploreSearch = () => {
    const classes = useStyles();

    const searchState = useSelector((state: AppState) => state.search);

    const [showSearchWidget, setShowSearchWidget] = useState(true);

    const [activeTab, setActiveTab] = useState(() => {
        if (searchState.primary.value && searchState.secondary.value) {
            return 1;
        }
        return 0;
    });

    const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
        setActiveTab(newValue);
    };

    return (
        <Box
            sx={{
                position: 'absolute',
                top: '1rem',
                left: '1rem',
                width: activeTab === 2 && showSearchWidget ? '600px' : '410px',
            }}>
            <Paper sx={{ height: '40px', display: 'flex', gap: 1 }} elevation={0}>
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
                    onChange={handleTabChange}
                    onClick={() => setShowSearchWidget(true)}
                    sx={{ height: '40px', minHeight: '40px' }}
                    TabIndicatorProps={{
                        sx: { height: 3, backgroundColor: '#6798B9' },
                    }}>
                    {TabsContent}
                </Tabs>
            </Paper>

            <Collapse in={showSearchWidget}>
                <Paper sx={{ mt: 1, p: 1 }} elevation={0}>
                    <TabPanels
                        tabs={[
                            <NodeSearch searchType={PRIMARY_SEARCH} labelText='Search Nodes' />,
                            <PathfindingSearch />,
                            <CypherInput />,
                        ]}
                        activeTab={activeTab}
                    />
                </Paper>
            </Collapse>
        </Box>
    );
};

const TabsContent = [
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
].map(({ label, icon }) => (
    <Tab
        label={label}
        key={label}
        icon={<FontAwesomeIcon icon={icon} />}
        iconPosition='start'
        sx={{
            height: '40px',
            minHeight: '40px',
            color: 'black',
        }}
    />
));

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
