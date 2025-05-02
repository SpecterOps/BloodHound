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
                href='https://learn.microsoft.com/en-us/previous-versions/windows/it-pro/windows-server-2003/cc755321(v=ws.10)'>
                Microsoft AD Trust Technical Documentation
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1134/005/'>
                T1134.005: Access Token Manipulation: SID-History Injection
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1187/'>
                T1187: Forced Authentication
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1649/'>
                T1649: Steal or Forge Authentication Certificates
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1558/'>
                T1558: Steal or Forge Kerberos Tickets
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1550/003/'>
                T1550.003: Use Alternate Authentication Material: Pass the Ticket
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://attack.mitre.org/techniques/T1484/001/'>
                T1484.001: Domain or Tenant Policy Modification: Group Policy Modification
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://bloodhound.specterops.io/resources/edges/abuse-tgt-delegation'>
                AbuseTGTDelegation
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://bloodhound.specterops.io/resources/edges/spoof-sid-history'>
                SpoofSIDHistory
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://posts.specterops.io/from-da-to-ea-with-esc5-f9f045aa105c'>
                From DA to EA with ESC5
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://www.pkisolutions.com/escalating-from-child-domains-admins-to-enterprise-admins-in-5-minutes-by-abusing-ad-cs-a-follow-up/'>
                Escalating from child domain’s admins to enterprise admins in 5 minutes by abusing AD CS, a follow up
            </Link>
            <br />
            <Link
                target='_blank'
                rel='noopener'
                href='https://blog.improsec.com/tech-blog/sid-filter-as-security-boundary-between-domains-part-4-bypass-sid-filtering-research'>
                SID filter as security boundary between domains? (Part 4) - Bypass SID filtering research
            </Link>
            <br />
            <Link target='_blank' rel='noopener' href='https://github.com/FSecureLABS/SharpGPOAbuse'>
                SharpGPOAbuse
            </Link>
            <br />
        </Box>
    );
};

export default References;
