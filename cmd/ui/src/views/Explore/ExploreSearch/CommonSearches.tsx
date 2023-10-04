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
import { FC, createContext, useContext, useState } from 'react';
import { CommonSearchType, apiClient, CommonSearches as prebuiltSearchList } from 'bh-shared-ui';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faTrash } from '@fortawesome/free-solid-svg-icons';
import makeStyles from '@mui/styles/makeStyles';
import { useMutation, useQuery, useQueryClient } from 'react-query';
import { useDispatch } from 'react-redux';
import { addSnackbar } from 'src/ducks/global/actions';
import { setCypherQueryTerm } from 'src/ducks/searchbar/actions';
import { startCypherQuery } from 'src/ducks/explore/actions';

const AD_TAB = 'Active Directory';
const AZ_TAB = 'Azure';
const CUSTOM_TAB = 'Custom Searches';

const useStyles = makeStyles((theme) => ({
    tabs: {
        height: '35px',
        minHeight: '35px',
        mt: 1,
    },
    tab: {
        height: '35px',
        minHeight: '35px',
        color: 'black',
    },
    list: {
        position: 'relative',
        overflow: 'auto',
        maxHeight: 300,
        '& ul': { padding: 0 },
    },
}));

const PrebuiltSearchContext = createContext();

const PrebuiltSearches = ({ children, activeTab }) => {
    const dispatch = useDispatch();

    const show = (currenTab: string) => currenTab === activeTab;

    const onClick = (query: string) => {
        dispatch(setCypherQueryTerm(query));
        dispatch(startCypherQuery(query));
    };

    return <PrebuiltSearchContext.Provider value={{ show, onClick }}>{children}</PrebuiltSearchContext.Provider>;
};

const CommonSearches = () => {
    const classes = useStyles();

    const [activeTab, setActiveTab] = useState(AD_TAB);

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
                className={classes.tabs}
                TabIndicatorProps={{
                    sx: { height: 3, backgroundColor: '#6798B9' },
                }}>
                <Tab label={AD_TAB} key={AD_TAB} value={AD_TAB} className={classes.tab} />
                <Tab label={AZ_TAB} key={AZ_TAB} value={AZ_TAB} className={classes.tab} />
                <Tab label={CUSTOM_TAB} key={CUSTOM_TAB} value={CUSTOM_TAB} className={classes.tab} />
            </Tabs>

            <PrebuiltSearches activeTab={activeTab}>
                <ActiveDirectoryPrebuiltSearches />
                <AzurePrebuiltSearches />
                <PersonalSearchList />
            </PrebuiltSearches>
        </Box>
    );
};

const ActiveDirectoryPrebuiltSearches = () => {
    const { show } = useContext(PrebuiltSearchContext);

    const byADCategory = ({ category }: { category: string }) => category === AD_TAB;
    const toListSections = ({ subheader, queries }: CommonSearchType) => ({ subheader, lineItems: queries });

    const listSections = prebuiltSearchList.filter(byADCategory).map(toListSections);

    return show ? <SearchList listSections={listSections} /> : null;
};

const AzurePrebuiltSearches = () => {
    const { show } = useContext(PrebuiltSearchContext);

    const byAZCategory = ({ category }: { category: string }) => category === AZ_TAB;
    const toListSections = ({ subheader, queries }: CommonSearchType) => ({ subheader, lineItems: queries });

    const listSections = prebuiltSearchList.filter(byAZCategory).map(toListSections);

    return show ? <SearchList listSections={listSections} /> : null;
};

interface SearchListProps {
    listSections: ListSection[];

    deleteHandler?: any;
}

type ListSection = {
    subheader: string;
    lineItems: LineItem[];
};

type LineItem = {
    id?: number;

    description: string;
    cypher: string;
    canEdit?: boolean;
};

const SearchList: FC<SearchListProps> = ({ listSections, deleteHandler }) => {
    const classes = useStyles();

    const { onClick } = useContext(PrebuiltSearchContext);

    return (
        <List dense disablePadding className={classes.list}>
            {listSections.map((section) => {
                const { subheader, lineItems } = section;

                return (
                    <Box key={subheader}>
                        <ListSubheader sx={{ fontWeight: 'bold' }}>{subheader} </ListSubheader>

                        {lineItems?.map((lineItem) => {
                            const { id, description, cypher, canEdit = false } = lineItem;

                            return (
                                <ListItem
                                    disablePadding
                                    key={id}
                                    secondaryAction={
                                        canEdit && (
                                            <IconButton size='small' onClick={() => deleteHandler(id)}>
                                                <FontAwesomeIcon icon={faTrash} />
                                            </IconButton>
                                        )
                                    }>
                                    <ListItemButton onClick={() => onClick(cypher)}>
                                        <ListItemText primary={description} />
                                    </ListItemButton>
                                </ListItem>
                            );
                        })}
                    </Box>
                );
            })}
        </List>
    );
};

// `PersonalSearchList` is a more specific implementation of `SearchList`.  It includes
// additional fetching logic to fetch and delete queries saved by the user
const PersonalSearchList = () => {
    const dispatch = useDispatch();
    const queryClient = useQueryClient();

    const [queries, setQueries] = useState<LineItem[]>([]);

    useQuery({
        queryKey: 'userSavedQueries',
        queryFn: () => {
            return apiClient
                .getUserSavedQueries()
                .then((response) => {
                    const queries = response.data.data;

                    const queriesToDisplay = queries.map((query) => ({
                        description: query.name,
                        cypher: query.query,
                        canEdit: true,
                        id: query.id,
                    }));

                    setQueries(queriesToDisplay);
                })
                .catch(() => {
                    setQueries([]);
                });
        },
    });

    const mutation = useMutation({
        mutationFn: (queryId: number) => {
            return apiClient.deleteUserQuery(queryId);
        },
        onSettled: () => {
            queryClient.invalidateQueries({ queryKey: 'userSavedQueries' });
        },
        onSuccess: () => {
            dispatch(addSnackbar(`Query deleted.`, 'userDeleteQuery'));
        },
    });

    const { show } = useContext(PrebuiltSearchContext);

    return show ? (
        <SearchList
            listSections={[
                {
                    subheader: 'User Saved Searches: ',
                    lineItems: queries,
                },
            ]}
            deleteHandler={mutation.mutate}
        />
    ) : null;
};

export default CommonSearches;
