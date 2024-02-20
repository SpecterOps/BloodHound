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

import { Box } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import React, { useEffect } from 'react';
import { useQueryClient } from 'react-query';
import { useLocation } from 'react-router-dom';
import { AppNotifications } from 'bh-shared-ui';
import Notifier from 'src/components/Notifier';
import { initialize } from 'src/ducks/auth/authSlice';
import { ROUTE_EXPIRED_PASSWORD, ROUTE_LOGIN, ROUTE_USER_DISABLED } from 'src/ducks/global/routes';
import { featureFlagKeys, getFeatureFlags } from 'src/hooks/useFeatureFlags';
import { useAppDispatch, useAppSelector } from 'src/store';
import { initializeBHEClient } from 'src/utils';
import Content from 'src/views/Content';
import Header from 'src/components/Header';

const useStyles = makeStyles((theme) => ({
    applicationContainer: {
        display: 'flex',
        position: 'relative',
        flexDirection: 'column',
        height: '100%',
        overflow: 'hidden',
    },
    applicationHeader: {
        flexGrow: 0,
        zIndex: theme.zIndex.drawer + 1,
    },
    applicationContent: {
        flexGrow: 1,
        overflowY: 'auto',
        overflowX: 'hidden',
    },
}));

const App: React.FC = () => {
    const classes = useStyles();
    const authState = useAppSelector((state) => state.auth);
    const queryClient = useQueryClient();
    const dispatch = useAppDispatch();
    const location = useLocation();

    // initialize authentication state and BHE client request/response handlers
    useEffect(() => {
        if (!authState.isInitialized) {
            dispatch(initialize());
            initializeBHEClient();
        }
    }, [dispatch, authState.isInitialized]);

    // prefetch feature flags
    useEffect(() => {
        queryClient.prefetchQuery(featureFlagKeys.all, getFeatureFlags);
    }, [queryClient]);

    // block rendering until authentication initialization is complete
    if (!authState.isInitialized) {
        return null;
    }

    const showHeader = !['', '/', ROUTE_LOGIN, ROUTE_EXPIRED_PASSWORD, ROUTE_USER_DISABLED].includes(location.pathname);

    return (
        <>
            <Box className={classes.applicationContainer}>
                {showHeader && (
                    <Box className={classes.applicationHeader}>
                        <Header />
                    </Box>
                )}
                <Box className={classes.applicationContent}>
                    <Content />
                </Box>
                <AppNotifications />
                <Notifier />
            </Box>
        </>
    );
};

export default App;
