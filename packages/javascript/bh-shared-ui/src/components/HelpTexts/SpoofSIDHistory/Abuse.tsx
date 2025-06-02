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

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                When spoofing SID history over a same-forest trust, any domain SID from the target domain can be used.
                The most common choice is the Enterprise Admins SID, as it grants full control over the target domain.
            </Typography>

            <Typography variant='body2'>
                Spoofing SID history over a cross-forest trust is more restricted. In this scenario, SID filtering
                removes SIDs with a RID below 1000, meaning built-in AD groups like Domain Admins and Enterprise Admins
                cannot be used. Additionally, group memberships for global and universal groups are not applied based on
                SID history, making accounts in groups like Domain Admins and Enterprise Admins ineffective as targets.
                <p className='my-4'>
                    The attack target must be a user, computer, or a non-builtin group with permissions granted directly
                    or through built-in/domain local groups (NOT through membership of global/universal groups).
                </p>
                Common viable targets with indirect full control over the environment include:
                <ul>
                    <li>The Exchange Windows Permissions group</li>
                    <li>Entra ID sync (MSOL_) accounts</li>
                    <li>Custom groups with administrative control over Tier Zero assets</li>
                </ul>
                <br />
                Alternatively, an attacker can target a domain controller (DC) and use resource-based constrained
                delegation (RBCD) to obtain a local TGT as the DC, which can then be used for a DCSync attack on the
                target domain. However, the RBCD attack requires control over an account (user or computer) in the
                target forest. If no such account is available and the default permissions for creating computers have
                not been restricted, the attacker can first spoof SID history against a target with permissions to
                create computer accounts, to then perform the RBCD attack against a DC.
            </Typography>

            <Typography variant='body2'>
                The spoofed SID can be added to SID history at three different levels for the attacker-controlled user
                of the trusted domain:
                <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                    <li>In the user's SID History AD attribute</li>
                    <li>In the user's Kerberos TGT</li>
                    <li>In the user's Kerberos inter-realm TGT</li>
                </ol>
                <p className='my-4'>
                    The first option enables the attack over both Kerberos and NTLM, whereas the latter two only apply
                    to Kerberos authentication. However, modifying the SID History attribute is risky—it cannot be
                    edited directly via LDAP or built-in AD tools. Mimikatz supports modifying it with the command{' '}
                    <code>sid::patch</code> followed by <code>sid::add</code>, but <code>sid::patch</code> does not work
                    on Windows Server 2016 and later. It is possible to modify the SID History attribute using the
                    DSInternals command <code>Add-ADDBSidHistory</code>, but this requires stopping and restarting the
                    NTDS service, which is not recommended in a production environment.
                </p>
                The second and third options are safer. The following example demonstrates the second option.
            </Typography>

            <Typography variant='body1'>Step 1) Obtain krbtgt Credentials</Typography>
            <Typography variant='body2'>
                The krbtgt credentials can be obtained in multiple ways with administrative access to a DC in the
                trusted domain, such as via a DCSync attack.
            </Typography>

            <Typography variant='body1'>Step 2) Forge and Inject a Golden Ticket</Typography>
            <Typography variant='body2'>
                Generate a Golden Ticket (Kerberos TGT) in the trusted domain with the target's SID added in SID
                history. Alternatively, a Diamond Ticket can be created for better OPSEC.
            </Typography>

            <Typography variant='body2'>On Windows, use Rubeus:</Typography>
            <Typography component={'pre'}>
                {
                    'Rubeus.exe golden /aes256:<krbtgt AES256 secret key> /user:<trusted domain user SAMAccountName> /id:<trusted domain user RID> /domain:<trusted domain DNSname> /sid:<trusted domain SID> /sids:<target SID> /dc:<trusted domain DC DNSname> /nowrap /ptt'
                }
            </Typography>
            <Typography variant='body2'>
                This command injects the ticket into memory, allowing access to the target domain with the permissions
                of the target.
            </Typography>

            <Typography variant='body2'>On Linux, use ticketer.py from Impacket:</Typography>
            <Typography component={'pre'}>
                {
                    'ticketer.py -nthash <krbtgt NT hash> -aesKey <krbtgt AES256 secret key> -domain-sid <trusted domain SID> -domain <trusted domain DNSname> -extra-sid <target SID> <trusted domain user SAMAccountName>'
                }
            </Typography>
            <Typography variant='body2'>
                The ticketer.py command saves the Golden Ticket as a <code>.ccache</code> file. To use it with tools
                supporting Kerberos authentication, set the <code>KRB5CCNAME</code> environment variable:
            </Typography>
            <Typography component={'pre'}>{'export KRB5CCNAME=$path_to_ticket.ccache'}</Typography>
        </>
    );
};

export default Abuse;
