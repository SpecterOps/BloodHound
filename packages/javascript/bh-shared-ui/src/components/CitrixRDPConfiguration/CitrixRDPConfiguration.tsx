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
//import { useGetConfiguration, useUpdateConfiguration } from '../../hooks';
import { useState } from 'react';

export const configurationData = {
    title: 'Citrix RDP Support',
    description:
        'When enabled, post-processing for the CanRDP edge will look for the presence of the default "Direct Access Users" group and assume that only local Administrators and members of this group can RDP to the system without validation that Citrix VDA is present and correctly configured. Use with caution.',
};

// To do: Add this to the shared ui, just using here for ease
const CitrixRDPConfiguration: FC = () => {
    const [isEnabled, setIsEnabled] = useState(false);
    const [isOpenDialog, setIsOpenDialog] = useState(false);

    // To do: make sure we have correct loading behavior and subsequent behavior for setting existing saved state
    // To do: make sure this is sending and getting the data we need
    // const { data, isLoading, isError, isSuccess } = useGetConfiguration();
    // const updateConfigurationMutation = useUpdateConfiguration();

    // console.log(data, isLoading, isError, isSuccess);

    const toggleShowDialog = () => {
        setIsOpenDialog((prev) => !prev);
    };

    const handleSwitchChange = () => {
        setIsEnabled((prev) => !prev);
        toggleShowDialog();
    };

    const handleConfirm = () => {
        // To do: add correct call args
        //updateConfigurationMutation.mutate({ key: 'test' } as any);
        toggleShowDialog();
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
                isEnabled={isEnabled}
                handleCancel={handleSwitchChange}
                handleConfirm={handleConfirm}
            />
        </>
    );
};

export default CitrixRDPConfiguration;
