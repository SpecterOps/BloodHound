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
import { Outlet } from '@tanstack/react-router';
import {
    FileUploadDialog,
    GenericErrorBoundaryFallback,
    Permission,
    getExcludedIds,
    useExecuteOnFileDrag,
    useFileUploadDialogContext,
    useKeybindings,
    useKeyboardShortcutsDialogContext,
    usePermissions,
} from 'bh-shared-ui';
import React, { Suspense, useEffect } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import KeyboardShortcutsDialog from 'src/components/KeyboardShortcutsDialog';
import { ListAssetGroups } from 'src/ducks/assetgroups/actionCreators';
import { authExpiredSelector, fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import { fetchAssetGroups } from 'src/ducks/global/actions';
import { useAppDispatch, useAppSelector } from 'src/store';

const Content: React.FC = () => {
    const dispatch = useAppDispatch();
    const authState = useAppSelector((state) => state.auth);
    const isAuthExpired = useAppSelector(authExpiredSelector);
    const { showFileIngestDialog, setShowFileIngestDialog } = useFileUploadDialogContext();
    const { showKeyboardShortcutsDialog, setShowKeyboardShortcutsDialog } = useKeyboardShortcutsDialogContext();
    const isFullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);
    const { checkPermission } = usePermissions();
    const hasPermissionToUpload = checkPermission(Permission.GRAPH_DB_INGEST);

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

    useKeybindings({
        KeyH: () => {
            if (isFullyAuthenticated) setShowKeyboardShortcutsDialog(!showKeyboardShortcutsDialog);
        },
        KeyU: () => {
            if (isFullyAuthenticated) setShowFileIngestDialog(!showFileIngestDialog);
        },
    });

    return (
        <Box className='relative w-full h-full min-h-full'>
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
                    <Outlet />
                    {isFullyAuthenticated && (
                        <>
                            <KeyboardShortcutsDialog
                                open={showKeyboardShortcutsDialog}
                                onClose={() => setShowKeyboardShortcutsDialog(false)}
                            />
                            <FileUploadDialog
                                open={showFileIngestDialog}
                                onClose={() => setShowFileIngestDialog(false)}
                            />
                        </>
                    )}
                </Suspense>
            </ErrorBoundary>
        </Box>
    );
};

export default Content;
