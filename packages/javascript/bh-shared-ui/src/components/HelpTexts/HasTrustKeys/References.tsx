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
            <Link
                target='_blank'
                rel='noopener'
                href='https://blog.improsec.com/tech-blog/sid-filter-as-security-boundary-between-domains-part-7-trust-account-attack-from-trusting-to-trusted'>
                SID filter as security boundary between domains? (Part 7) - Trust account attack
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/gentilkiwi/mimikatz'>
                Mimikatz GitHub
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/dirkjanm/krbrelayx'>
                krbrelayx GitHub
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://snovvcrash.rocks/2021/05/21/calculating-kerberos-keys.html'>
                A Note on Calculating Kerberos Keys for AD Accounts
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/GhostPack/Rubeus'>
                Rubeus GitHub
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://github.com/fortra/impacket/blob/master/examples/getTGT.py'>
                Impacket getTGT.py
            </Link>
        </Box>
    );
};

export default References;
