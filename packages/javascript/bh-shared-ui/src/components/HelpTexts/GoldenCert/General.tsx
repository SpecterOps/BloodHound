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
import { typeFormat } from '../utils';
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName }) => {
    return (
        <>
            <Typography variant='body2'>
                The {typeFormat(sourceType)} {sourceName} has a certificate private key that can be abused to sign
                "golden" certificates for authentication of any enabled principal in the AD forest of domain{' '}
                {targetName}.
            </Typography>
            <Typography variant='body2'>
                The {typeFormat(sourceType)} {sourceName} hosts the enrollment service of an enterprise CA which implies
                it has the private key of the enterprise CA's certificate. This private key allows an attacker to sign
                certificates for authentication as any enabled principal in the AD forest of domain {targetName}, as the
                enterprise CA is trusted for NT authentication and chain up to a root CA.
            </Typography>
            <Typography variant='body2'>
                It may not be possible to obtain the certificate private key if it is protected with a Trusted Platform
                Module (TPM) or using a Hardware Security Module (HSM). However, it may still be possible to compromise
                the AD forest. Administrative access to the enterprise CA host lets an attacker publish certificate
                templates, approve denied enrollment requests, and more. The {typeFormat(sourceType)} {sourceName} will
                have an ESC7 edge to the domain {targetName} if any such attack has been found possible by BloodHound.
            </Typography>
        </>
    );
};

export default General;
