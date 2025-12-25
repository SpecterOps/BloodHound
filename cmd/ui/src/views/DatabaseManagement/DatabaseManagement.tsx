// Copyright 2024 Specter Ops, Inc.
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

import { Button } from '@bloodhoundenterprise/doodleui';
import { Alert, Box, Checkbox, FormControl, FormControlLabel, FormGroup, Typography } from '@mui/material';
import {
    DeleteConfirmationDialog,
    FeatureFlag,
    PageWithTitle,
    Permission,
    SourceKindsCheckboxes,
    apiClient,
    useMountEffect,
    useNotifications,
    usePermissions,
} from 'bh-shared-ui';
import { ClearDatabaseRequest } from 'js-client-library';
import { FC, useReducer } from 'react';
import { useMutation } from 'react-query';
import { useSelector } from 'react-redux';
import { selectAllAssetGroupIds, selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';

const initialState: State = {
    deleteAllAssetGroupSelectors: false,
    deleteCollectedGraphData: false,
    deleteCustomHighValueSelectors: false,
    deleteDataQualityHistory: false,
    deleteFileIngestHistory: false,
    deleteHasSessionEdges: false,
    deleteSourceKinds: [],

    noSelectionError: false,
    mutationError: false,
    showSuccessMessage: false,

    openDialog: false,
};

type State = {
    // checkbox state
    deleteAllAssetGroupSelectors: boolean;
    deleteCollectedGraphData: boolean;
    deleteCustomHighValueSelectors: boolean;
    deleteDataQualityHistory: boolean;
    deleteFileIngestHistory: boolean;
    deleteHasSessionEdges: boolean;
    deleteSourceKinds: number[];

    // error state
    noSelectionError: boolean;
    mutationError: boolean;
    mutationErrorMessage?: string;
    showSuccessMessage: boolean;

    // modal state
    openDialog: boolean;
};

type Action =
    | { type: 'no_selection_error' }
    | { type: 'mutation_error'; message?: string }
    | { type: 'mutation_success' }
    | { type: 'selection'; targetName: string; checked: boolean }
    | { type: 'source_kinds'; checked: number[] }
    | { type: 'open_dialog' }
    | { type: 'close_dialog' };

const reducer = (state: State, action: Action): State => {
    switch (action.type) {
        case 'no_selection_error': {
            return {
                ...state,
                noSelectionError: true,
                mutationError: false,
            };
        }
        case 'mutation_error': {
            return {
                ...state,
                mutationError: true,
                noSelectionError: false,
                mutationErrorMessage: action.message,
            };
        }
        case 'mutation_success': {
            return {
                ...state,
                // reset checkboxes
                deleteAllAssetGroupSelectors: false,
                deleteCollectedGraphData: false,
                deleteCustomHighValueSelectors: false,
                deleteDataQualityHistory: false,
                deleteFileIngestHistory: false,
                deleteHasSessionEdges: false,
                deleteSourceKinds: [],

                showSuccessMessage: true,
            };
        }
        case 'selection': {
            const { targetName, checked } = action;
            return {
                ...state,
                [targetName]: checked,
                noSelectionError: false,
            };
        }
        case 'source_kinds': {
            const { checked } = action;
            return {
                ...state,
                deleteSourceKinds: checked,
                noSelectionError: false,
            };
        }
        case 'open_dialog': {
            const noSelection =
                [
                    state.deleteAllAssetGroupSelectors,
                    state.deleteCollectedGraphData,
                    state.deleteCustomHighValueSelectors,
                    state.deleteDataQualityHistory,
                    state.deleteFileIngestHistory,
                    state.deleteHasSessionEdges,
                ].filter(Boolean).length === 0 && state.deleteSourceKinds.length === 0;

            if (noSelection) {
                return {
                    ...state,
                    noSelectionError: true,
                };
            } else {
                return {
                    ...state,
                    noSelectionError: false,
                    openDialog: true,
                };
            }
        }
        case 'close_dialog': {
            return {
                ...state,
                openDialog: false,
            };
        }
        default: {
            return state;
        }
    }
};

const useDatabaseManagement = () => {
    const [state, dispatch] = useReducer(reducer, initialState);

    const allAssetGroupIds = useSelector(selectAllAssetGroupIds);
    const highValueAssetGroupId = useSelector(selectTierZeroAssetGroupId);

    const {
        deleteAllAssetGroupSelectors,
        deleteCollectedGraphData,
        deleteCustomHighValueSelectors,
        deleteDataQualityHistory,
        deleteFileIngestHistory,
        deleteHasSessionEdges,
        deleteSourceKinds,
    } = state;

    const mutation = useMutation({
        mutationFn: async ({ deleteThisData }: { deleteThisData: ClearDatabaseRequest }) => {
            return apiClient.clearDatabase({
                ...deleteThisData,
            });
        },
        onError: (error: any) => {
            // show UI message that data deletion failed
            if (error?.response?.status === 500 && error?.response?.data?.errors?.length > 0) {
                const message = error?.response?.data?.errors?.[0].message;
                dispatch({ type: 'mutation_error', message });
            } else {
                dispatch({ type: 'mutation_error' });
            }
        },
        onSuccess: () => {
            // show UI message that data deletion is happening
            dispatch({ type: 'mutation_success' });
        },
    });

    const handleMutation = () => {
        const assetGroupIds = [];

        if (deleteAllAssetGroupSelectors) {
            assetGroupIds.push(...allAssetGroupIds);
        } else if (deleteCustomHighValueSelectors) {
            assetGroupIds.push(highValueAssetGroupId);
        }

        // dedupe high value asset group id if both checkboxes are selected
        const dedupe = (arr: number[]): number[] => {
            return arr.filter((value, index) => arr.indexOf(value) === index);
        };

        const deleteAssetGroupSelectors = dedupe(assetGroupIds);

        mutation.mutate({
            deleteThisData: {
                deleteAssetGroupSelectors,
                deleteCollectedGraphData,
                deleteDataQualityHistory,
                deleteFileIngestHistory,
                deleteRelationships: deleteHasSessionEdges ? ['HasSession'] : [],
                deleteSourceKinds,
            },
        });
    };

    return { handleMutation, state, dispatch };
};

const DatabaseManagement: FC = () => {
    const { handleMutation, state, dispatch } = useDatabaseManagement();
    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.WIPE_DB);

    const { addNotification, dismissNotification } = useNotifications();
    const notificationKey = 'database-management-permission';

    const effect: React.EffectCallback = () => {
        if (!hasPermission) {
            addNotification(
                `Your user role does not allow managing the database. Please contact your administrator for details.`,
                notificationKey,
                {
                    persist: true,
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );
        }

        return () => dismissNotification(notificationKey);
    };

    useMountEffect(effect);

    const {
        deleteAllAssetGroupSelectors,
        deleteCustomHighValueSelectors,
        deleteDataQualityHistory,
        deleteFileIngestHistory,
        deleteHasSessionEdges,
        deleteSourceKinds,
    } = state;

    const handleCheckbox = (event: React.ChangeEvent<HTMLInputElement>) => {
        dispatch({
            type: 'selection',
            targetName: event.target.name,
            checked: event.target.checked,
        });
    };

    const setSourceKinds = (checked: number[]) => {
        dispatch({
            type: 'source_kinds',
            checked,
        });
    };

    return (
        <PageWithTitle
            title='Database Management'
            data-testid='database-management'
            pageDescription={
                <Typography variant='body2' paragraph>
                    Manage your BloodHound data. Select from the options below which data should be deleted.
                </Typography>
            }>
            <Box>
                <Alert severity='warning' className='mt-4'>
                    <strong>Caution: </strong> This change is irreversible and will delete data from your environment.
                </Alert>

                <Box display='flex' flexDirection='column' alignItems='start'>
                    <FormControl
                        variant='standard'
                        className='py-4'
                        error={state.noSelectionError || state.mutationError}>
                        {state.noSelectionError ? <Alert severity='error'>Please make a selection.</Alert> : null}
                        {state.mutationError ? (
                            <Alert severity='error'>
                                {state.mutationErrorMessage
                                    ? state.mutationErrorMessage
                                    : 'There was an error processing your request.'}
                            </Alert>
                        ) : null}
                        {state.showSuccessMessage ? (
                            <Alert severity='info'>
                                Deletion of the data is under way. Depending on data volume, this may take some time to
                                complete.
                            </Alert>
                        ) : null}

                        <FormGroup className='pt-2'>
                            <FeatureFlag
                                flagKey='clear_graph_data'
                                enabled={
                                    <SourceKindsCheckboxes
                                        checked={deleteSourceKinds}
                                        disabled={!hasPermission}
                                        onChange={setSourceKinds}
                                    />
                                }
                            />
                            <FormControlLabel
                                label='Custom High Value selectors'
                                control={
                                    <Checkbox
                                        checked={deleteCustomHighValueSelectors}
                                        onChange={handleCheckbox}
                                        name='deleteCustomHighValueSelectors'
                                        disabled={!hasPermission}
                                    />
                                }
                            />
                            <FormControlLabel
                                label='All asset group selectors'
                                control={
                                    <Checkbox
                                        checked={deleteAllAssetGroupSelectors}
                                        onChange={handleCheckbox}
                                        name='deleteAllAssetGroupSelectors'
                                        disabled={!hasPermission}
                                    />
                                }
                            />
                            <FormControlLabel
                                label='File ingest log history'
                                control={
                                    <Checkbox
                                        checked={deleteFileIngestHistory}
                                        onChange={handleCheckbox}
                                        name='deleteFileIngestHistory'
                                        disabled={!hasPermission}
                                    />
                                }
                            />
                            <FormControlLabel
                                label='Data quality history'
                                control={
                                    <Checkbox
                                        checked={deleteDataQualityHistory}
                                        onChange={handleCheckbox}
                                        name='deleteDataQualityHistory'
                                        disabled={!hasPermission}
                                    />
                                }
                            />
                            <FormControlLabel
                                label="HasSession edges"
                                control={
                                    <Checkbox
                                        checked={deleteHasSessionEdges}
                                        onChange={handleCheckbox}
                                        name='deleteHasSessionEdges'
                                        disabled={!hasPermission}
                                    />
                                }
                            />
                        </FormGroup>
                    </FormControl>

                    <Button disabled={!hasPermission} onClick={() => dispatch({ type: 'open_dialog' })}>
                        Delete
                    </Button>
                </Box>
            </Box>

            <DeleteConfirmationDialog
                open={state.openDialog}
                onCancel={() => {
                    dispatch({ type: 'close_dialog' });
                }}
                onConfirm={() => {
                    dispatch({ type: 'close_dialog' });
                    handleMutation();
                }}
                itemName='data from the current environment'
                itemType='environment data'
            />
        </PageWithTitle>
    );
};

export default DatabaseManagement;
