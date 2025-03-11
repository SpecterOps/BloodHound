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

import { faCode, faDirections, faMinus, faPlus, faSearch } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { Tab, Tabs, useMediaQuery, useTheme } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import { ExploreQueryParams, ExploreSearchTab, Icon, MappedStringLiteral, cn, useExploreParams } from 'bh-shared-ui';
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
}));

const tabMap = {
    node: 0,
    pathfinding: 1,
    cypher: 2,
} satisfies MappedStringLiteral<ExploreSearchTab, number>;

const getTab = (exploreSearchTab: ExploreQueryParams['exploreSearchTab']) => {
    if (exploreSearchTab && exploreSearchTab in tabMap) return exploreSearchTab as keyof typeof tabMap;
    return 'node';
};

const ExploreSearchV2: React.FC = () => {
    /* Hooks */
    const classes = useStyles();

    const theme = useTheme();

    const matches = useMediaQuery(theme.breakpoints.down('md'));

    const { exploreSearchTab, setExploreParams } = useExploreParams();

    const activeTab = getTab(exploreSearchTab);

    const [showSearchWidget, setShowSearchWidget] = useState(true);

    /* Event Handlers */
    const handleTabChange = (newTabIndex: number) => {
        switch (newTabIndex) {
            case 0:
                setExploreParams({ searchType: 'node', exploreSearchTab: 'node' });
                break;
            case 1:
                setExploreParams({ searchType: 'pathfinding', exploreSearchTab: 'pathfinding' });
                break;
            case 2:
                setExploreParams({ searchType: 'cypher', exploreSearchTab: 'cypher' });
                break;
        }
    };

    return (
        <div
            className={cn('h-full min-h-0 w-[410px] flex gap-4 flex-col rounded-lg shadow-[1px solid white]', {
                'w-[600px]': activeTab === 'cypher' && showSearchWidget,
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
                    value={tabMap[activeTab]}
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
                        <NodeSearch />,
                        <PathfindingSearch />,
                        <CypherSearch />,
                        /* eslint-enable react/jsx-key */
                    ]}
                    activeTab={tabMap[activeTab]}
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

export default ExploreSearchV2;
