// Copyright 2026 Specter Ops, Inc.
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

import { Typography } from 'doodle-ui';
import { FC } from 'react';
import Opsec from '../AZEntraDSContributor/Opsec';
import References from '../AZEntraDSContributor/References';
import Composition from './Composition';

const General: FC = () => (
    <Typography variant='body2' component='div'>
        BloodHound creates this post-processed relationship only when one effective principal has:
        <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
            <li>AZContributor or raw AZEntraDSContributor over the target AZEntraDS resource.</li>
            <li>Application Administrator in the target tenant.</li>
            <li>Groups Administrator in the target tenant.</li>
        </ol>
        Post-processing accounts for inherited ARM scope, nested Azure group membership, and effective Entra role
        assignments. Role scope is matched from <code>AZRole.tenantid</code>; a tenant-to-role AZContains relationship
        is not required.
    </Typography>
);

const Abuse: FC = () => (
    <Typography variant='body2' component='div'>
        The source can change the Microsoft Entra Domain Services (Entra DS) managed domain's security configuration and
        broad synchronization boundary when all three authorization components are present.
        <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
            <li>
                Obtain an Azure Resource Manager access token and all three authorization components represented by the
                edge: AZContributor or AZEntraDSContributor, Application Administrator, and Groups Administrator.
            </li>
            <li>
                Read the Entra DS resource and retain its location and current nested settings:
                <Typography component={'pre'} variant='body2'>
                    GET https://management.azure.com/{'{resource-id}'}?api-version=2025-06-01
                </Typography>
            </li>
            <li>
                Choose a supported management-plane change. To change identity synchronization, see ManageEntraDSSync.
            </li>
            <li>
                Submit the change with:
                <Typography component={'pre'} variant='body2'>
                    PUT https://management.azure.com/{'{resource-id}'}?api-version=2025-06-01
                </Typography>
            </li>
        </ol>
    </Typography>
);

const AZManageEntraDS = {
    general: General,
    abuse: Abuse,
    opsec: Opsec,
    references: References,
    composition: Composition,
};

export default AZManageEntraDS;
