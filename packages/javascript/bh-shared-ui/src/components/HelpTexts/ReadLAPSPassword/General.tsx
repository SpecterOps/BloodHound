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
import { groupSpecialFormat } from '../utils';
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} the ability to read the password set by Local Administrator
                Password Solution (LAPS) on the computer {targetName}.
            </Typography>
            <Typography variant='body2'>
                For systems using legacy LAPS, the following AD computer object properties are relevant: -
                **ms-Mcs-AdmPwd**: The plaintext LAPS password. - **ms-Mcs-AdmPwdExpirationTime**: The LAPS password
                expiration time.
            </Typography>

            <Typography variant='body2'>
                For systems using Windows LAPS (2023 edition), the following AD computer object properties are relevant:
                - **msLAPS-Password**: The plaintext LAPS password. - **msLAPS-PasswordExpirationTime**: The LAPS
                password expiration time. - **msLAPS-EncryptedPassword**: The encrypted LAPS password. -
                **msLAPS-EncryptedPasswordHistory**: The encrypted LAPS password history. -
                **msLAPS-EncryptedDSRMPassword**: The encrypted Directory Services Restore Mode (DSRM) password. -
                **msLAPS-EncryptedDSRMPasswordHistory**: The encrypted DSRM password history.
            </Typography>
        </>
    );
};

export default General;
