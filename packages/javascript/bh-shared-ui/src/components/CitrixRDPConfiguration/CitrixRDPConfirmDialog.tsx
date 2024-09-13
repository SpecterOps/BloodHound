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
import { FC } from 'react';
import { Typography, Button, useTheme, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';

type CitrixRDPConfirmDialogProps = {
    open: boolean;
    futureSwitchState: boolean;
    onCancel: () => void;
    onConfirm: () => void;
};
export const dialogTitle = 'Confirm environment configuration';
const enabledDialogDescription =
    'Analysis has been added with Citrix Configuration, this will ensure that BloodHound can account for Direct Access RDP connections. \n\nCompensating controls handled within Citrix are not handled by BloodHound at this time.';
const disabledDialogDescription =
    'Analysis has been removed with Citrix Configuration, this will result in BloodHound performing analysis to account for this change';

const CitrixRDPConfirmDialog: FC<CitrixRDPConfirmDialogProps> = ({ open, futureSwitchState, onCancel, onConfirm }) => {
    const theme = useTheme();

    return (
        <Dialog
            open={open}
            maxWidth='md'
            aria-labelledby='citrix-rdp-alert-dialog-title'
            aria-describedby='citrix-rdp-alert-dialog-description'>
            <DialogTitle id='citrix-rdp-alert-dialog-title' sx={{ fontSize: '20px' }}>
                {dialogTitle}
            </DialogTitle>
            <DialogContent sx={{ paddingBottom: 0 }}>
                <Typography variant='body2' sx={{ paddingBottom: '16px', whiteSpace: 'break-spaces' }}>
                    {futureSwitchState ? enabledDialogDescription : disabledDialogDescription}
                </Typography>
                <Typography variant='body2'>
                    Select <b>`Confirm`</b> to proceed. Changes will be reflected upon completion of next analysis.
                </Typography>
                <Typography variant='body2'>
                    Select <b>`Cancel`</b> to return to previous configuration.
                </Typography>
            </DialogContent>
            <DialogActions>
                <Button sx={{ color: theme.palette.button.secondary }} onClick={onCancel}>
                    Cancel
                </Button>
                <Button sx={{ color: theme.palette.button.primary }} onClick={onConfirm}>
                    Confirm
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default CitrixRDPConfirmDialog;
