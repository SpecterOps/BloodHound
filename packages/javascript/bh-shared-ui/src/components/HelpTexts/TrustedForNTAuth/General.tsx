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

const General: FC<EdgeInfoProps> = ({ sourceName }) => {
    return (
        <>
            <Typography variant='body2'>
                The NTAuthStore contains the certificate of the Enterprise CA, {sourceName}. The consequence of the
                relationship is that certificate issued by the Enterprise CA are trusted for authentication in the AD
                forest of the NTAuthStore.
            </Typography>
        </>
    );
};

export default General;
