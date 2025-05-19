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
import { ConfigurationKey, parseRestrictOutboundNTLMDefaultValueConfiguration } from 'js-client-library';
import { FC, useState } from 'react';
import { useGetConfiguration, useUpdateConfiguration } from '../../hooks';
import { useNotifications } from '../../providers';
import CardWithSwitch from '../CardWithSwitch';
import ConfirmCitrixRDPDialog from './RestrictOutboundNTLMDefaultValueConfirmDialog';

export const configurationData = {
    title: 'Default Restrict Outbound NTLM Property Value',
    description:
        ''
};

const RestrictOutboundNTLMDefaultValueConfiguration: FC = () => {
    const [isOpenDialog, setIsOpenDialog] = useState(false);

    const { addNotification } = useNotifications();
    const { data, isFetching } = useGetConfiguration();
    const updateConfiguration = useUpdateConfiguration();

    const restrictOutboundNTLMDefaultValueConfigurationEnabled = parseRestrictOutboundNTLMDefaultValueConfiguration(data)?.value.enabled;

    // optimistically update switch state during pending mutations
    const haveUnsettledRequests = updateConfiguration.isLoading || isFetching;

    const computeSwitchState = (): boolean => {
        if (haveUnsettledRequests && updateConfiguration.variables?.key === ConfigurationKey.RestrictOutboundNTLMDefaultValue) {
            return updateConfiguration.variables?.value.enabled;
        } else {
            return !!restrictOutboundNTLMDefaultValueConfigurationEnabled;
        }
    };

    const switchState = computeSwitchState();

    const toggleShowDialog = () => {
        setIsOpenDialog((prev) => !prev);
    };

    const handleConfirm = () => {
        toggleShowDialog();
        updateConfiguration.mutateAsync(
            {
                key: ConfigurationKey.RestrictOutboundNTLMDefaultValue,
                value: { enabled: !switchState },
            },
            {
                onError: () => {
                    addNotification('There was an error updating configuration.');
                },
            }
        );
    };

    return (
        <>
            <CardWithSwitch
                title={configurationData.title}
                isEnabled={switchState}
                description={configurationData.description}
                disableSwitch={haveUnsettledRequests}
                onSwitchChange={toggleShowDialog}
            />
            <ConfirmCitrixRDPDialog
                open={isOpenDialog}
                futureSwitchState={!switchState}
                onCancel={toggleShowDialog}
                onConfirm={handleConfirm}
            />
        </>
    );
};

export default RestrictOutboundNTLMDefaultValueConfiguration;
