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
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogDescription,
    DialogPortal,
    DialogTitle,
    VisuallyHidden,
} from '@bloodhoundenterprise/doodleui';
import { FC } from 'react';

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
        <Dialog open={open} aria-labelledby='citrix-rdp-alert-dialog-title'>
            <DialogPortal>
                <DialogContent
                    id='citrix-rdp-alert-dialog-description'
                    aria-describedby='citrix-rdp-alert-dialog-description'
                    className='pb-0 text-sm'>
                    <DialogTitle className='text-xl'>{dialogTitle}</DialogTitle>
                    <VisuallyHidden>
                        <DialogDescription>Confrim Citrix Configuration Changes</DialogDescription>
                    </VisuallyHidden>
                    <p className='pb-4 whitespace-break-spaces'>
                        {futureSwitchState ? enabledDialogDescription : disabledDialogDescription}
                    </p>
                    <p>
                        Select <b>Confirm</b> to proceed. Changes will be reflected upon completion of next analysis.
                    </p>
                    <p>
                        Select <b>Cancel</b> to return to previous configuration.
                    </p>
                    <DialogActions>
                        <Button variant={'text'} onClick={onCancel}>
                            Cancel
                        </Button>
                        <Button variant={'text'} fontColor={'primary'} onClick={onConfirm}>
                            Confirm
                        </Button>
                    </DialogActions>
                </DialogContent>
            </DialogPortal>
        </Dialog>
    );
};

export default CitrixRDPConfirmDialog;
