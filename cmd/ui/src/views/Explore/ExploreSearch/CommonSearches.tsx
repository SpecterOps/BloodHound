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

import { Box, Tabs, Tab, Typography } from '@mui/material';
import { useState } from 'react';
import { PrebuiltSearchList, CommonSearches as prebuiltSearchList, PersonalSearchList } from 'bh-shared-ui';
import makeStyles from '@mui/styles/makeStyles';
import { useDispatch } from 'react-redux';
import { cypherQueryEdited } from 'src/ducks/searchbar/actions';
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

const CommonSearches = () => {
    const classes = useStyles();
    const dispatch = useDispatch();

    const [activeTab, setActiveTab] = useState(AD_TAB);

    const handleTabChange = (event: React.SyntheticEvent, newValue: string) => {
        setActiveTab(newValue);
    };

    const adSections = prebuiltSearchList
        .filter(({ category }) => category === 'Active Directory')
        .map(({ subheader, queries }) => ({ subheader, lineItems: queries }));

    const azSections = prebuiltSearchList
        .filter(({ category }) => category === 'Azure')
        .map(({ subheader, queries }) => ({ subheader, lineItems: queries }));

    const handleClick = (query: string) => {
        dispatch(cypherQueryEdited(query));
        dispatch(startCypherQuery(query));
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

            {activeTab === AD_TAB && <PrebuiltSearchList listSections={adSections} clickHandler={handleClick} />}
            {activeTab === AZ_TAB && <PrebuiltSearchList listSections={azSections} clickHandler={handleClick} />}
            {activeTab === CUSTOM_TAB && <PersonalSearchList clickHandler={handleClick} />}
        </Box>
    );
};

export default CommonSearches;
