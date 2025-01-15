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
import { Link, Box } from '@mui/material';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/tactics/TA0008/'>
                ATT&amp;CK T0008: Lateral Movement
            </Link>
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1021/'>
                ATT&amp;CK T1021: Remote Services
            </Link>
            <Link
                target='_blank'
                rel='noopener'
                href='https://docs.microsoft.com/en-us/azure/active-directory/devices/howto-vm-sign-in-azure-ad-windows'>
                Login to Windows virtual machine in Azure using Azure Active Directory authentication
            </Link>
        </Box>
    );
};

export default References;
