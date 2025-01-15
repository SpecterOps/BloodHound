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

const Opsec: FC = () => {
    return (
        <>
            <Typography variant='body2'>
                If the target computer is a workstation and a user is currently logged on, one of two things will
                happen. If the user you are abusing is the same user as the one logged on, you will effectively take
                over their session and kick the logged on user off, resulting in a message to the user. If the users are
                different, you will be prompted to kick the currently logged on user off the system and log on. If the
                target computer is a server, you will be able to initiate the connection without issue provided the user
                you are abusing is not currently logged in.
            </Typography>
            <Typography variant='body2'>
                Remote desktop will create Logon and Logoff events with the access type RemoteInteractive.
            </Typography>
        </>
    );
};

export default Opsec;
