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

import { Box, Typography } from '@mui/material';
import { PageWithTitle, CardWithToggle, useGetConfiguration } from 'bh-shared-ui';
import { FC, useState } from 'react';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import { Button } from '@bloodhoundenterprise/doodleui';
import { faTriangleExclamation } from '@fortawesome/free-solid-svg-icons';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome';

// To do: Add this to the shared ui, just using here for ease
const CitrixRDPConfiguration: FC = () => {
    const [isEnabled, setIsEnabled] = useState(false);
    const [isOpenDialog, setIsOpenDialog] = useState(false);
    const configurationData = {
        title: 'Citrix RDP Support',
        description:
            'When enabled, post-processing for the CanRDP edge will look for the presence of the default &quot;Direct Access Users&quot; group and assume that only local Administrators and members of this group can RDP to the system without validation that Citrix VDA is present and correctly configured.Use with caution.',
    };

    // To do: make sure we have correct loading behavior and subsequent behavior for setting existing saved state
    // to do: make sure this is sending and getting the data we need
    const { data, isLoading, isError, isSuccess } = useGetConfiguration();

    console.log(data, isLoading, isError, isSuccess);

    const handleToggleChange = () => {
        setIsEnabled((prev) => !prev);
        toggleShowDialog();
    };

    const toggleShowDialog = () => {
        setIsOpenDialog((prev) => !prev);
    };

    const handleCancel = () => {
        toggleShowDialog();
        setIsEnabled((prev) => !prev);
    };

    const handleConfirm = async () => {
        //const { data, isLoading, isError, isSuccess } = useUpdateConfiguration();
    };

    return (
        <>
            <CardWithToggle
                title={configurationData.title}
                isEnabled={isEnabled}
                description={configurationData.description}
                onToggleChange={handleToggleChange}
            />
            <ConfirmCitrixRDPDialog
                open={isOpenDialog}
                handleCancel={handleCancel}
                handleConfirm={handleConfirm}
                isLoading={isLoading}
            />
        </>
    );
};

// To do: Abstract to shared ui, make a citrixRDPFolder there
const ConfirmCitrixRDPDialog: FC<{
    open: boolean;
    handleCancel: () => void;
    handleConfirm: () => void;
    isLoading: boolean;
}> = ({ open, handleCancel, handleConfirm, isLoading }) => {
    return (
        <Dialog
            open={open}
            maxWidth='md'
            aria-labelledby='citrix-rdp-alert-dialog-title'
            aria-describedby='citrix-rdp-alert-dialog-description'>
            <DialogTitle id='citrix-rdp-alert-dialog-title'>Confirm environment configuration</DialogTitle>
            <DialogContent sx={{ paddingBottom: 0 }}>
                <Box sx={{ display: 'flex', alignItems: 'center', paddingBottom: '16px' }}>
                    <FontAwesomeIcon icon={faTriangleExclamation} size='2x' />
                    <Typography sx={{ marginLeft: '20px' }}>
                        Analysis has been added with Citrix Configuration, this will ensure that BloodHound can account
                        for Direct Access RDP connections. Compensating controls handled within Citrix are not handled
                        by BloodHound at this time.
                    </Typography>
                </Box>
                <Typography>
                    Select <b>`Confirm`</b> to proceed and to start analysis.
                </Typography>
                <Typography>
                    Select <b>`Cancel`</b> to return to previous configuration.
                </Typography>
            </DialogContent>
            <DialogActions>
                <Button variant='tertiary' onClick={() => handleCancel()} disabled={isLoading}>
                    Cancel
                </Button>
                <Button onClick={() => handleConfirm()} disabled={isLoading}>
                    Confirm
                </Button>
            </DialogActions>
        </Dialog>
    );
};

const BloodHoundConfiguration = () => {
    return (
        <PageWithTitle
            title='Bloodhound Configuration'
            pageDescription={
                // To do: Add max-width to parent content
                <Typography variant='body2' paragraph sx={{ maxWidth: '955px' }}>
                    Brief Description of the Feature Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do
                    eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud
                    exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
                </Typography>
            }>
            <Box sx={{ display: 'flex', flexDirection: 'column', gap: '24px', maxWidth: '860px' }}>
                <CitrixRDPConfiguration />
            </Box>
        </PageWithTitle>
    );
};

export default BloodHoundConfiguration;
