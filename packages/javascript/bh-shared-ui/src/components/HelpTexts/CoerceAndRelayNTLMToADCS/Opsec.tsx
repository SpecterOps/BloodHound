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

import { Typography } from '@mui/material';
import { FC } from 'react';

const Opsec: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                <b>Detection of NTLM Relay</b>
                NTLM relayed authentications can be detected by login events where the IP address does not match the
                computer’s actual IP address. This detection technique is described in the blog post:{' '}
                <a href={'https://posts.bluraven.io/detecting-ntlm-relay-attacks-d92e99e68fb9'}>
                    Detecting NTLM Relay Attacks
                </a>
                .
            </Typography>

            <Typography variant={'body2'}>
                <b>Detection of Certificate Usage</b>
                Authentication using the obtained certificate is another detection opportunity. If Kerberos
                authentication is used, a domain controller will generate Windows Event ID 4768 ("A Kerberos
                authentication ticket (TGT) was requested"). This event will include the attacker’s IP address rather
                than the target computer’s IP address. Similarly, for Schannel authentication, Event ID 4624 will reveal
                the incorrect IP address. These detection techniques are described in detail under DETECT2 in the
                whitepaper:
                <a href={'https://specterops.io/wp-content/uploads/sites/3/2022/06/Certified_Pre-Owned.pdf'}>
                    Certified Pre-Owned.
                </a>
            </Typography>
        </>
    );
};

export default Opsec;
