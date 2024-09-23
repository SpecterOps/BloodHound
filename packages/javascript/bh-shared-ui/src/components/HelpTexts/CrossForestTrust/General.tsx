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
import { EdgeInfoProps } from '../index';

const General: FC<EdgeInfoProps> = ({ sourceName, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                The CrossForestTrust edge represents a cross-forest (inter-realm) trust relationship between two
                domains/forests. In this relationship, the {sourceName} domain has a cross-forest trust to the{' '}
                {targetName} domain, allowing principals (users and computers) from {targetName} to access resources in{' '}
                {sourceName}.
            </Typography>

            <Typography variant='body1'>Trust Edge Properties</Typography>
            <Typography variant='body2'>
                BloodHound stores the following properties for CrossForestTrust edges (listed under Relationship
                Information):
            </Typography>
            <Typography variant='body2'>
                <b>Spoof SID History Blocked</b>
                <br />
                An attacker with control over the trusted domain may attempt to escalate privileges in the trusting
                domain using a <i>spoof SID history</i> attack by injecting the SID of a privileged principal from the
                trusting domain into their authentication request. However, this attack can be prevented by SID
                filtering.
                <br />
                <br />
                SID filtering removes domain SIDs from authentication requests of foreign principals based on the trust
                configuration. Built-in SIDs (S-1-5-32-*) are always filtered, regardless of the configuration. For
                cross-forest trusts, blocking spoof SID history attacks requires filtering out SIDs that do not belong
                to the forest of the trusted domain.
                <br />
                <br />
                Cross-forest trusts blocks spoof SID history attacks by default. Forest trusts filter out SIDs that do
                not belong to the forest of the trusted domain. However, SID filtering for forest trusts can be relaxed
                by enabling the <code>trustAttributes</code> flag <code>TREAT_AS_EXTERNAL</code>. When enabled, SID
                filtering behaves like it does for external trusts (without quarantine mode), meaning only SIDs with a
                RID below 1000 (built-in AD principals) are filtered, allowing a potential spoof SID history attack.
                <br />
                <br />
                External trusts have <i>quarantine mode</i> (<code>trustAttributes</code> flag{' '}
                <code>QUARANTINED_DOMAIN</code>) enabled by default, protecting the trusting domain by filtering out
                SIDs that do not belong to the trusted domain in authentication requests. An external trust with
                quarantine mode disabled does not block spoof SID history attacks.
                <br />
                <br />
                SID filtering is managed from the outbound side of the trust. As a result, this property is only created
                if trust data from the outbound side has been ingested.
            </Typography>
            <Typography variant='body2'>
                <b>TGT Delegation</b>
                <br />
                TGT delegation determines whether unconstrained delegation is allowed over the trust. When a principal
                from the trusted domain authenticates against a Kerberos resource with unconstrained delegation in the
                trusting domain, their Kerberos TGT (Ticket Granting Ticket) is forwarded to the resource as part of
                Kerberos authentication, but only if TGT delegation is enabled (true).
                <br />
                <br />
                By default, TGT delegation is disabled for cross-forest trusts. It is enabled if the{' '}
                <code>trustAttributes</code> flag <code>CROSS_ORGANIZATION_ENABLE_TGT_DELEGATION</code> is enabled for
                the trust and <i>quarantine mode</i> (<code>trustAttributes</code> flag <code>QUARANTINED_DOMAIN</code>)
                is NOT enabled.
                <br />
                <br />
                TGT delegation is controlled from the inbound side of the trust, and the property is therefore only
                created if trust data from this side has been ingested.
            </Typography>
            <Typography variant='body2'>
                <b>Transitive</b>
                <br />
                Transitivity defines whether the trust extends beyond the two domains involved. A transitive trust
                allows access not only to principals of the trusted domain but also to those from other domains trusted
                by the trusted domain.
                <br />
                <br />
                Forest trusts are always transitive, external trusts are non-transitive.
                <br />
                <br />
                Attackers can bypass the limitations of non-transitive trusts by manually requesting local Kerberos TGTs
                for each domain in the trust chain. They can then use these local TGTs to access Kerberos resources that
                would otherwise be denied if requested directly. For more details, refer to "External trusts are evil"
                under References.
            </Typography>
            <Typography variant='body2'>
                <b>Trust Attributes</b>
                <br />
                This property stores the raw value of the <code>trustAttributes</code> LDAP attribute, which defines the
                trust's configuration settings. BloodHound retains this property from both the outbound and inbound
                sides of the trust, as they may differ.
            </Typography>
            <Typography variant='body2'>
                <b>Trust Type</b>
                <br />
                The trust type for cross-forest trusts can be Forest, External, Kerberos (Realm), or Unknown. Refer to
                the "Microsoft AD Trust Technical Documentation" under References for more details.
            </Typography>
        </>
    );
};

export default General;
