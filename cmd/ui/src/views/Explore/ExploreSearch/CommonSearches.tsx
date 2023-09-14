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

import {
    List,
    ListSubheader,
    ListItem,
    ListItemText,
    ListItemButton,
    Box,
    Tabs,
    Tab,
    Typography,
    IconButton,
} from '@mui/material';
import { useState } from 'react';
import { CommonSearches as prebuiltSearchList } from 'bh-shared-ui';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrash } from '@fortawesome/free-solid-svg-icons';

interface CommonSearchesProps {
    onClickListItem: (query: string) => void;
}

const ACTIVE_DIRECTORY_TAB = 'Active Directory';
const AZURE_TAB = 'Azure';
const CUSTOM_TAB = 'Custom Searches';

const CommonSearches = ({ onClickListItem }: CommonSearchesProps) => {
    const [activeTab, setActiveTab] = useState(ACTIVE_DIRECTORY_TAB);

    const handleTabChange = (event: React.SyntheticEvent, newValue: string) => {
        setActiveTab(newValue);
    };

    return (
        <Box>
            <Typography variant='h5' sx={{ mb: 2, mt: 2 }}>
                Pre-built Searches
            </Typography>

            <Tabs
                value={activeTab}
                onChange={handleTabChange}
                sx={{ height: '35px', minHeight: '35px', mt: 1 }}
                TabIndicatorProps={{
                    sx: { height: 3, backgroundColor: '#6798B9' },
                }}>
                <Tab
                    label={ACTIVE_DIRECTORY_TAB}
                    key={ACTIVE_DIRECTORY_TAB}
                    value={ACTIVE_DIRECTORY_TAB}
                    sx={{
                        height: '35px',
                        minHeight: '35px',
                        color: 'black',
                    }}
                />
                <Tab
                    label={AZURE_TAB}
                    key={AZURE_TAB}
                    value={AZURE_TAB}
                    sx={{
                        height: '35px',
                        minHeight: '35px',
                        color: 'black',
                    }}
                />
                <Tab
                    label={CUSTOM_TAB}
                    key={CUSTOM_TAB}
                    value={CUSTOM_TAB}
                    sx={{
                        height: '35px',
                        minHeight: '35px',
                        color: 'black',
                    }}
                />
            </Tabs>

            <List
                dense
                disablePadding
                sx={{
                    position: 'relative',
                    overflow: 'auto',
                    maxHeight: 300,
                    '& ul': { padding: 0 },
                }}>
                {activeTab === CUSTOM_TAB ? ( // list of user-saved queries
                    <Box>
                        <ListSubheader sx={{ fontWeight: 'bold' }}>User Saved Searches: </ListSubheader>
                        {userSavedQueries.map((query) => {
                            return (
                                <ListItem
                                    disablePadding
                                    key={query.query}
                                    secondaryAction={
                                        <IconButton size='small'>
                                            <FontAwesomeIcon icon={faTrash} />
                                        </IconButton>
                                    }>
                                    <ListItemButton onClick={() => onClickListItem(query.query)}>
                                        <ListItemText primary={query.name} />
                                    </ListItemButton>
                                </ListItem>
                            );
                        })}
                    </Box>
                ) : (
                    // lsit of pre-built queries
                    prebuiltSearchList
                        .filter(({ category }) => category === activeTab)
                        .map(({ category, subheader, queries }) => {
                            return (
                                <Box key={`${category}-${subheader}`}>
                                    <ListSubheader sx={{ fontWeight: 'bold' }}>{subheader}: </ListSubheader>
                                    {queries.map((query) => {
                                        return (
                                            <ListItem disablePadding key={query.description}>
                                                <ListItemButton onClick={() => onClickListItem(query.cypher)}>
                                                    <ListItemText primary={query.description} />
                                                </ListItemButton>
                                            </ListItem>
                                        );
                                    })}
                                </Box>
                            );
                        })
                )}
            </List>
        </Box>
    );
};

const userSavedQueries = [
    {
        userId: 1,
        query: 'match (n) return n limit 11',
        name: 'special query 1',
    },
    {
        userId: 1,
        query: 'match (n) return n limit 22',
        name: 'special query 2',
    },
];
export default CommonSearches;
