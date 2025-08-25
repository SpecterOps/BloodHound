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

import { Box, Link } from '@mui/material';
import { FC } from 'react';

const References: FC = () => {
    return (
        <Box sx={{ overflowX: 'auto' }}>
            <Link
                target='_blank'
                rel='noopener'
                href='https://specterops.io/wp-content/uploads/sites/3/2022/06/Certified_Pre-Owned.pdf'>
                Certified Pre-Owned
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/GhostPack/Certify/wiki/4-%E2%80%90-Escalation-Techniques#managecertificates'>
                Certify wiki - Escalation Techniques - ManageCertificates
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/ly4k/Certipy/wiki/06-%E2%80%90-Privilege-Escalation#esc7-dangerous-permissions-on-ca'>
                ESC7: Dangerous Permissions on CA
            </Link>
            <br />
        </Box>
    );
};

export default References;
