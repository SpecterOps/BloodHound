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
                The {typeFormat(sourceType)} {sourceName} can be used to enroll certificates on behalf of other
                principals for the certificate template {targetName}.
            </Typography>
            <Typography variant='body2'>
                The certificate template {sourceName} is configured to be used as an enrollment agent. The certificate
                template {targetName} is configured to allow enrollment by enrollment agents. Both certificate templates
                are published by an enterprise CA which is trusted for NT authentication and chain up to a root CA for
                the domain. This enables a principal with a certificate of certificate template {sourceName} to enroll
                on behalf of other principals for certificate template {targetName} as long as enrollment agent
                restrictions configured on the enterprise CA permit it.
            </Typography>
        </>
    );
};

export default General;
