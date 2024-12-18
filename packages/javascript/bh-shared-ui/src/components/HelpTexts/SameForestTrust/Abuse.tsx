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
import { Link, Typography } from '@mui/material';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                An attacker with control over any domain within the forest can escalate their privileges to compromise
                other domains through multiple methods.
            </Typography>

            <Typography variant='body1'>Spoof SID history</Typography>
            <Typography variant='body2'>
                An attacker can spoof the SID history of a principal in the target domain, tricking the target domain
                into treating the attacker as a privileged user. See the SpoofSIDHistory edge for more information. 
            </Typography>
            <Typography variant='body2'>
                See the SpoofSIDHistory{' '}
                <Link
                    target='blank'
                    rel='noopener'
                    href='https://support.bloodhoundenterprise.io/hc/en-us/sections/16600927744411-Edges'>
                    edge documentation
                </Link>{' '}for more information.
            </Typography>
            <Typography variant='body2'>
                This attack fails if SID filtering (quarantine) is enabled on the trust relationship in the opposite
                direction of the attack. The SID filtering blocks SIDs belonging to any other domain than the
                attacker-controlled domain. However, enabling this setting is rare and generally not recommended.
            </Typography>

            <Typography variant='body1'>Coerce to TGT</Typography>
            <Typography variant='body2'>
                An attacker can coerce a privileged computer (e.g., a DC) in the target domain to authenticate to an
                attacker-controlled computer configured with unconstrained delegation. This provides the attacker with a
                Kerberos TGT for the coerced computer.
            </Typography>
            <Typography variant='body2'>
                See the CoerceToTGT{' '}
                <Link
                    target='blank'
                    rel='noopener'
                    href='https://support.bloodhoundenterprise.io/hc/en-us/articles/30081024635419-CoerceToTGT'>
                    edge documentation
                </Link>{' '}for more information.
            </Typography>
            <Typography variant='body2'>
                 The attack fails if SID filtering (quarantine) is enabled, as this prevents TGTs from being sent across the trust
                boundary. Again, this setting is rarely configured.
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
                <br />
                1) Obtain a SYSTEM session on a DC in the attacker-controlled domain
                <br />
                2) Create a certificate template allowing ESC1 abuse
                <br />
                3) Publish the certificate template to an enterprise CA
                <br />
                4) Enroll the certificate as a privileged user in the target domain
                <br />
                5) Authenticate as the privileged user in the target domain using the certificate
            </Typography>
            <Typography variant='body2'>
                See this blog post for further details:{' '}
                <Link
                    target='blank'
                    rel='noopener'
                    href='https://posts.specterops.io/from-da-to-ea-with-esc5-f9f045aa105c'>
                    From DA to EA with ESC5
                </Link>.
                <br />
                <br />
                If ADCS is not installed: An attacker can install ADCS and exploit it, as detailed in the blog post:{' '}
                <Link
                    target='blank'
                    rel='noopener'
                    href='https://www.pkisolutions.com/escalating-from-child-domains-admins-to-enterprise-admins-in-5-minutes-by-abusing-ad-cs-a-follow-up/'>
                    Escalating from child domain’s admins to enterprise admins in 5 minutes by abusing AD CS, a follow
                    up
                </Link>.
            </Typography>
            <Typography variant='body1'>GPO linked on Site</Typography>
            <Typography variant='body2'>
                AD sites are stored in the Configuration NC. An attacker with SYSTEM access to a DC can link a malicious
                GPO to the site of a any DC in the forest. Attack steps: Create a malicious GPO in the
                attacker-controlled domain Identify the site name for a target DC Obtain a SYSTEM session on a DC in the
                attacker-controlled domain Link the malicious GPO to the target site Wait for the GPO to apply on the
                target DC.
                <br />
                <br />
                For further details see this blog post:{' '}
                <Link
                    target='blank'
                    rel='noopener'
                    href='https://blog.improsec.com/tech-blog/sid-filter-as-security-boundary-between-domains-part-4-bypass-sid-filtering-research'>
                    SID filter as security boundary between domains? (Part 4) - Bypass SID filtering research
                </Link>.
            </Typography>
        </>
    );
};

export default Abuse;
