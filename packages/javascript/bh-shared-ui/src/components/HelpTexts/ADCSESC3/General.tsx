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
                {groupSpecialFormat(sourceType, sourceName)} the capability to perform the ADCS ESC3 attack against the
                domain {targetName}.
            </Typography>
            <Typography variant='body2'>
                The principal has permission to enroll on a certificate allowing them to obtain an enrollment agent
                certificate. They also have permission to enroll for a certificate template that permits enrollment by
                enrollment agents and can be used for authentication. Additionally, they also have enrollment
                permissions for an enterprise CA with the necessary templates published. This enterprise CA is trusted
                for NT authentication in the forest, along with the CA certificate chain up to the root CA certificate.
                This setup lets the principal enroll certificates for any AD forest user or computer, enabling
                authentication and impersonation of any AD forest user or computer without their credentials, unless the
                target user or computer is protected by enrollment agent restrictions on the enterprise CA.
            </Typography>
        </>
    );
};

export default General;
