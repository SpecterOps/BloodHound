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
import CardWithSwitch from '../CardWithSwitch';
import ConfirmCitrixRDPDialog from './CitrixRDPConfirmDialog';
import { useGetConfiguration } from '../../hooks';
import { useState } from 'react';

// To do: Add this to the shared ui, just using here for ease
const CitrixRDPConfiguration: FC = () => {
    const [isEnabled, setIsEnabled] = useState(false);
    const [isOpenDialog, setIsOpenDialog] = useState(false);
    const configurationData = {
        title: 'Citrix RDP Support',
        description:
            'When enabled, post-processing for the CanRDP edge will look for the presence of the default "Direct Access Users" group and assume that only local Administrators and members of this group can RDP to the system without validation that Citrix VDA is present and correctly configured.Use with caution.',
        enabledDialogText:
            'Analysis has been added with Citrix Configuration, this will ensure that BloodHound can account for Direct Access RDP connections. Compensating controls handled within Citrix are not handled by BloodHound at this time.',
        disabledDialogText:
            'Analysis has been removed with Citrix Configuration, this will result in BloodHound performing analysis to account for this change.',
    };

    // To do: make sure we have correct loading behavior and subsequent behavior for setting existing saved state
    // to do: make sure this is sending and getting the data we need
    const { data, isLoading, isError, isSuccess } = useGetConfiguration();

    console.log(data, isLoading, isError, isSuccess);

    const handleSwitchChange = () => {
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
            <CardWithSwitch
                title={configurationData.title}
                isEnabled={isEnabled}
                description={configurationData.description}
                onSwitchChange={handleSwitchChange}
            />
            <ConfirmCitrixRDPDialog
                open={isOpenDialog}
                dialogDescription={
                    isEnabled ? configurationData.enabledDialogText : configurationData.disabledDialogText
                }
                handleCancel={handleCancel}
                handleConfirm={handleConfirm}
            />
        </>
    );
};

export default CitrixRDPConfiguration;
