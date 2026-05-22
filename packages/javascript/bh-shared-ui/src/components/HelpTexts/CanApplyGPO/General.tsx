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

const General: FC<EdgeInfoProps> = ({ sourceName, sourceType, targetName, targetType }) => {
    return (
        <>
            <Typography variant='body2'>
                The {sourceType} {sourceName} can apply a GPO to the {targetType} {targetName}.
            </Typography>
            <Typography variant='body2'>
                This edge is created when the principal has write access to the gPLink attribute of the domain or
                organizational unit (OU) containing the account, allowing them to link a GPO that will affect the
                target.
            </Typography>
            <Typography variant='body2'>
                Click the Composition accordion to view where the principal has write access to gPLink.
            </Typography>
        </>
    );
};

export default General;
