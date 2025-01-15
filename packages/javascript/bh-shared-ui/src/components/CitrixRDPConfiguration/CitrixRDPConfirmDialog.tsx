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
import { Typography, Dialog, DialogActions, DialogContent, DialogTitle } from '@mui/material';
import { Button } from '@bloodhoundenterprise/doodleui';

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
    return (
        <Dialog
            open={open}
            maxWidth='sm'
            aria-labelledby='citrix-rdp-alert-dialog-title'
            aria-describedby='citrix-rdp-alert-dialog-description'>
            <DialogTitle id='citrix-rdp-alert-dialog-title' sx={{ fontSize: '1.25rem' }}>
                {dialogTitle}
            </DialogTitle>
            <DialogContent id='citrix-rdp-alert-dialog-description' sx={{ paddingBottom: 0 }}>
                <Typography sx={{ paddingBottom: '1rem', whiteSpace: 'break-spaces', fontSize: '0.75rem' }}>
                    {futureSwitchState ? enabledDialogDescription : disabledDialogDescription}
                </Typography>
                <Typography sx={{ fontSize: '0.75rem' }}>
                    Select <b>`Confirm`</b> to proceed. Changes will be reflected upon completion of next analysis.
                </Typography>
                <Typography sx={{ fontSize: '0.75rem' }}>
                    Select <b>`Cancel`</b> to return to previous configuration.
                </Typography>
            </DialogContent>
            <DialogActions>
                <Button variant={'text'} onClick={onCancel}>
                    Cancel
                </Button>
                <Button variant={'text'} fontColor={'primary'} onClick={onConfirm}>
                    Confirm
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default CitrixRDPConfirmDialog;
