// Copyright 2025 Specter Ops, Inc.
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
            <Link target='_blank' rel='noopener' href='https://wald0.com/?p=179'>
                A Red Teamer's Guide to GPOs and OUs
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/FSecureLABS/SharpGPOAbuse'>
                GitHub: SharpGPOAbuse
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/Hackndo/pyGPOAbuse'>
                GitHub: pyGPOAbuse
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://labs.withsecure.com/publications/ou-having-a-laugh'>
                OU having a laugh?
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.synacktiv.com/publications/ounedpy-exploiting-hidden-organizational-units-acl-attack-vectors-in-active-directory'>
                OUned.py: exploiting hidden Organizational Units ACL attack vectors in Active Directory
            </Link>
        </Box>
    );
};

export default References;
