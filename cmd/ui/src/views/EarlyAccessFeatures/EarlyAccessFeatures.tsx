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

import { faCheckCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Alert,
    AlertTitle,
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Paper,
    Skeleton,
    Typography,
} from '@mui/material';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Flag, useFeatureFlags, useToggleFeatureFlag } from 'src/hooks/useFeatureFlags';
import { ContentPage } from 'bh-shared-ui';

export const EarlyAccessFeatureToggle: React.FC<{
    flag: Flag;
    onClick: (flagId: number) => void;
    disabled: boolean;
}> = ({ flag, onClick, disabled }) => {
    const handleOnClick = () => {
        onClick(flag.id);
    };

    return (
        <Paper>
            <Box p={2} display='flex' justifyContent='space-between' flexWrap='wrap' style={{ rowGap: '1rem' }}>
                <Box overflow='hidden'>
                    <Typography variant='h6'>{flag.name}</Typography>
                    <Typography variant='body1'>{flag.description}</Typography>
                </Box>
                <Box>
                    <Button
                        disabled={disabled}
                        variant='outlined'
                        color={flag.enabled ? 'primary' : 'inherit'}
                        sx={{
                            borderColor: () => {
                                if (!flag.enabled) return 'rgba(0,0,0,0.23)';
                            },
                        }}
                        onClick={handleOnClick}
                        startIcon={flag.enabled ? <FontAwesomeIcon icon={faCheckCircle} fixedWidth /> : null}>
                        {flag.enabled ? 'Enabled' : 'Disabled'}
                    </Button>
                </Box>
            </Box>
        </Paper>
    );
};

export const EarlyAccessFeaturesWarningDialog: React.FC<{
    open: boolean;
    onCancel: () => void;
    onConfirm: () => void;
}> = ({ open, onCancel, onConfirm }) => {
    return (
        <Dialog
            open={open}
            disableEscapeKeyDown
            PaperProps={{
                //@ts-ignore
                'data-testid': 'early-access-features-warning-dialog',
            }}>
            <DialogTitle>Heads up!</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    The features on this page are under active development and may be unstable, broken, or incomplete.
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button
                    color='inherit'
                    onClick={onCancel}
                    data-testid='early-access-features-warning-dialog_button-close'>
                    {'Take me back'}
                </Button>
                <Button
                    color='primary'
                    onClick={onConfirm}
                    data-testid='early-access-features-warning-dialog_button-confirm'>
                    {'I understand, show me the new stuff!'}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

const EarlyAccessFeatures: React.FC = () => {
    const navigate = useNavigate();
    const { data, isLoading, isError } = useFeatureFlags();
    const toggleFeatureFlag = useToggleFeatureFlag();
    const [showWarningDialog, setShowWarningDialog] = useState(true);

    return (
        <>
            <ContentPage title='Early Access Features' data-testid='early-access-features'>
                {!showWarningDialog &&
                    (isLoading ? (
                        <Paper elevation={0}>
                            <Box p={2}>
                                <Typography variant='h6'>
                                    <Skeleton />
                                </Typography>
                                <Typography variant='body1'>
                                    <Skeleton />
                                </Typography>
                            </Box>
                        </Paper>
                    ) : isError ? (
                        <Alert severity='error'>
                            <AlertTitle>Could Not Display Early Access Features</AlertTitle>
                            An unexpected error occurred. Please refresh this page or try again later.
                        </Alert>
                    ) : data!.filter((flag) => flag.user_updatable).length === 0 ? (
                        <Paper elevation={0}>
                            <Box p={2}>
                                <Typography variant='h6'>No Early Access Features Available</Typography>
                                <Typography variant='body1'>
                                    There are no early access features available at this time. Please check back later.
                                </Typography>
                            </Box>
                        </Paper>
                    ) : (
                        data!
                            .filter((flag) => flag.user_updatable)
                            .sort((a, b) => a.id - b.id)
                            .map((flag, index) => (
                                <Box
                                    mt={index === 0 ? 0 : 2}
                                    key={flag.id}
                                    data-testid={`early-access-features_toggle-${index}`}>
                                    <EarlyAccessFeatureToggle
                                        flag={flag}
                                        onClick={(flagId) => {
                                            toggleFeatureFlag.mutate(flagId);
                                        }}
                                        disabled={showWarningDialog}
                                    />
                                </Box>
                            ))
                    ))}
            </ContentPage>
            <EarlyAccessFeaturesWarningDialog
                open={showWarningDialog}
                onCancel={() => navigate(-1)}
                onConfirm={() => setShowWarningDialog(false)}
            />
        </>
    );
};

export default EarlyAccessFeatures;
