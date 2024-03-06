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

import { faCircleXmark } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    TextField,
    Typography,
    useTheme,
} from '@mui/material';
import { FC, useState } from 'react';

const confirmationText = 'Please delete my data';

const ConfirmationDialog: FC<{ open: boolean; handleClose: () => void; handleDelete: () => void }> = ({
    open,
    handleClose,
    handleDelete,
}) => {
    const theme = useTheme();

    const [input, setInput] = useState('');
    const [error, setError] = useState(false);

    const handleConfirm = () => {
        if (input.toLowerCase() !== confirmationText.toLowerCase()) {
            setError(true);
        } else {
            // resets local state
            setError(false);
            setInput('');

            // handle events
            handleClose();
            handleDelete();
        }
    };

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const newInput = event.target.value;
        // clear out error state once the user types the phrase
        if (newInput.toLowerCase() === confirmationText.toLowerCase()) {
            setError(false);
        } else {
            setError(true);
        }
        setInput(newInput);
    };

    return open ? (
        <Dialog maxWidth='lg' open>
            <DialogTitle>Confirm deleting data</DialogTitle>
            <DialogContent dividers>
                <Box display='flex' flexDirection='column' gap={2}>
                    <Typography variant='body1'>Continuing will delete all data from your environment.</Typography>
                    <Typography
                        variant='body1'
                        color={theme.palette.error.main}
                        fontWeight={theme.typography.fontWeightMedium}>
                        This change is irreversible.
                    </Typography>
                    <Typography variant='body1'>Please input the phrase prior to click confirm.</Typography>
                    <TextField
                        placeholder={confirmationText}
                        variant='standard'
                        value={input}
                        onChange={handleChange}
                        error={error}
                        helperText={
                            error ? (
                                <>
                                    <FontAwesomeIcon icon={faCircleXmark} /> Please input the phrase prior to clicking
                                    confirm.
                                </>
                            ) : null
                        }></TextField>
                </Box>
            </DialogContent>
            <DialogActions>
                <Button
                    onClick={() => {
                        handleClose();
                        setError(false);
                        setInput('');
                    }}
                    sx={{ color: 'black' }}>
                    Cancel
                </Button>
                <Button onClick={handleConfirm}>Confirm</Button>
            </DialogActions>
        </Dialog>
    ) : null;
};

export default ConfirmationDialog;
