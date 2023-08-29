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
            <Link target='_blank' rel='noopener' href='https://github.com/GhostPack/Rubeus#s4u'>
                https://github.com/GhostPack/Rubeus#s4u
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://labs.mwrinfosecurity.com/blog/trust-years-to-earn-seconds-to-break/'>
                https://labs.mwrinfosecurity.com/blog/trust-years-to-earn-seconds-to-break/
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://blog.harmj0y.net/activedirectory/s4u2pwnage/'>
                https://blog.harmj0y.net/activedirectory/s4u2pwnage/
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://twitter.com/gentilkiwi/status/806643377278173185'>
                https://twitter.com/gentilkiwi/status/806643377278173185
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.coresecurity.com/blog/kerberos-delegation-spns-and-more'>
                https://www.coresecurity.com/blog/kerberos-delegation-spns-and-more
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://blog.harmj0y.net/redteaming/from-kekeo-to-rubeus/'>
                https://blog.harmj0y.net/redteaming/from-kekeo-to-rubeus/
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://blog.harmj0y.net/redteaming/another-word-on-delegation/'>
                https://blog.harmj0y.net/redteaming/another-word-on-delegation/
            </Link>
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.thehacker.recipes/ad/movement/kerberos/delegations/constrained'>
                https://www.thehacker.recipes/ad/movement/kerberos/delegations/constrained
            </Link>
        </Box>
    );
};

export default References;
