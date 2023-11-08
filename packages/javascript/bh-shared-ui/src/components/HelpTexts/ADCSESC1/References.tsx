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
                href='https://specterops.io/wp-content/uploads/sites/3/2022/06/Certified_Pre-Owned.pdf'>
                https://specterops.io/wp-content/uploads/sites/3/2022/06/Certified_Pre-Owned.pdf
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://book.hacktricks.xyz/windows-hardening/active-directory-methodology/ad-certificates/domain-escalation#misconfigured-certificate-templates-esc1'>
                https://book.hacktricks.xyz/windows-hardening/active-directory-methodology/ad-certificates/domain-escalation#misconfigured-certificate-templates-esc1
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://hideandsec.sh/books/cheatsheets-82c/page/active-directory-certificate-services'>
                https://hideandsec.sh/books/cheatsheets-82c/page/active-directory-certificate-services
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/ly4k/Certipy'>
                https://github.com/ly4k/Certipy
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/GhostPack/Rubeus'>
                https://github.com/GhostPack/Rubeus
            </Link>
        </Box>
    );
};

export default References;
