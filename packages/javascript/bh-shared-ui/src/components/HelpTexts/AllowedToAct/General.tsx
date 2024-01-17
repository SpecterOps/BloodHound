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

import { FC } from 'react';
import { groupSpecialFormat } from '../utils';
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} is added to the msds-AllowedToActOnBehalfOfOtherIdentity
                attribute on the computer {targetName}.
            </Typography>

            <Typography variant='body2'>
                An attacker can use this account to execute a modified S4U2self/S4U2proxy abuse chain to impersonate any
                domain user to the target computer system and receive a valid service ticket "as" this user.
            </Typography>

            <Typography variant='body2'>
                One caveat is that impersonated users can not be in the "Protected Users" security group or otherwise
                have delegation privileges revoked. Another caveat is that the principal added to the
                msDS-AllowedToActOnBehalfOfOtherIdentity DACL *must* have a service principal name (SPN) set in order to
                successfully abuse the S4U2self/S4U2proxy process. If an attacker does not currently control an account
                with a SPN set, an attacker can abuse the default domain MachineAccountQuota settings to add a computer
                account that the attacker controls via the Powermad project.
            </Typography>
        </>
    );
};

export default General;
