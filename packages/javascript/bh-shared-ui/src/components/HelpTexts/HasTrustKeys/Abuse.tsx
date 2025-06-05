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

import { Link, Typography } from '@mui/material';
import { FC } from 'react';
import CodeController from '../CodeController/CodeController';

const Abuse: FC = () => {
    return (
        <>
            <Typography variant='body1'>Step 1: Obtain Trust Keys</Typography>
            <Typography variant='body2'>
                Trust keys can be dumped with administrative access to a domain controller of the source domain.
            </Typography>
            <Typography variant='body2'>On Windows, use Mimikatz to dump the trust keys:</Typography>
            <CodeController>{`lsadump::trust /patch`}</CodeController>
            <Typography variant='body2'>
                The trust keys for the target trust account appear under "[ Out ]" for the target domain.
            </Typography>

            <Typography variant='body1'>Step 2: Authenticate as the Trust Account</Typography>
            <Typography variant='body2'>
                The RC4 version of the trust keys serves as the RC4 Kerberos secret key for the trust account. This can
                be used directly to request a Kerberos Ticket-Granting Ticket (TGT).
            </Typography>
            <Typography variant='body2'>
                The AES trust keys are not identical to the AES Kerberos secret keys of the trust account due to
                different salt values. However, you can derive the AES Kerberos secret keys using the cleartext trust
                key and tools like krbrelayx.py. (See reference:{' '}
                <Link
                    target='_blank'
                    rel='noopener'
                    href='https://snovvcrash.rocks/2021/05/21/calculating-kerberos-keys.html'>
                    A Note on Calculating Kerberos Keys for AD Accounts
                </Link>{' '}
                ).
            </Typography>
            <Typography variant='body2'>
                When authenticating as a trust account, there are two key limitations:
                <ol style={{ listStyleType: 'decimal', paddingLeft: '1.5em' }}>
                    <li>Only Kerberos authentication is supported (NTLM authentication is not possible)</li>
                    <li>
                        Only network logins work (interactive logins such as RUNAS, console login, and RDP are not
                        possible)
                    </li>
                </ol>
            </Typography>
            <Typography variant='body2'>On Windows, use Rubeus to obtain a TGT:</Typography>
            <CodeController>
                {`Rubeus.exe asktgt /user:<trust account SAMAccountName> /domain:<target domain DNS name> /rc4:<RC4 trust key> /nowrap /ptt`}
            </CodeController>
            <Typography variant='body2'>On Linux, use Impacket's getTGT.py to obtain a TGT:</Typography>
            <CodeController>
                {`python getTGT.py <target domain DNS name>/<trust account SAMAccountName> -hashes : <RC4 trust key>`}
            </CodeController>
        </>
    );
};

export default Abuse;
