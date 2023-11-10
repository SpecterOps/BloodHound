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
                The {typeFormat(sourceType)} {sourceName} is delegated the privilege to enroll certificates of the
                certificate template {targetName} as an enrollment agent.
            </Typography>
            <Typography variant='body2'>
                The certificate template is published to an enterprise CA where the enrollment agent restrictions
                are configured to allow this principal to enroll certificates against this template as an enrollment
                agent. BloodHound does not assess what principals the enrollment agent is allowed to enroll on behalf of.
            </Typography>
        </>
    );
};

export default General;
