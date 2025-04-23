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
import { EdgeInfoProps } from '../index';

const General: FC<EdgeInfoProps> = ({ sourceName, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                The SameForestTrust edge represents a same-forest (intra-realm) trust relationship between two domains
                within the same AD forest. In this relationship, the {sourceName} domain has a same-forest trust with
                the {targetName} domain, allowing principals (users and computers) from {targetName} to access resources
                in {sourceName}.
            </Typography>
            <Typography variant='body2'>
                Since both domains belong to the same forest, they inherently trust each other, granting implicit access
                to resources across domains. It also means compromising one of the domains enable compromise of the
                other.
            </Typography>
            <Typography variant='body1'>Trust Edge Properties</Typography>
            <Typography variant='body2'>
                BloodHound stores the following properties for SameForestTrust edges (listed under Relationship
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
                configuration. Built-in SIDs (S-1-5-32-*) are always filtered, regardless of the configuration. SID
                filtering blocks the spoof SID history attack over same-forest trusts when the trust is configured in{' '}
                <i>quarantine mode</i> (<code>trustAttributes</code> flag <code>QUARANTINED_DOMAIN</code> enabled),
                which filters out SIDs that do not belong to the trusted domain. Same-forest trusts does not have
                quarantine mode enabled by default.
                <br />
                <br />
                SID filtering is managed from the outbound side of the trust, and the "Spoof SID History Blocked"
                property is therefore only created if trust data from this side has been ingested.
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
                By default, TGT delegation is enabled for same-forest trusts. It is disabled if <i>quarantine mode</i> (
                <code>trustAttributes</code> flag <code>QUARANTINED_DOMAIN</code>) is enabled.
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
                Same-forest trusts are always transitive.
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
                The trust type for same-forest trusts can be TreeRoot, ParentChild, or CrossLink (Shortcut). Refer to
                the "Microsoft AD Trust Technical Documentation" under References for more details.
            </Typography>
        </>
    );
};

export default General;
