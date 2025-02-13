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

import { Button } from '@bloodhoundenterprise/doodleui';
import { faCheckCircle } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';
import {
    Alert,
    AlertTitle,
    Box,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Paper,
    Skeleton,
    Typography,
} from '@mui/material';
import {
    Flag,
    PageWithTitle,
    Permission,
    useFeatureFlags,
    useMountEffect,
    useNotifications,
    usePermissions,
    useToggleFeatureFlag,
} from 'bh-shared-ui';
import { FC, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { setDarkMode } from 'src/ducks/global/actions';
import { useAppDispatch } from 'src/store';

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
            <Box sx={{ p: 2, display: 'flex', justifyContent: 'space-between', gap: '1rem' }}>
                <Box overflow='hidden'>
                    <Typography variant='h6'>{flag.name}</Typography>
                    <Typography variant='body1'>{flag.description}</Typography>
                </Box>
                <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                    <Button disabled={disabled} onClick={handleOnClick} className='w-32'>
                        <Box display={'flex'} alignItems={'center'}>
                            {flag.enabled ? (
                                <FontAwesomeIcon style={{ marginRight: '8px' }} icon={faCheckCircle} fixedWidth />
                            ) : null}
                            <Typography>{flag.enabled ? 'Enabled' : 'Disabled'}</Typography>
                        </Box>
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
                    variant='tertiary'
                    onClick={onCancel}
                    data-testid='early-access-features-warning-dialog_button-close'>
                    {'Take me back'}
                </Button>
                <Button
                    variant='primary'
                    onClick={onConfirm}
                    data-testid='early-access-features-warning-dialog_button-confirm'>
                    {'I understand, show me the new stuff!'}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

const EarlyAccessFeatures: FC = () => {
    const [showWarningDialog, setShowWarningDialog] = useState(true);
    const dispatch = useAppDispatch();
    const navigate = useNavigate();
    const { data, isLoading, isError } = useFeatureFlags();
    const toggleFeatureFlag = useToggleFeatureFlag();

    const { checkPermission } = usePermissions();
    const hasPermission = checkPermission(Permission.APP_WRITE_APPLICATION_CONFIGURATION);

    const { addNotification, dismissNotification } = useNotifications();
    const notificationKey = 'manage-feature-flags-permission';

    const effect: React.EffectCallback = (): void => {
        if (!hasPermission) {
            addNotification(
                `Your role does not grant permission to manage feature flags. Please contact your administrator for details.`,
                notificationKey,
                {
                    persist: true,
                    anchorOrigin: { vertical: 'top', horizontal: 'right' },
                }
            );
        }

        return dismissNotification(notificationKey);
    };

    useMountEffect(effect);

    return (
        <>
            <PageWithTitle
                title='Early Access Features'
                data-testid='early-access-features'
                pageDescription={
                    <Typography variant='body2' paragraph>
                        Enable or disable features available under early access. These features may be unstable, broken,
                        or incomplete, but are available for testing.
                    </Typography>
                }>
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
                                            // TODO: Consider adding more flexibility/composability to side effects for toggling feature flags on and off
                                            if (flag.key === 'dark_mode') {
                                                dispatch(setDarkMode(false));
                                            }
                                            toggleFeatureFlag.mutate(flagId);
                                        }}
                                        disabled={showWarningDialog || !hasPermission}
                                    />
                                </Box>
                            ))
                    ))}
            </PageWithTitle>
            <EarlyAccessFeaturesWarningDialog
                open={showWarningDialog}
                onCancel={() => navigate(-1)}
                onConfirm={() => setShowWarningDialog(false)}
            />
        </>
    );
};

export default EarlyAccessFeatures;
