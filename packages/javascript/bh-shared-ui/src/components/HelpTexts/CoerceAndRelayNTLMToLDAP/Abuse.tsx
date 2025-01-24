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

import {Typography} from '@mui/material';
import {FC} from 'react';

const Abuse: FC = () => {
    return (
        <Typography variant='body2'>
            An attacker who is a member of "Authenticated Users" triggers a webclient based coercion from the target
            computer to their attacker host. Since the connection originates from the webclient instead of SMB, the
            attacker can relay the inbound auth attempt to LDAP(S) if the LDAP(S) signing/EPA settings allow it. The
            relay to LDAP(S) is used to abuse RBCD or Shadow Credentials against the victim computer account.
        </Typography>
    );
};

export default Abuse;
