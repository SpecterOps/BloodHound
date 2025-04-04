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
import { Tab, Tabs, useMediaQuery, useTheme } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import {
    CYPHER_SEARCH,
    CypherSearch,
    Icon,
    NodeSearch,
    PATHFINDING_SEARCH,
    PRIMARY_SEARCH,
    PathfindingSearch,
    cn,
    searchbarActions,
} from 'bh-shared-ui';
import React, { useState } from 'react';
import { useAppDispatch, useAppSelector } from 'src/store';
import {
    useCypherSearchSwitch,
    useNodeSearchSwitch,
    usePathfindingFilterSwitch,
    usePathfindingSearchSwitch,
} from './switches';

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
}));

const tabNameMap = {
    primary: 0,
    secondary: 1,
    cypher: 2,
    tierZero: 3,
};

interface ExploreSearchProps {
    onTabChange?: (tab: string) => void;
}

const ExploreSearch = ({ onTabChange = () => {} }: ExploreSearchProps) => {
    /* Hooks */
    const classes = useStyles();

    const theme = useTheme();

    const matches = useMediaQuery(theme.breakpoints.down('md'));

    const dispatch = useAppDispatch();

    const tabKey = useAppSelector((state) => state.search.activeTab);

    const activeTab = tabNameMap[tabKey];

    const [showSearchWidget, setShowSearchWidget] = useState(true);

    const nodeSearchState = useNodeSearchSwitch();
    const pathfindingSearchState = usePathfindingSearchSwitch();
    const pathfindingFilterState = usePathfindingFilterSwitch();
    const cypherSearchState = useCypherSearchSwitch();

    /* Event Handlers */
    const handleTabChange = (newTabIndex: number) => {
        switch (newTabIndex) {
            case 0:
                onTabChange('search');
                dispatch(searchbarActions.primarySearch());
                return dispatch(searchbarActions.tabChanged(PRIMARY_SEARCH));
            case 1:
                onTabChange('pathfinding');
                dispatch(searchbarActions.pathfindingSearch());
                return dispatch(searchbarActions.tabChanged(PATHFINDING_SEARCH));
            case 2:
                onTabChange('cypher');
                dispatch(searchbarActions.cypherSearch());
                return dispatch(searchbarActions.tabChanged(CYPHER_SEARCH));
        }
    };

    return (
        <div
            className={cn('h-full min-h-0 w-[410px] flex gap-4 flex-col rounded-lg shadow-[1px solid white]', {
                'w-[600px]': activeTab === tabNameMap.cypher && showSearchWidget,
            })}>
            <div className='h-10 w-full flex gap-1 rounded-lg pointer-events-auto bg-[#f4f4f4] dark:bg-[#222222]'>
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
                    className='h-10 min-h-10 w-full'
                    TabIndicatorProps={{
                        className: 'h-[3px]',
                    }}>
                    {getTabsContent(matches)}
                </Tabs>
            </div>

            <div
                className={cn('hidden min-h-0 p-2 rounded-lg pointer-events-auto bg-[#f4f4f4] dark:bg-[#222222]', {
                    block: showSearchWidget,
                })}>
                <TabPanels
                    tabs={[
                        // This linting rule is disabled because the elements in this array do not require a key prop.
                        /* eslint-disable react/jsx-key */
                        <NodeSearch nodeSearchState={nodeSearchState} />,
                        <PathfindingSearch
                            pathfindingSearchState={pathfindingSearchState}
                            pathfindingFilterState={pathfindingFilterState}
                        />,
                        <CypherSearch cypherSearchState={cypherSearchState} />,
                        /* eslint-enable react/jsx-key */
                    ]}
                    activeTab={activeTab}
                />
            </div>
        </div>
    );
};

const getTabsContent = (matches: boolean) => {
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
            className='h-10 min-h-10'
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
                        <div role='tabpanel' key={index} className='h-full'>
                            {tab}
                        </div>
                    );
                } else {
                    return null;
                }
            })}
        </>
    );
};

export default ExploreSearch;
