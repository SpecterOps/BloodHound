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
import { Typography } from '@mui/material';
import { EdgeInfoProps } from '../index';

const WindowsAbuse: FC<EdgeInfoProps> = ({ sourceName, sourceType }) => {
    return (
        <>
            <Typography variant='body2'>To abuse this permission, use Whisker. </Typography>

            <Typography variant='body2'>
                You may need to authenticate to the Domain Controller as{' '}
                {sourceType === 'User' || sourceType === 'Computer'
                    ? `${sourceName} if you are not running a process as that user/computer`
                    : `a member of ${sourceName} if you are not running a process as a member`}
            </Typography>

            <Typography component={'pre'}>{'Whisker.exe add /target:<TargetPrincipal>'}</Typography>

            <Typography variant='body2'>For other optional parameters, view the Whisker documentation.</Typography>
        </>
    );
};

export default WindowsAbuse;
