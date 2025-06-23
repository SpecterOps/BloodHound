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
import {
    CypherSearch,
    Icon,
    NodeSearch,
    PathfindingSearch,
    cn,
    encodeCypherQuery,
    useCypherSearch,
    useExploreParams,
    useNodeSearch,
    usePathfindingSearch,
    type ExploreQueryParams,
    type ExploreSearchTab,
    type MappedStringLiteral,
    type PathfindingFilters,
} from 'bh-shared-ui';
import React, { useState } from 'react';

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

const ExploreSearch: React.FC<{ pathfindingFilters: PathfindingFilters }> = ({ pathfindingFilters }) => {
    /* Hooks */
    const classes = useStyles();

    const theme = useTheme();

    const matches = useMediaQuery(theme.breakpoints.down('md'));

    const { exploreSearchTab, setExploreParams } = useExploreParams();

    const nodeSearchState = useNodeSearch();
    const pathfindingSearchState = usePathfindingSearch();
    const cypherSearchState = useCypherSearch();

    const activeTab = getTab(exploreSearchTab);

    const [showSearchWidget, setShowSearchWidget] = useState(true);

    /* Event Handlers */
    const handleTabChange = (newTabIndex: number) => {
        const tabs = ['node', 'pathfinding', 'cypher'] as ExploreSearchTab[];
        const nextTab = tabs[newTabIndex];

        const teardownParams = teardownActiveTab(activeTab);
        const setupParams = setupNextTab(nextTab);

        setExploreParams({ ...teardownParams, ...setupParams });
    };

    // Clean up query params from previous tab, removing query params when the field has been edited since the last search
    const teardownActiveTab = (tab: ExploreSearchTab): Partial<ExploreQueryParams> => {
        const params: Partial<ExploreQueryParams> = {};

        if (tab === 'node') {
            if (!nodeSearchState.selectedItem) {
                params.primarySearch = null;
                pathfindingSearchState.handleSourceNodeEdited(nodeSearchState.searchTerm);
            }
        }
        if (tab === 'pathfinding') {
            if (!pathfindingSearchState.sourceSelectedItem) {
                params.primarySearch = null;
                nodeSearchState.editSourceNode(pathfindingSearchState.sourceSearchTerm);
            }
            if (!pathfindingSearchState.destinationSelectedItem) {
                params.secondarySearch = null;
            }
        }
        if (tab === 'cypher') {
            if (!cypherSearchState.cypherQuery) {
                params.cypherSearch = null;
            }
        }
        return params;
    };

    // Set up up query params for the incoming tab. should only update the query type if the query can be performed
    const setupNextTab = (tab: ExploreSearchTab): Partial<ExploreQueryParams> => {
        const params: Partial<ExploreQueryParams> = {};

        if (tab === 'node') {
            if (nodeSearchState.selectedItem) {
                params.searchType = 'node';
            }
            params.exploreSearchTab = 'node';
        }
        if (tab === 'pathfinding') {
            if (pathfindingSearchState.sourceSelectedItem && pathfindingSearchState.destinationSelectedItem) {
                params.searchType = 'pathfinding';
            } else if (pathfindingSearchState.sourceSelectedItem || pathfindingSearchState.destinationSelectedItem) {
                params.searchType = 'node';
            }
            params.exploreSearchTab = 'pathfinding';
        }
        if (tab === 'cypher') {
            params.searchType = 'cypher';
            params.cypherSearch = encodeCypherQuery(cypherSearchState.cypherQuery);
            params.exploreSearchTab = 'cypher';
        }
        return params;
    };

    return (
        <div
            data-testid='explore_search-container'
            className={cn('h-full min-h-0 w-[410px] flex gap-4 flex-col rounded-lg shadow-[1px solid white]', {
                'w-[600px]': activeTab === 'cypher' && showSearchWidget,
            })}>
            <div
                className='h-10 w-full flex gap-1 rounded-lg pointer-events-auto bg-[#f4f4f4] dark:bg-[#222222]'
                data-testid='explore_search-container_header'>
                <Icon
                    data-testid='explore_search-container_header_expand-collapse-button'
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
                        <NodeSearch nodeSearchState={nodeSearchState} />,
                        <PathfindingSearch
                            pathfindingSearchState={pathfindingSearchState}
                            pathfindingFilterState={pathfindingFilters}
                        />,
                        <CypherSearch cypherSearchState={cypherSearchState} />,
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
            data-testid={`explore_search-container_header_${label.toLowerCase()}-tab`}
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
