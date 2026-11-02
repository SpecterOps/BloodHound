// Copyright 2026 Specter Ops, Inc.
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
    FileUploadDialog,
    GenericErrorBoundaryFallback,
    useExecuteOnFileDrag,
    useFileUploadDialogContext,
    useKeybindings,
    useKeyboardShortcutsDialogContext,
    useQuickUploadEnabled,
} from 'bh-shared-ui';
import { LoadingSpinner } from 'doodle-ui';
import React, { Suspense, useEffect } from 'react';
import { ErrorBoundary } from 'react-error-boundary';
import { Route, Routes } from 'react-router-dom';
import AuthenticatedRoute from 'src/components/AuthenticatedRoute';
import KeyboardShortcutsDialog from 'src/components/KeyboardShortcutsDialog';
import { ListAssetGroups } from 'src/ducks/assetgroups/actionCreators';
import { fullyAuthenticatedSelector } from 'src/ducks/auth/authSlice';
import { fetchAssetGroups } from 'src/ducks/global/actions';
import { ROUTES } from 'src/routes';
import { useAppDispatch, useAppSelector } from 'src/store';

const Content: React.FC = () => {
    const dispatch = useAppDispatch();
    const authState = useAppSelector((state) => state.auth);
    const { showFileIngestDialog, setShowFileIngestDialog } = useFileUploadDialogContext();
    const { showKeyboardShortcutsDialog, setShowKeyboardShortcutsDialog } = useKeyboardShortcutsDialogContext();
    const isFullyAuthenticated = useAppSelector(fullyAuthenticatedSelector);

    useEffect(() => {
        if (isFullyAuthenticated) {
            dispatch(fetchAssetGroups());
            dispatch(ListAssetGroups());
        }
    }, [authState, isFullyAuthenticated, dispatch]);

    const isQuickUploadEnabled = useQuickUploadEnabled();
    const permitFileUploadModalLaunch = isFullyAuthenticated && isQuickUploadEnabled;

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
        <main className='relative w-full h-full min-h-full'>
            <ErrorBoundary fallbackRender={GenericErrorBoundaryFallback}>
                <Suspense fallback={<LoadingSpinner variant='overlay' spinnerClassName='size-20' />}>
                    <Routes>
                        {ROUTES.map((route) => {
                            return route.authenticationRequired ? (
                                <Route
                                    path={route.path}
                                    element={
                                        <AuthenticatedRoute disallowedRoles={route.disallowedRoles}>
                                            <route.component />
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
        </main>
    );
};

export default Content;
