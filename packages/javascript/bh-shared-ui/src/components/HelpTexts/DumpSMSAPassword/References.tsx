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
            <Link target='_blank' rel='noopener' href='https://simondotsh.com/infosec/2022/12/12/assessing-smsa.html'>
                https://simondotsh.com/infosec/2022/12/12/assessing-smsa.html
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.ired.team/offensive-security/credential-access-and-credential-dumping/dumping-lsa-secrets'>
                https://www.ired.team/offensive-security/credential-access-and-credential-dumping/dumping-lsa-secrets
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/gentilkiwi/mimikatz'>
                https://github.com/gentilkiwi/mimikatz
            </Link>
        </Box>
    );
};

export default References;
