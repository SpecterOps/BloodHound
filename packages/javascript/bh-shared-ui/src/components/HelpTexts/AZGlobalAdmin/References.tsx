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

import { Box, Link } from '@mui/material';
import { FC } from 'react';

const References: FC = () => {
    return (
        <Box className='overflow-x-auto'>
            <Link target='_blank' rel='noopener noreferrer' href='https://learn.microsoft.com/en-us/entra/identity/role-based-access-control/permissions-reference#global-administrator'>
                Microsoft Entra built-in roles: Global Administrator
            </Link>
            <br />
            <Link target='_blank' rel='noopener noreferrer' href='https://blog.netspi.com/attacking-azure-cloud-shell/'>
                https://blog.netspi.com/attacking-azure-cloud-shell/
            </Link>
        </Box>
    );
};

export default References;
