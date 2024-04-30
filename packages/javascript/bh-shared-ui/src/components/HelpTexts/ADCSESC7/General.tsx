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

import { FC } from 'react';
import { useHelpTextStyles, groupSpecialFormat } from '../utils';
import { EdgeInfoProps } from '../index';
import { Typography } from '@mui/material';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType }) => {
    const classes = useHelpTextStyles();
    return (
        <>
            <Typography variant='body2' className={classes.containsCodeEl}>
                {groupSpecialFormat(sourceType, sourceName)} has the privileges to perform the ADCS ESC4 attack against
                the target domain.
                <br />
                <br />
                The principal has permissions to modify the settings on one or more certificate templates, enabling the
                principal configure the certificate templates for ADCS ESC1 conditions, which allows them to specify an
                alternate subject name and use the certificate for authentication. They also has enrollment permission
                for an enterprise CA with the necessary templates published. This enterprise CA is trusted for NT
                authentication and chains up to a root CA for the forest. This setup lets the principal modify the
                certificate templates to allow enrollment as any targeted AD forest user or computer without knowing
                their credentials, and impersonation of those targets by certificate authentication.
            </Typography>
        </>
    );
};

export default General;
