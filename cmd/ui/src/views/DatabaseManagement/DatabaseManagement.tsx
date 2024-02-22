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

import { Alert, Box, Button, Checkbox, FormControl, FormControlLabel, FormGroup, Typography } from '@mui/material';
import { ContentPage, apiClient } from 'bh-shared-ui';
import { useReducer } from 'react';
import ConfirmationDialog from './ConfirmationDialog';
import { useMutation } from 'react-query';
import { useSelector } from 'react-redux';
import { selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';

type DataTypes = {
    collectedGraphData: boolean;
    highValueSelectors: boolean;
    fileIngestHistory: boolean;
    dataQualityHistory: boolean;
};

const initialState: State = {
    collectedGraphData: false,
    highValueSelectors: false,
    fileIngestHistory: false,
    dataQualityHistory: false,

    noSelectionError: false,
    mutationError: false,
    showSuccessMessage: false,

    openDialog: false,
};

type State = {
    // checkbox state
    collectedGraphData: boolean;
    highValueSelectors: boolean;
    fileIngestHistory: boolean;
    dataQualityHistory: boolean;

    // error state
    noSelectionError: boolean;
    mutationError: boolean;
    showSuccessMessage: boolean;

    // modal state
    openDialog: boolean;
};

type Action =
    | { type: 'no_selection_error' }
    | { type: 'mutation_error' }
    | { type: 'mutation_success' }
    | { type: 'selection'; targetName: string; checked: boolean }
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
            };
        }
        case 'mutation_success': {
            return {
                ...state,
                // reset checkboxes
                collectedGraphData: false,
                dataQualityHistory: false,
                fileIngestHistory: false,
                highValueSelectors: false,

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
        case 'open_dialog': {
            const { collectedGraphData, dataQualityHistory, fileIngestHistory, highValueSelectors } = state;
            if (
                [collectedGraphData, dataQualityHistory, fileIngestHistory, highValueSelectors].filter(Boolean)
                    .length === 0
            ) {
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
    const { collectedGraphData, highValueSelectors, fileIngestHistory, dataQualityHistory } = state;

    const tierZeroAssetGroupId = useSelector(selectTierZeroAssetGroupId);

    const mutation = useMutation({
        mutationFn: async ({ deleteThisData, assetGroupId }: { deleteThisData: DataTypes; assetGroupId: number }) => {
            return apiClient.databaseManagement({
                ...deleteThisData,
                assetGroupId,
            });
        },
        onError: () => {
            // show UI message that data deletion failed
            dispatch({ type: 'mutation_error' });
        },
        onSuccess: () => {
            // show UI message that data deletion is happening
            dispatch({ type: 'mutation_success' });
        },
    });

    const handleMutation = () => {
        mutation.mutate({
            deleteThisData: {
                collectedGraphData,
                dataQualityHistory,
                fileIngestHistory,
                highValueSelectors,
            },
            assetGroupId: tierZeroAssetGroupId,
        });
    };

    return { handleMutation, state, dispatch };
};

const DatabaseManagement = () => {
    const { handleMutation, state, dispatch } = useDatabaseManagement();

    const { collectedGraphData, highValueSelectors, fileIngestHistory, dataQualityHistory } = state;

    const handleCheckbox = (event: React.ChangeEvent<HTMLInputElement>) => {
        dispatch({
            type: 'selection',
            targetName: event.target.name,
            checked: event.target.checked,
        });
    };

    return (
        <ContentPage title='Clear BloodHound Data'>
            <Box>
                <Typography variant='body1'>
                    Manage your BloodHound data. Select from the options below which data should be deleted.
                </Typography>
                <Alert severity='warning' sx={{ mt: '1rem' }}>
                    <strong>Caution: </strong> This change is irreversible and will delete data from your environment.
                </Alert>

                <Box display='flex' flexDirection='column' alignItems='start'>
                    <FormControl
                        variant='standard'
                        sx={{ paddingBlock: 2 }}
                        error={state.noSelectionError || state.mutationError}>
                        {state.noSelectionError ? <Alert severity='error'>Please make a selection.</Alert> : null}
                        {state.mutationError ? (
                            <Alert severity='error'>There was an error processing your request.</Alert>
                        ) : null}
                        {state.showSuccessMessage ? (
                            <Alert severity='info'>
                                Deletion of the data is under way. Depending on data volume, this may take some time to
                                complete.
                            </Alert>
                        ) : null}

                        <FormGroup sx={{ paddingTop: 1 }}>
                            <FormControlLabel
                                label='Collected graph data (all nodes and edges)'
                                control={
                                    <Checkbox
                                        checked={collectedGraphData}
                                        onChange={handleCheckbox}
                                        name='collectedGraphData'
                                    />
                                }
                            />
                            <FormControlLabel
                                label='Custom High Value selectors'
                                control={
                                    <Checkbox
                                        checked={highValueSelectors}
                                        onChange={handleCheckbox}
                                        name='highValueSelectors'
                                    />
                                }
                            />
                            <FormControlLabel
                                label='File Ingest Log history'
                                control={
                                    <Checkbox
                                        checked={fileIngestHistory}
                                        onChange={handleCheckbox}
                                        name='fileIngestHistory'
                                    />
                                }
                            />
                            <FormControlLabel
                                label='Data Quality history'
                                control={
                                    <Checkbox
                                        checked={dataQualityHistory}
                                        onChange={handleCheckbox}
                                        name='dataQualityHistory'
                                    />
                                }
                            />
                        </FormGroup>
                    </FormControl>

                    <Button
                        color='primary'
                        variant='contained'
                        disableElevation
                        sx={{ width: '150px' }}
                        onClick={() => dispatch({ type: 'open_dialog' })}>
                        Proceed
                    </Button>
                </Box>
            </Box>

            <ConfirmationDialog
                open={state.openDialog}
                handleClose={() => dispatch({ type: 'close_dialog' })}
                handleDelete={() => handleMutation()}
            />
        </ContentPage>
    );
};

export default DatabaseManagement;
