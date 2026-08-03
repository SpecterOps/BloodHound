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

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                The role permits Azure Resource Manager operations against the target managed domain. Potential abuse
                includes changing synchronization eligibility, weakening NTLM, Kerberos, LDAP signing, or channel
                binding settings, and changing Secure LDAP exposure or configuration.
            </Typography>

            <Typography variant='body2'>
                Microsoft documents Application Administrator and Groups Administrator Entra roles as additional
                prerequisites for changing managed-domain security settings and synchronization scope. RBAC-only
                behavior has not been validated, so those specific abuse routes depend on the service's additional
                authorization checks.
            </Typography>
        </>
    );
};

export default Abuse;
