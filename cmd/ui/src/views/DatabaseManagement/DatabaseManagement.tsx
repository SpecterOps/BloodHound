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

import {
    Box,
    Button,
    Checkbox,
    FormControl,
    FormControlLabel,
    FormGroup,
    FormHelperText,
    Typography,
    useTheme,
} from '@mui/material';
import { ContentPage, apiClient } from 'bh-shared-ui';
import { useState } from 'react';
import ConfirmationDialog from './ConfirmationDialog';
import { useMutation } from 'react-query';
import { useSelector } from 'react-redux';
import { selectTierZeroAssetGroupId } from 'src/ducks/assetgroups/reducer';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import { faCircleXmark, faCircleExclamation } from '@fortawesome/free-solid-svg-icons';

type DataTypes = {
    collectedGraphData: boolean;
    highValueSelectors: boolean;
    fileIngestHistory: boolean;
    dataQualityHistory: boolean;
};

const useDatabaseManagement = (state: DataTypes, setState: React.Dispatch<React.SetStateAction<DataTypes>>) => {
    const [showSuccessMessage, setShowSuccessMessage] = useState(false);

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
            setShowSuccessMessage(false);
        },
        onSuccess: () => {
            // show UI message that data deletion is happening
            setShowSuccessMessage(true);
            setState({
                collectedGraphData: false,
                dataQualityHistory: false,
                fileIngestHistory: false,
                highValueSelectors: false,
            });
        },
    });

    const handleDelete = () => {
        mutation.mutate({ deleteThisData: state, assetGroupId: tierZeroAssetGroupId });
    };

    return { handleDelete, showSuccessMessage };
};

const DatabaseManagement = () => {
    const theme = useTheme();

    const [state, setState] = useState<DataTypes>({
        collectedGraphData: false,
        highValueSelectors: false,
        fileIngestHistory: false,
        dataQualityHistory: false,
    });

    const [error, setError] = useState(false);
    const [open, setOpen] = useState(false);

    const handleCheckbox = (event: React.ChangeEvent<HTMLInputElement>) => {
        setState({
            ...state,
            [event.target.name]: event.target.checked,
        });
    };

    const handleOpenDialog = () => {
        // if no checkboxes have been checked, display error
        if (Object.values(state).filter(Boolean).length === 0) {
            setError(true);
        } else {
            // clear out any potential error state from previous submission when at least one checkbox has been checked
            setError(false);
            setOpen(true);
        }
    };

    const handleCloseDialog = () => {
        setOpen(false);
    };

    const { handleDelete, showSuccessMessage } = useDatabaseManagement(state, setState);

    const { collectedGraphData, highValueSelectors, fileIngestHistory, dataQualityHistory } = state;

    return (
        <ContentPage title='Clear BloodHound data'>
            <div>
                <Typography variant='body1'>
                    Manage your BloodHound data. Select from the options below which data should be deleted.
                </Typography>
                <Typography variant='body1'>
                    <strong>Caution: </strong> This change is irreversible and will delete data from your environment.
                </Typography>

                <Box display='flex' flexDirection='column' alignItems='start'>
                    <FormControl variant='standard' sx={{ paddingBlock: 2 }} error={error}>
                        {error ? (
                            <Box color={theme.palette.error.main} display='flex' alignItems='center' gap='0.3rem'>
                                <FontAwesomeIcon icon={faCircleXmark} />
                                <FormHelperText sx={{ marginTop: '1.5px', color: theme.palette.error.main }}>
                                    Please make a selection.
                                </FormHelperText>
                            </Box>
                        ) : null}

                        {showSuccessMessage ? (
                            <Box color={theme.palette.info.main} display='flex' alignItems='center' gap='0.3rem'>
                                <FontAwesomeIcon icon={faCircleExclamation} />
                                <FormHelperText sx={{ marginTop: '1.5px', color: theme.palette.info.main }}>
                                    Deletion of the data is under way. Depending on data volume, this may take some time
                                    to complete.
                                </FormHelperText>
                            </Box>
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
                        onClick={handleOpenDialog}>
                        Proceed
                    </Button>
                </Box>
            </div>

            <ConfirmationDialog open={open} handleClose={handleCloseDialog} handleDelete={handleDelete} />
        </ContentPage>
    );
};

export default DatabaseManagement;
