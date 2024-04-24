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
            <Link
                target='_blank'
                rel='noopener'
                href='https://posts.specterops.io/adcs-esc13-abuse-technique-fda4272fbd53'>
                ADCS ESC13 Abuse Technique
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2008-R2-and-2008/dd378897(v=ws.10)?redirectedfrom=MSDN'>
                Authentication Mechanism Assurance for AD DS in Windows Server 2008 R2 Step-by-Step Guide
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.gradenegger.eu/en/using-authentication-mechanism-assurance-ama-to-secure-the-login-of-administrative-accounts/'>
                Use Authentication Mechanism Assurance (AMA) to secure administrative account logins
            </Link>
        </Box>
    );
};

export default References;
