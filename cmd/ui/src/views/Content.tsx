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

import { Box, CircularProgress } from '@mui/material';
import makeStyles from '@mui/styles/makeStyles';
import {
    CommandDialog,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
    ExploreHistoryDialog,
    FileUploadDialog,
    GenericErrorBoundaryFallback,
    Permission,
    ROUTE_PRIVILEGE_ZONES,
    getExcludedIds,
    useAppNavigate,
    useExecuteOnFileDrag,
    useFeatureFlags,
    useFileUploadDialogContext,
    usePermissions,
} from 'bh-shared-ui';
import React, { Suspense, useEffect } from 'react';

import { ErrorBoundary } from 'react-error-boundary';
import { Route, Routes } from 'react-router-dom';
import AuthenticatedRoute from 'src/components/AuthenticatedRoute';
import { ListAssetGroups } from 'src/ducks/assetgroups/actionCreators';
import { authExpiredSelector, fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import { fetchAssetGroups } from 'src/ducks/global/actions';
import { ROUTES } from 'src/routes';
import {
    ROUTE_ADMINISTRATION_BLOODHOUND_CONFIGURATION,
    ROUTE_ADMINISTRATION_DATA_QUALITY,
    ROUTE_ADMINISTRATION_DB_MANAGEMENT,
    ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES,
    ROUTE_ADMINISTRATION_FILE_INGEST,
    ROUTE_ADMINISTRATION_MANAGE_USERS,
    ROUTE_ADMINISTRATION_SSO_CONFIGURATION,
    ROUTE_API_EXPLORER,
    ROUTE_DOWNLOAD_COLLECTORS,
    ROUTE_EXPLORE,
    ROUTE_GROUP_MANAGEMENT,
    ROUTE_MY_PROFILE,
} from 'src/routes/constants';
import { useAppDispatch, useAppSelector } from 'src/store';
import { endpoints } from 'src/utils';

const useStyles = makeStyles({
    content: {
        position: 'relative',
        width: '100%',
        height: '100%',
        minHeight: '100%',
    },
});

const Content: React.FC = () => {
    const classes = useStyles();
    const dispatch = useAppDispatch();
    const navigate = useAppNavigate();
    const authState = useAppSelector((state) => state.auth);
    const isAuthExpired = useAppSelector(authExpiredSelector);
    const { showFileIngestDialog, setShowFileIngestDialog } = useFileUploadDialogContext();
    const isFullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const { checkPermission } = usePermissions();
    const hasPermissionToUpload = checkPermission(Permission.GRAPH_DB_INGEST);
    const fullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const enableFeatureFlagRequests = !!authState.isInitialized && fullyAuthenticated;
    const featureFlags = useFeatureFlags({ enabled: enableFeatureFlagRequests });
    const tierFlag = featureFlags?.data?.find((flag) => {
        return flag.key === 'tier_management_engine';
    });

    useEffect(() => {
        if (isFullyAuthenticated) {
            dispatch(fetchAssetGroups());
            dispatch(ListAssetGroups());
        }
    }, [authState, isFullyAuthenticated, dispatch]);

    const permitFileUploadModalLaunch =
        !!authState.sessionToken && !!authState.user && !isAuthExpired && !getExcludedIds() && !!hasPermissionToUpload;
    // Display ingest dialog when a processable file is dragged into the browser client
    useExecuteOnFileDrag(() => setShowFileIngestDialog(true), {
        condition: () => permitFileUploadModalLaunch,
        acceptedTypes: ['application/json', 'application/zip'],
    });

    const [commandPaletteOpen, setCommandPaletteOpen] = React.useState(false);
    // const [bloodhoundOpen, setBloodhoundOpen] = React.useState(false);

    useEffect(() => {
        const down = (e: KeyboardEvent) => {
            if (e.key === 'Escape') {
                e.preventDefault();
                setCommandPaletteOpen(false);
            }
            if (e.key === 'k' && (e.metaKey || e.ctrlKey)) {
                e.preventDefault();
                setCommandPaletteOpen((open) => !open);
            }
        };

        document.addEventListener('keydown', down);
        return () => document.removeEventListener('keydown', down);
    }, []);

    const navigateHandler = (path: string) => {
        navigate(path);
        setCommandPaletteOpen(false);
    };

    const showFileInjestDialog = () => {
        setShowFileIngestDialog(true);
        setCommandPaletteOpen(false);
    };

    // const showBloodhound = () => {
    //     setBloodhoundOpen(true);
    //     setCommandPaletteOpen(false);
    // };

    return (
        <Box className={classes.content}>
            <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                <Suspense
                    fallback={
                        <Box
                            position='absolute'
                            top='0'
                            left='0'
                            right='0'
                            bottom='0'
                            display='flex'
                            alignItems='center'
                            justifyContent='center'
                            zIndex={1000}>
                            <CircularProgress color='primary' size={80} />
                        </Box>
                    }>
                    <Routes>
                        {ROUTES.map((route) => {
                            return route.authenticationRequired ? (
                                <Route
                                    path={route.path}
                                    element={
                                        // Note: We add a left padding value to account for pages that have nav bar, h-full is because when adding the div it collapsed the views
                                        <AuthenticatedRoute>
                                            <div className={`h-full ${route.navigation && 'pl-nav-width'} `}>
                                                <route.component />
                                            </div>
                                        </AuthenticatedRoute>
                                    }
                                    key={route.path}
                                />
                            ) : (
                                <Route path={route.path} element={<route.component />} key={route.path} />
                            );
                        })}
                    </Routes>
                    {isFullyAuthenticated && (
                        <FileUploadDialog open={showFileIngestDialog} onClose={() => setShowFileIngestDialog(false)} />
                    )}
                </Suspense>
                <CommandDialog
                    open={commandPaletteOpen}
                    onOpenChange={() => {
                        setCommandPaletteOpen((prev) => !prev);
                    }}
                    className='overflow-visible'>
                    <CommandInput placeholder='Type a command or search...' />
                    <CommandGroup heading='Tools'>
                        <CommandItem>
                            <ExploreHistoryDialog />
                        </CommandItem>
                    </CommandGroup>
                    <CommandList>
                        <CommandEmpty>No results found.</CommandEmpty>
                        <CommandGroup>
                            <div onClick={() => navigateHandler(ROUTE_EXPLORE)}>
                                <CommandItem>Explore</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(`${ROUTE_EXPLORE}?exploreSearchTab=pathfinding`)}>
                                <CommandItem> Explore Pathfinding</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(`${ROUTE_EXPLORE}?exploreSearchTab=cypher`)}>
                                <CommandItem> Explore Cypher</CommandItem>
                            </div>
                            {tierFlag?.enabled ? (
                                <div onClick={() => navigateHandler(ROUTE_PRIVILEGE_ZONES)}>
                                    <CommandItem>Privilege Zones</CommandItem>
                                </div>
                            ) : (
                                <div onClick={() => navigateHandler(ROUTE_GROUP_MANAGEMENT)}>
                                    <CommandItem>Group Management</CommandItem>
                                </div>
                            )}
                            <div onClick={() => navigateHandler(ROUTE_MY_PROFILE)}>
                                <CommandItem>Profile</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(ROUTE_DOWNLOAD_COLLECTORS)}>
                                <CommandItem>Download Collectors</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(ROUTE_API_EXPLORER)}>
                                <CommandItem>API Explorer</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(ROUTE_ADMINISTRATION_BLOODHOUND_CONFIGURATION)}>
                                <CommandItem>Bloodhound Configuration</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(ROUTE_ADMINISTRATION_DATA_QUALITY)}>
                                <CommandItem>Data Quality</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(ROUTE_ADMINISTRATION_DB_MANAGEMENT)}>
                                <CommandItem>Database Management</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(ROUTE_ADMINISTRATION_EARLY_ACCESS_FEATURES)}>
                                <CommandItem>Early Access Features</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(ROUTE_ADMINISTRATION_FILE_INGEST)}>
                                <CommandItem>File Injest</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(ROUTE_ADMINISTRATION_MANAGE_USERS)}>
                                <CommandItem>Manage Users</CommandItem>
                            </div>
                            <div onClick={() => navigateHandler(ROUTE_ADMINISTRATION_SSO_CONFIGURATION)}>
                                <CommandItem>SSO Configuration</CommandItem>
                            </div>
                            <div onClick={showFileInjestDialog}>
                                <CommandItem>Quick Upload</CommandItem>
                            </div>
                            {/* <div onClick={showBloodhound}>
                                <CommandItem>Bloodhound!</CommandItem>
                            </div> */}
                        </CommandGroup>
                        <CommandGroup heading='API Explorer'>
                            {endpoints.map((endpoint) => {
                                const splitString = endpoint.split(' ');
                                return (
                                    <CommandItem
                                        key={endpoint}
                                        onSelect={() => {
                                            setCommandPaletteOpen(false);
                                            navigate('/api-explorer', { state: { scrollTo: endpoint } });
                                        }}>
                                        <span className='font-bold'>{splitString[0].toUpperCase()}</span>
                                        <span>{' ' + splitString[1]}</span>
                                    </CommandItem>
                                );
                            })}
                        </CommandGroup>
                    </CommandList>
                </CommandDialog>
                {/* {bloodhoundOpen && (
                    <div className='absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2'>
                        <Card>
                            <AppIcon.Bloodhound />
                        </Card>
                    </div>
                )} */}
            </ErrorBoundary>
        </Box>
    );
};

export default Content;
