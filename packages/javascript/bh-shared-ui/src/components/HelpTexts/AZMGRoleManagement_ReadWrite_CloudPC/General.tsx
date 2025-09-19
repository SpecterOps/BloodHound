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


import { Typography } from '@mui/material';
import { FC } from 'react';

const General: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                Information for <code>RoleManagement.ReadWrite.CloudPC</code>.
            </Typography>
            <ul><li><strong>Application:</strong> Allows the app to read and manage the Cloud PC role-based access control (RBAC) settings, without a signed-in user. This includes reading and managing Cloud PC role definitions and memberships.</li></ul>
        </>
    );
};

export default General;