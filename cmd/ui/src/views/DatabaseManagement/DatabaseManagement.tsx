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
} from '@mui/material';
import { ContentPage } from 'bh-shared-ui';
import { useState } from 'react';
import ConfirmationDialog from './ConfirmationDialog';

const DatabaseManagement = () => {
    const [checked, setChecked] = useState({
        collectedGraphData: false,
        highValueSelectors: false,
        fileIngestHistory: false,
        dataQualityHistory: false,
    });

    const [error, setError] = useState(false);
    const [open, setOpen] = useState(false);

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setChecked({
            ...checked,
            [event.target.name]: event.target.checked,
        });
    };

    const handleProceed = () => {
        // if nothing is checked, display error
        if (Object.values(checked).filter(Boolean).length === 0) {
            setError(true);
        } else {
            // clear out error on succesful submission
            setError(false);
            setOpen(true);
        }
    };

    const handleClose = () => {
        setOpen(false);
    };

    const { collectedGraphData, highValueSelectors, fileIngestHistory, dataQualityHistory } = checked;

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
                        {error ? <FormHelperText>Please make a selection</FormHelperText> : null}
                        <FormGroup>
                            <FormControlLabel
                                label='Collected graph data (all nodes and edges)'
                                control={
                                    <Checkbox
                                        checked={collectedGraphData}
                                        onChange={handleChange}
                                        name='collectedGraphData'
                                    />
                                }
                            />
                            <FormControlLabel
                                label='Custom High Value selectors'
                                control={
                                    <Checkbox
                                        checked={highValueSelectors}
                                        onChange={handleChange}
                                        name='highValueSelectors'
                                    />
                                }
                            />
                            <FormControlLabel
                                label='File Ingest Log history'
                                control={
                                    <Checkbox
                                        checked={fileIngestHistory}
                                        onChange={handleChange}
                                        name='fileIngestHistory'
                                    />
                                }
                            />
                            <FormControlLabel
                                label='Data Quality history'
                                control={
                                    <Checkbox
                                        checked={dataQualityHistory}
                                        onChange={handleChange}
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
                        onClick={handleProceed}>
                        Proceed
                    </Button>
                </Box>
            </div>

            <ConfirmationDialog open={open} handleClose={handleClose} />
        </ContentPage>
    );
};

export default DatabaseManagement;
