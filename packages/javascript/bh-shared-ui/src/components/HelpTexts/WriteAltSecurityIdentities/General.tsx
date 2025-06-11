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

import { Typography } from '@mui/material';
import { FC } from 'react';
import { EdgeInfoProps } from '../index';
import { groupSpecialFormat, typeFormat } from '../utils';

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName, targetType }) => {
    return (
        <>
            <Typography variant='body2'>
                {groupSpecialFormat(sourceType, sourceName)} the ability to write to the AltSecurityIdentities attribute
                of the {typeFormat(targetType)} {targetName}.
            </Typography>

            <Typography variant='body2'>
                The altSecurityIdentities attribute allows you to specify <i>explicit certificate mappings</i> for a
                principal. An explicit certificate mapping is a reference to a certificate, which can be used to
                authenticate as the principal.
            </Typography>
        </>
    );
};

export default General;
